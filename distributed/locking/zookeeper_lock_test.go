package locking

import (
	logrus "github.com/Sirupsen/logrus"
	log "github.com/nickbruun/gocommons/logging"
	"github.com/nickbruun/gocommons/zkutils"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"testing"
	"time"
)

func TestZooKeeperLock(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	// * Set up the test cluster.
	testCluster, connMan := zkutils.CreateTestClusterAndConnMan(t, 1)
	defer testCluster.Stop()
	defer connMan.Close()

	provider := NewZooKeeperLockProvider(connMan, zk.WorldACL(zk.PermAll))

	// * Test simple locking and unlocking.
	{
		l := provider.GetLock("/my/lock")

		for i := 0; i < 3; i++ {
			AssertLock(t, l)
			log.Debug("Acquired lock")
			AssertUnlock(t, l)
			log.Debug("Released lock")
			AssertLockBeforeTimeout(t, l, time.Second)
			log.Debug("Acquired lock before 1 s timeout")
			AssertUnlock(t, l)
			log.Debug("Released lock")
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

			AssertLock(t, l1)
			log.Debug("Acquired lock 1, waiting for lock 2 to time out")

			go func() {
				AssertLockTimedOut(t, l2, 10*time.Millisecond)
				log.Debug("Lock 2 timed out")
				l2TimedOut.Done()
			}()

			go func() {
				AssertLock(t, l3)
				log.Debug("Lock 3 acquired")
				l3Acquired.Done()
			}()

			go func() {
				AssertLock(t, l4)
				log.Debug("Lock 4 acquired")
				l4Acquired.Done()
			}()

			l2TimedOut.Wait()

			AssertUnlock(t, l1)
			log.Debug("Released lock 1, waiting for lock 3 to be acquired")

			l3Acquired.Wait()

			AssertUnlock(t, l3)
			log.Debug("Released lock 3, waiting for lock 4 to be acquired")

			l4Acquired.Wait()

			AssertUnlock(t, l4)
			log.Debug("Released lock 4")
		}
	}

	// * Test lock failure by removing lock node.
	{
		l := provider.GetLock("/lock/path")
		failCh := AssertLock(t, l)
		log.Debug("Acquried lock, removing node")

		children, _, err := connMan.Conn.Children("/lock/path")
		if err != nil {
			t.Fatalf("Failed to list children of /lock/path: %v", err)
		}

		for _, c := range children {
			connMan.Conn.Delete("/lock/path/"+c, -1)
		}

		select {
		case <-failCh:
		case <-time.After(time.Second):
			t.Fatalf("Lock did not fail after 1 s")
		}

		if err := l.Unlock(); err != ErrNotLocked {
			t.Fatalf("Expected attempt at unlocking to return ErrNotLocked, but returned: %v", err)
		}
	}
}
