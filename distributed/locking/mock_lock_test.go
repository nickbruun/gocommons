package locking

import (
	logrus "github.com/Sirupsen/logrus"
	log "github.com/nickbruun/gocommons/logging"
	"sync"
	"testing"
	"time"
)

func TestMockLock(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	provider := NewMockLockProvider()

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
