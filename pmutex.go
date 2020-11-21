package mfs

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const pmLocked int32 = -1

// PMutex - Read Write Try Mutex with change priority (Promote and Reduce)
// F methods (like LockF and TryLockF) Locks mutex if mutex already locked then this methods will be first in lock queue
// Promote - lock mutex from RLock to Lock
// Reduce - lock mutex from Lock to RLock
type PMutex struct {
	state int32
	pr    int32
	prom  int32
	mx    sync.Mutex
	ch    chan struct{}
}

func (m *PMutex) chGet() chan struct{} {
	m.mx.Lock()
	if m.ch == nil {
		m.ch = make(chan struct{}, 1)
	}
	r := m.ch
	m.mx.Unlock()
	return r
}

func (m *PMutex) chClose() {
	// it's need only when exists parallel
	// to make faster need add counter to add drop listners of chan

	if m.ch == nil {
		return // it neet to test!!!! theoreticly works when channel get operation is befor atomic operations
	}

	var o chan struct{}
	m.mx.Lock()
	if m.ch != nil {
		o = m.ch
		m.ch = nil
	}
	m.mx.Unlock()
	if o != nil {
		close(o)
	}
}

// Lock - locks mutex
func (m *PMutex) Lock() {

	if m.pr > 0 {
		m.lockS(false)
		return
	}

	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {

		return
	}

	// Slow way
	m.lockS(false)
}

// LockF - locks mutex first (out of turn)
func (m *PMutex) LockF() {
	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {

		return
	}

	// Slow way
	m.lockS(true)
}

// Unlock - unlocks mutex
func (m *PMutex) Unlock() {
	if atomic.CompareAndSwapInt32(&m.state, pmLocked, 0) {
		m.chClose()
		return
	}

	panic("PMutex: Unlock fail")
}

// Reduce - lock mutex from Lock to RLock
func (m *PMutex) Reduce() {
	if atomic.CompareAndSwapInt32(&m.state, pmLocked, 1) {
		m.chClose()
		return
	}

	panic("PMutex: Reduce fail")
}

// TryLock - try locks mutex with context
func (m *PMutex) TryLock(ctx context.Context) bool {

	if m.pr > 0 {
		return m.lockST(ctx, false)
	}

	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
		return true
	}

	// Slow way
	return m.lockST(ctx, false)
}

// TryLockF - try locks mutex with context first (out of turn)
func (m *PMutex) TryLockF(ctx context.Context) bool {
	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
		return true
	}

	// Slow way
	return m.lockST(ctx, true)
}

// LockD - try locks mutex with time duration
func (m *PMutex) LockD(d time.Duration) bool {
	if m.pr > 0 {
		return m.lockSD(d, false)
	}

	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
		return true
	}

	// Slow way
	return m.lockSD(d, false)
}

// LockDF - try locks mutex with time duration first (out of turn)
func (m *PMutex) LockDF(d time.Duration) bool {
	if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
		return true
	}

	// Slow way
	return m.lockSD(d, true)
}

// RLock - read locks mutex
func (m *PMutex) RLock() {
	if m.pr > 0 {
		m.rlockS(false)
		return
	}

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return
	}

	// Slow way
	m.rlockS(false)
}

// RLockF - read locks mutex first (out of turn)
func (m *PMutex) RLockF() {

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return
	}

	// Slow way
	m.rlockS(true)
}

// RUnlock - unlocks mutex
func (m *PMutex) RUnlock() {
	i := atomic.AddInt32(&m.state, -1)
	if i > 0 {
		m.chClose()
		return
	} else if i == 0 {
		m.chClose()
		return
	}

	panic("PMutex: RUnlock fail")
}

// RTryLock - try read locks mutex with context
func (m *PMutex) RTryLock(ctx context.Context) bool {
	if m.pr > 0 {
		return m.rlockST(ctx, false)
	}

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return true
	}

	// Slow way
	return m.rlockST(ctx, false)
}

// RTryLockF - try read locks mutex with context first (out of turn)
func (m *PMutex) RTryLockF(ctx context.Context) bool {

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return true
	}

	// Slow way
	return m.rlockST(ctx, true)
}

// RLockD - try read locks mutex with time duration
func (m *PMutex) RLockD(d time.Duration) bool {
	if m.pr > 0 {
		return m.rlockSD(d, false)
	}

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return true
	}

	// Slow way
	return m.rlockSD(d, false)
}

// RLockDF - try read locks mutex with time duration first (out of turn)
func (m *PMutex) RLockDF(d time.Duration) bool {

	k := atomic.LoadInt32(&m.state)
	if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
		return true
	}

	// Slow way
	return m.rlockSD(d, true)
}

func (m *PMutex) lockS(f bool) {

	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}
				return
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		}
	}

}

func (m *PMutex) lockST(ctx context.Context, f bool) bool {
	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}

				return true
			}
		}

		if ctx == nil {
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}
			return false
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-ctx.Done():
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}
			return false
		}

	}
}

func (m *PMutex) lockSD(d time.Duration, f bool) bool {
	// may be use context.WithTimeout(context.Background(), d) however NO it's not fun
	t := time.After(d)

	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			if atomic.CompareAndSwapInt32(&m.state, 0, pmLocked) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}
				return true
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-t:
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}
			return false
		}

	}
}

