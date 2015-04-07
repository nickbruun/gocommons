package zkutils

import (
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"time"
)

// Connection session watcher.
//
// Watches for possible losses of sessions.
type connSessionWatcher struct {
	ec             <-chan zk.Event
	sessionTimeout time.Duration
	recvTimeout    time.Duration
	watchers       []chan bool
	lock           sync.Mutex
}

// New connection session watcher.
func newConnSessionWatcher(ec <-chan zk.Event, sessionTimeout time.Duration, recvTimeout time.Duration) *connSessionWatcher {
	w := &connSessionWatcher{
		ec:             ec,
		sessionTimeout: sessionTimeout,
		recvTimeout:    recvTimeout,
		watchers:       make([]chan bool, 0),
	}

	go w.watch()

	return w
}

// Publish the possible loss of a session.
func (w *connSessionWatcher) publish(expired bool) {
	w.lock.Lock()
	defer w.lock.Unlock()

	for _, ch := range w.watchers {
		ch <- expired
		close(ch)
	}
	w.watchers = make([]chan bool, 0)
}

// Watch a session from the point where it was acquired until it is somehow
// lost.
//
// Returns true if the event channel has been closed.
func (w *connSessionWatcher) watchSession() (done bool) {
	// Wait for the either the session to actively expire or the connection to
	// disconnect.
	for {
		ev, ok := <-w.ec
		if !ok {
			return true
		}

		if ev.Type != zk.EventSession {
			continue
		}

		switch ev.State {
		case zk.StateExpired:
			w.publish(true)
			return false

		case zk.StateDisconnected:
			// Loop until we either hit the possible expiration deadline or the
			// session is reacquired.
			deadlineChan := time.After(w.sessionTimeout - w.recvTimeout)
			resolved := false

			for !resolved {
				select {
				case ev, ok = <-w.ec:
					if !ok {
						return true
					}

					if ev.Type != zk.EventSession {
						break
					}

					switch ev.State {
					case zk.StateHasSession:
						resolved = true

					case zk.StateExpired:
						w.publish(true)
						return false
					}

				case <-deadlineChan:
					w.publish(false)
					return false
				}
			}
		}
	}
}

// Watch.
func (w *connSessionWatcher) watch() {
	// Wait for us to get a session.
	for {
		ev, ok := <-w.ec
		if !ok {
			break
		}

		if ev.Type != zk.EventSession || ev.State != zk.StateHasSession {
			continue
		}

		// Watch the newly established session.
		if w.watchSession() {
			break
		}
	}

	// Close all outstanding watchers by indicating a hard session loss, as the
	// end of the event channel means the connection has been closed.
	w.publish(true)
}

// Add watcher.
//
// Returns a one-shot channel which will emit a boolean value indicating
// whether or not the session loss is due to explicit expiration.
func (w *connSessionWatcher) AddWatcher() <-chan bool {
	ch := make(chan bool, 1)

	w.lock.Lock()
	defer w.lock.Unlock()

	w.watchers = append(w.watchers, ch)

	return ch
}
