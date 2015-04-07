package locking

import (
	logrus "github.com/Sirupsen/logrus"
	"github.com/nickbruun/gocommons/zkutils"
	"github.com/samuel/go-zookeeper/zk"
	"testing"
	"time"
)

func TestZooKeeperLock(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	// * Set up the test cluster.
	testCluster, connMan := zkutils.CreateTestClusterAndConnMan(t, 1)
	defer testCluster.Stop()
	defer connMan.Close()

	// * Test simple locking and unlocking.
	provider := NewZooKeeperLockProvider(connMan, zk.WorldACL(zk.PermAll))

	{
		l := provider.GetLock("/my/lock")

		for i := 0; i < 3; i++ {
			if _, err := l.Lock(); err != nil {
				t.Fatalf("Failed to acquire lock: %v", err)
			}

			t.Logf("Acquired lock")

			if err := l.Unlock(); err != nil {
				t.Fatalf("Failed to release lock: %v", err)
			}

			t.Logf("Released lock")

			if _, err := l.LockTimeout(time.Second); err != nil {
				t.Fatalf("Failed to acquire lock with 1 s timeout: %v", err)
			}

			t.Logf("Acquired lock before 1 s timeout")

			if err := l.Unlock(); err != nil {
				t.Fatalf("Failed to release lock: %v", err)
			}

			t.Logf("Released lock")
		}
	}
}