func (m *PMutex) rlockS(f bool) {

	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	var k int32
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}

				return
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		}

	}

}

func (m *PMutex) rlockST(ctx context.Context, f bool) bool {

	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	var k int32
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}

				return true
			}
		}

		if ctx == nil {
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}

			return false
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-ctx.Done():
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}

			return false
		}

	}

}

func (m *PMutex) rlockSD(d time.Duration, f bool) bool {

	t := time.After(d)

	if f {
		m.mx.Lock()
		m.pr++
		m.mx.Unlock()
	}

	ch := m.chGet()
	var k int32
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k >= 0 && atomic.CompareAndSwapInt32(&m.state, k, k+1) {
				if f {
					m.mx.Lock()
					m.pr--
					m.mx.Unlock()
				}

				return true
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-t:
			if f {
				m.mx.Lock()
				m.pr--
				m.mx.Unlock()
			}

			return false
		}

	}

}

// Promote - lock mutex from RLock to Lock
func (m *PMutex) Promote() {

	if m.pr > 0 {
		m.promoteS(false)
		return
	}

	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {

		return
	}

	// Slow way
	m.promoteS(false)
}

// PromoteF - lock mutex from RLock to Lock first (out of turn)
func (m *PMutex) PromoteF() {

	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {

		return
	}

	// Slow way
	m.promoteS(true)
}

// TryPromote - try locks mutex from RLock to Lock with context
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) TryPromote(ctx context.Context) bool {

	if m.pr > 0 {
		return m.promoteST(ctx, false)
	}

	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {
		return true
	}

	// Slow way
	return m.promoteST(ctx, false)
}

// TryPromoteF - try locks mutex from RLock to Lock with context first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) TryPromoteF(ctx context.Context) bool {
	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {
		return true
	}

	// Slow way
	return m.promoteST(ctx, true)
}

// PromoteD - try locks mutex from RLock to Lock with time duration
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) PromoteD(d time.Duration) bool {
	if m.pr > 0 {
		return m.promoteSD(d, false)
	}

	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {
		return true
	}

	// Slow way
	return m.promoteSD(d, false)
}

// PromoteDF - try locks mutex from RLock to Lock with time duration first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) PromoteDF(d time.Duration) bool {
	if atomic.CompareAndSwapInt32(&m.state, 1, pmLocked) {
		return true
	}

	// Slow way
	return m.promoteSD(d, true)
}

func (m *PMutex) promoteS(f bool) {

	if f {
		m.mx.Lock()
		m.pr++
		m.prom++
		m.mx.Unlock()
	} else {
		m.mx.Lock()
		m.prom++
		m.mx.Unlock()
	}

	var k int32

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k == m.prom {
				m.mx.Lock()
				if k == m.prom {
					if atomic.CompareAndSwapInt32(&m.state, k, pmLocked) {
						if f {
							m.pr--
							m.prom--
						} else {
							m.prom--
						}

						m.mx.Unlock()
						return
					}
				}
				m.mx.Unlock()
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		}
	}

}

func (m *PMutex) promoteST(ctx context.Context, f bool) bool {

	if f {
		m.mx.Lock()
		m.pr++
		m.prom++
		m.mx.Unlock()
	} else {
		m.mx.Lock()
		m.prom++
		m.mx.Unlock()
	}

	var k int32

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k == m.prom {
				m.mx.Lock()
				if k == m.prom {
					if atomic.CompareAndSwapInt32(&m.state, k, pmLocked) {
						if f {
							m.pr--
							m.prom--
						} else {
							m.prom--
						}

						m.mx.Unlock()
						return true
					}
				}
				m.mx.Unlock()
			}
		}

		if ctx == nil {
			if f {
				m.mx.Lock()
				m.pr--
				m.prom--
				m.mx.Unlock()
			} else {
				m.mx.Lock()
				m.prom--
				m.mx.Unlock()
			}
			m.RUnlock()
			return false
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-ctx.Done():
			if f {
				m.mx.Lock()
				m.pr--
				m.prom--
				m.mx.Unlock()
			} else {
				m.mx.Lock()
				m.prom--
				m.mx.Unlock()
			}
			m.RUnlock()
			return false
		}

	}

}

func (m *PMutex) promoteSD(d time.Duration, f bool) bool {

	t := time.After(d)

	if f {
		m.mx.Lock()
		m.pr++
		m.prom++
		m.mx.Unlock()
	} else {
		m.mx.Lock()
		m.prom++
		m.mx.Unlock()
	}

	var k int32

	ch := m.chGet()
	for {

		if f || m.pr == 0 {
			k = atomic.LoadInt32(&m.state)
			if k == m.prom {
				m.mx.Lock()
				if k == m.prom {
					if atomic.CompareAndSwapInt32(&m.state, k, pmLocked) {
						if f {
							m.pr--
							m.prom--
						} else {
							m.prom--
						}

						m.mx.Unlock()
						return true
					}
				}
				m.mx.Unlock()
			}
		}

		select {
		case <-ch:
			ch = m.chGet()
		case <-t:
			if f {
				m.mx.Lock()
				m.pr--
				m.prom--
				m.mx.Unlock()
			} else {
				m.mx.Lock()
				m.prom--
				m.mx.Unlock()
			}
			m.RUnlock()
			return false
		}

	}

}
