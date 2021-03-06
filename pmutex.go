package mfs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PMutex - Read Write Try Mutex with change priority (Promote and Reduce)
// F methods (like LockF and TryLockF) Locks mutex if mutex already locked then this methods will be first in lock queue
// Promote - lock mutex from RLock to Lock
// Reduce - lock mutex from Lock to RLock
type PMutex struct {
	state     int32
	pr        int32
	prom      int32
	promState int32
	mx        sync.Mutex
	ch        chan struct{}
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

// chClose - unlocks other routines needs mx.Lock
func (m *PMutex) chClose() {
	// it's need only when exists parallel
	// to make faster need add counter to add drop listners of chan
	if m.ch == nil {
		return
	}
	var o chan struct{}

	if m.ch != nil {
		o = m.ch
		m.ch = nil
	}
	if o != nil {
		close(o)
	}
}

// Lock - locks mutex
func (m *PMutex) Lock() {

	m.mx.Lock()

	if m.pr > 0 {
		m.mx.Unlock()
		m.lockS(false)
		return
	}

	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()
	// Slow way
	m.lockS(false)
}

// LockF - locks mutex first (out of turn)
func (m *PMutex) LockF() {

	m.mx.Lock()

	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()

	// Slow way
	m.lockS(true)
}

// Unlock - unlocks mutex
func (m *PMutex) Unlock() {

	m.mx.Lock()

	if m.state == -1 {
		m.state = m.promState
		m.promState = 0
		m.chClose()
	} else {
		panic(fmt.Sprintf("PMutex: Unlock fail (%v)", m.state))
	}
	m.mx.Unlock()
}

// Reduce - lock mutex from Lock to RLock
func (m *PMutex) Reduce() {

	m.mx.Lock()

	if m.state == -1 {
		m.state = m.promState + 1
		m.promState = 0
		m.chClose()
	} else {
		panic(fmt.Sprintf("PMutex: Reduce fail (%v)", m.state))
	}
	m.mx.Unlock()
}

// TryLock - try locks mutex with context
func (m *PMutex) TryLock(ctx context.Context) bool {

	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.lockST(ctx, false)
	}

	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.lockST(ctx, false)
}

// TryLockF - try locks mutex with context first (out of turn)
func (m *PMutex) TryLockF(ctx context.Context) bool {

	m.mx.Lock()
	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.lockST(ctx, true)
}

// LockD - try locks mutex with time duration
func (m *PMutex) LockD(d time.Duration) bool {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.lockSD(d, false)
	}

	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.lockSD(d, false)
}

// LockDF - try locks mutex with time duration first (out of turn)
func (m *PMutex) LockDF(d time.Duration) bool {
	m.mx.Lock()
	if m.state == 0 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.lockSD(d, true)
}

// RLock - read locks mutex
func (m *PMutex) RLock() {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		m.rlockS(false)
		return
	}

	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()

	// Slow way
	m.rlockS(false)
}

// RLockF - read locks mutex first (out of turn)
func (m *PMutex) RLockF() {
	m.mx.Lock()
	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()

	// Slow way
	m.rlockS(true)
}

// RUnlock - unlocks mutex
func (m *PMutex) RUnlock() {

	m.mx.Lock()

	if m.state > 0 {
		m.state--
		m.chClose()
	} else {
		panic(fmt.Sprintf("PMutex: RUnlock fail (%v)", m.state))
	}

	m.mx.Unlock()
}

// RTryLock - try read locks mutex with context
func (m *PMutex) RTryLock(ctx context.Context) bool {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.rlockST(ctx, false)
	}

	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.rlockST(ctx, false)
}

// RTryLockF - try read locks mutex with context first (out of turn)
func (m *PMutex) RTryLockF(ctx context.Context) bool {
	m.mx.Lock()
	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.rlockST(ctx, true)
}

// RLockD - try read locks mutex with time duration
func (m *PMutex) RLockD(d time.Duration) bool {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.rlockSD(d, false)
	}

	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.rlockSD(d, false)
}

// RLockDF - try read locks mutex with time duration first (out of turn)
func (m *PMutex) RLockDF(d time.Duration) bool {
	m.mx.Lock()
	if m.state >= 0 {
		m.state++
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

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

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == 0 {
				m.state = -1
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return
			}
		}
		m.mx.Unlock()

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

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == 0 {
				m.state = -1
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return true
			}
		}
		m.mx.Unlock()

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

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == 0 {
				m.state = -1
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return true
			}
		}
		m.mx.Unlock()

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
	for {

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state >= 0 {
				m.state++
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return
			}
		}
		m.mx.Unlock()

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
	for {

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state >= 0 {
				m.state++
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return true
			}
		}
		m.mx.Unlock()

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
	for {
		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state >= 0 {
				m.state++
				if f {
					m.pr--
				}
				m.mx.Unlock()
				return true
			}
		}
		m.mx.Unlock()

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
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		m.promoteS(false)
		return
	}

	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()

	// Slow way
	m.promoteS(false)
}

// PromoteF - lock mutex from RLock to Lock first (out of turn)
func (m *PMutex) PromoteF() {

	m.mx.Lock()
	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return
	}
	m.mx.Unlock()

	// Slow way
	m.promoteS(true)
}

// TryPromote - try locks mutex from RLock to Lock with context
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) TryPromote(ctx context.Context) bool {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.promoteST(ctx, false)
	}

	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.promoteST(ctx, false)
}

// TryPromoteF - try locks mutex from RLock to Lock with context first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) TryPromoteF(ctx context.Context) bool {
	m.mx.Lock()
	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.promoteST(ctx, true)
}

// PromoteD - try locks mutex from RLock to Lock with time duration
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) PromoteD(d time.Duration) bool {
	m.mx.Lock()
	if m.pr > 0 {
		m.mx.Unlock()
		return m.promoteSD(d, false)
	}

	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

	// Slow way
	return m.promoteSD(d, false)
}

// PromoteDF - try locks mutex from RLock to Lock with time duration first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *PMutex) PromoteDF(d time.Duration) bool {
	m.mx.Lock()
	if m.state == 1 {
		m.state = -1
		m.mx.Unlock()
		return true
	}
	m.mx.Unlock()

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

	ch := m.chGet()
	for {
		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == m.prom {
				m.state = -1
				if f {
					m.pr--
				}
				m.prom--
				m.promState = m.prom
				m.mx.Unlock()
				return

			}
		}
		m.mx.Unlock()

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

	ch := m.chGet()
	for {

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == m.prom {
				m.state = -1
				if f {
					m.pr--
				}
				m.prom--
				m.promState = m.prom
				m.mx.Unlock()
				return true

			}
		}
		m.mx.Unlock()

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

	ch := m.chGet()
	for {

		m.mx.Lock()
		if f || m.pr == 0 {
			if m.state == m.prom {
				m.state = -1
				if f {
					m.pr--
				}
				m.prom--
				m.promState = m.prom
				m.mx.Unlock()
				return true

			}
		}
		m.mx.Unlock()

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
