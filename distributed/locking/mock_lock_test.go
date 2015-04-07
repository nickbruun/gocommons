package locking

import (
	log "github.com/nickbruun/gocommons/logging"
	logrus "github.com/Sirupsen/logrus"
	"time"
	"testing"
	"sync"
)

func TestMockLock(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	provider := NewMockLockProvider()

	// * Test acquisition and release of lock.
	{
		l := provider.GetLock("/lock/path")

		for i := 0; i < 3; i++ {
			l.Lock()
			l.Unlock()
		}

		for i := 0; i < 3; i++ {
			l.LockTimeout(time.Second)
			l.Unlock()
		}
	}

	// * Test that locks that have timed out are skipped when acquring locks.
	{
		l1 := provider.GetLock("/lock/path")
		l2 := provider.GetLock("/lock/path")
		l3 := provider.GetLock("/lock/path")
		l4 := provider.GetLock("/lock/path")

		for i := 0; i < 3; i++ {
			var l2TimedOut, l3Acquired, l4Acquired sync.WaitGroup
			l2TimedOut.Add(1)
			l3Acquired.Add(1)
			l4Acquired.Add(1)

			l1.Lock()

			go func() {
				if _, err := l2.LockTimeout(10 * time.Millisecond); err != ErrTimeout {
					t.Fatalf("Expected lock to time out, but error was: %v", err)
				}

				l2TimedOut.Done()
			}()

			go func() {
				l3.Lock()
				l3Acquired.Done()
			}()

			go func() {
				l4.Lock()
				l4Acquired.Done()
			}()

			l2TimedOut.Wait()

			l1.Unlock()

			l3Acquired.Wait()

			l3.Unlock()

			l4Acquired.Wait()

			l4.Unlock()
		}
	}

	// * Test failing the provider.
	{
		l1 := provider.GetLock("/lock/path")
		l2 := provider.GetLock("/lock/path")

		var l1Waiting, l1Failed, l2Acquired sync.WaitGroup
		l1Waiting.Add(1)
		l1Failed.Add(1)
		l2Acquired.Add(1)

		failCh, err := l1.Lock()
		if err != nil {
			t.Fatalf("Unexpected error acquiring lock: %v", err)
		}

		go func() {
			l1Waiting.Done()
			<-failCh
			l1Failed.Done()
		}()

		go func() {
			l2.Lock()
			l2Acquired.Done()
		}()

		l1Waiting.Wait()

		provider.Fail()

		l1Failed.Wait()

		if err = l1.Unlock(); err != ErrNotLocked {
			t.Fatalf("Expected failed lock to return ErrNotLocked: %v", err)
		}

		provider.Recover()

		l2Acquired.Wait()

		l2.Unlock()
	}

	{
		l1 := provider.GetLock("/lock/path")
		l2 := provider.GetLock("/lock/path")

		var l1Waiting, l1Failed, l2Waiting, l2Acquired sync.WaitGroup
		l1Waiting.Add(1)
		l1Failed.Add(1)
		l2Waiting.Add(1)
		l2Acquired.Add(1)

		failCh, err := l1.Lock()
		if err != nil {
			t.Fatalf("Unexpected error acquiring lock: %v", err)
		}

		go func() {
			l1Waiting.Done()
			<-failCh
			l1Failed.Done()
		}()

		l1Waiting.Wait()
		log.Debug("Lock 1 waiting to fail")

		provider.Fail()
		log.Debug("Provider failed")

		l1Failed.Wait()
		log.Debug("Lock 1 failed")

		if err = l1.Unlock(); err != ErrNotLocked {
			t.Fatalf("Expected failed lock to return ErrNotLocked: %v", err)
		}

		go func() {
			l2Waiting.Done()
			l2.Lock()
			l2Acquired.Done()
		}()

		l2Waiting.Wait()
		log.Debug("Lock 2 waiting to acquire")

		provider.Recover()
		log.Debug("Provider recovered")

		l2Acquired.Wait()
		log.Debug("Lock 2 acquired")

		l2.Unlock()
	}
}
