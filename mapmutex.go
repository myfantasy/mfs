package mfs

import (
	"context"
	"sync"
	"time"
)

// MapMutex - Mutex is named (each name locks independently)
// Read Write Try Mutex with change priority (Promote and Reduce)
// F methods (like LockF and TryLockF) Locks mutex if mutex already locked then this methods will be first in lock queue
// Promote - lock mutex from RLock to Lock
// Reduce - lock mutex from Lock to RLock
type MapMutex struct {
	mxGlobal sync.Mutex
	mapMx    map[string]*PMutex
}

// GetMutex - gets PMutex from MapMutex by name
func (m *MapMutex) GetMutex(name string) *PMutex {
	m.mxGlobal.Lock()
	mx, ok := m.mapMx[name]
	if !ok {
		mx = &PMutex{}
		m.mapMx[name] = mx
	}
	m.mxGlobal.Unlock()

	return mx
}

// DoGlobal - do some action under global lock
func (m *MapMutex) DoGlobal(do func()) {
	m.mxGlobal.Lock()
	do()
	m.mxGlobal.Unlock()
}

// Lock - locks mutex
func (m *MapMutex) Lock(name string) {
	mx := m.GetMutex(name)
	mx.Lock()
}

// LockF - locks mutex first (out of turn)
func (m *MapMutex) LockF(name string) {
	mx := m.GetMutex(name)
	mx.LockF()
}

// Unlock - unlocks mutex
func (m *MapMutex) Unlock(name string) {
	mx := m.GetMutex(name)
	mx.Unlock()
}

// Reduce - lock mutex from Lock to RLock
func (m *MapMutex) Reduce(name string) {
	mx := m.GetMutex(name)
	mx.Reduce()
}

// TryLock - try locks mutex with context
func (m *MapMutex) TryLock(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.TryLock(ctx)
}

// TryLockF - try locks mutex with context first (out of turn)
func (m *MapMutex) TryLockF(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.TryLockF(ctx)
}

// LockD - try locks mutex with time duration
func (m *MapMutex) LockD(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.LockD(d)
}

// LockDF - try locks mutex with time duration first (out of turn)
func (m *MapMutex) LockDF(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.LockDF(d)
}

// RLock - read locks mutex
func (m *MapMutex) RLock(name string) {
	mx := m.GetMutex(name)
	mx.RLock()
}

// RLockF - read locks mutex first (out of turn)
func (m *MapMutex) RLockF(name string) {
	mx := m.GetMutex(name)
	mx.RLockF()
}

// RUnlock - unlocks mutex
func (m *MapMutex) RUnlock(name string) {
	mx := m.GetMutex(name)
	mx.RUnlock()
}

// RTryLock - try read locks mutex with context
func (m *MapMutex) RTryLock(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.RTryLock(ctx)
}

// RTryLockF - try read locks mutex with context first (out of turn)
func (m *MapMutex) RTryLockF(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.RTryLockF(ctx)
}

// RLockD - try read locks mutex with time duration
func (m *MapMutex) RLockD(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.RLockD(d)
}

// RLockDF - try read locks mutex with time duration first (out of turn)
func (m *MapMutex) RLockDF(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.RLockDF(d)
}

// Promote - lock mutex from RLock to Lock
func (m *MapMutex) Promote(name string) {
	mx := m.GetMutex(name)
	mx.Promote()
}

// PromoteF - lock mutex from RLock to Lock first (out of turn)
func (m *MapMutex) PromoteF(name string) {
	mx := m.GetMutex(name)
	mx.PromoteF()
}

// TryPromote - try locks mutex from RLock to Lock with context
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *MapMutex) TryPromote(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.TryPromote(ctx)
}

// TryPromoteF - try locks mutex from RLock to Lock with context first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *MapMutex) TryPromoteF(ctx context.Context, name string) bool {
	mx := m.GetMutex(name)
	return mx.TryPromoteF(ctx)
}

// PromoteD - try locks mutex from RLock to Lock with time duration
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *MapMutex) PromoteD(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.PromoteD(d)
}

// PromoteDF - try locks mutex from RLock to Lock with time duration first (out of turn)
// !!! If returns false then mutex is UNLOCKED if true mutex is locked as Lock
func (m *MapMutex) PromoteDF(d time.Duration, name string) bool {
	mx := m.GetMutex(name)
	return mx.PromoteDF(d)
}
