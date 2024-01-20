package lock

import (
	"errors"
	"sync"
)

var (
	ErrReadLock  = errors.New("can't acquire lock because the resource is being modified")
	ErrWriteLock = errors.New("can't acquire exclusive lock because the resource is in use")
	ErrCapacity  = errors.New("resource lock capacity limit reached")
)

const Capacity = 1024

type Locker struct {
	mx    sync.Mutex
	locks map[string]*entry
}

func NewLocker() *Locker {
	return &Locker{
		locks: map[string]*entry{},
	}
}

type entry struct {
	writing bool
	refs    int
}

func (e *entry) newUnlocker(l *Locker, key string) (unlock func()) {
	return func() {
		l.mx.Lock()
		defer l.mx.Unlock()
		e.refs--
		if e.refs <= 0 {
			delete(l.locks, key)
		}
		l = nil
	}
}

func (l *Locker) WriteLock(key string) (unlock func(), err error) {
	l.mx.Lock()
	defer l.mx.Unlock()
	e := l.locks[key]
	if e != nil {
		// existing lock is always an error
		return nil, ErrWriteLock
	}
	if len(l.locks) >= Capacity {
		return nil, ErrCapacity
	}
	e = &entry{writing: true, refs: 1}
	l.locks[key] = e
	return e.newUnlocker(l, key), nil
}

func (l *Locker) ReadLock(key string) (unlock func(), err error) {
	l.mx.Lock()
	defer l.mx.Unlock()
	e := l.locks[key]
	if e != nil {
		if e.writing {
			return nil, ErrReadLock
		}
		e.refs++
		return e.newUnlocker(l, key), nil
	}
	if len(l.locks) >= Capacity {
		return nil, ErrCapacity
	}
	e = &entry{refs: 1}
	l.locks[key] = e
	return e.newUnlocker(l, key), nil
}
