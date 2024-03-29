package lock_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/server/internal/lock"
)

func TestLocker(t *testing.T) {

	// test capacity
	t.Run("capacity", func(t *testing.T) {
		locker := lock.NewLocker()
		for i := 0; i < lock.Capacity; i++ {
			_, err := locker.ReadLock(fmt.Sprintf("tmp-%d", i))
			be.NilErr(t, err)
		}
		_, err := locker.ReadLock("tmp-err")
		be.True(t, errors.Is(err, lock.ErrCapacity))
		_, err = locker.WriteLock("tmp-err")
		be.True(t, errors.Is(err, lock.ErrCapacity))
	})

	t.Run("allow multiple read locks", func(t *testing.T) {
		locker := lock.NewLocker()
		_, err := locker.ReadLock("tmp-01")
		be.NilErr(t, err)
		_, err = locker.ReadLock("tmp-01")
		be.NilErr(t, err)
	})

	t.Run("error: existing write lock", func(t *testing.T) {
		id := "tmp-01"
		locker := lock.NewLocker()
		unlock, err := locker.WriteLock(id)
		be.NilErr(t, err)
		_, err = locker.ReadLock(id)
		be.True(t, errors.Is(err, lock.ErrReadLock))
		unlock()
		// try again
		_, err = locker.ReadLock(id)
		be.NilErr(t, err)
	})

	t.Run("error: existing read lock", func(t *testing.T) {
		id := "tmp-01"
		locker := lock.NewLocker()
		unlock, err := locker.ReadLock(id)
		be.NilErr(t, err)
		_, err = locker.WriteLock(id)
		be.True(t, errors.Is(err, lock.ErrWriteLock))
		unlock()
		// try again
		_, err = locker.WriteLock(id)
		be.NilErr(t, err)
	})

}
