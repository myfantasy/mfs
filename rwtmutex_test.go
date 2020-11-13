package mfs

import (
	"context"
	"sync"
	"testing"
	"time"
)

func BenchmarkRWTMutexLockUnlock(b *testing.B) {
	mx := RWTMutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkRWTMutexRLockRUnlock(b *testing.B) {
	mx := RWTMutex{}

	for i := 0; i < b.N; i++ {
		mx.RLock()
		mx.RUnlock()
	}
}

func BenchmarkRWTMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := RWTMutex{}

	for i := 0; i < b.N; i++ {
		mx.TryLock(ctx)
		mx.Unlock()
	}
}

func BenchmarkDT_RWTMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := RWTMutex{}

	for i := 0; i < b.N; i++ {
		mx.TryLock(ctx)

		go func() {
			mx.Unlock()
		}()

		mx.TryLock(ctx)
		mx.Unlock()
	}
}

func BenchmarkRWTMutexTryRLockRUnlock(b *testing.B) {
	ctx := context.Background()
	mx := RWTMutex{}

	for i := 0; i < b.N; i++ {
		mx.RTryLock(ctx)
		mx.RUnlock()
	}
}

func BenchmarkRWMutexLockUnlock(b *testing.B) {
	mx := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkRWMutexRLockRUnlock(b *testing.B) {
	mx := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		mx.RLock()
		mx.RUnlock()
	}
}

func BenchmarkDT_RWMutexLockUnlock(b *testing.B) {
	mx := sync.RWMutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()

		go func() {
			mx.Unlock()
		}()

		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkMutexLockUnlock(b *testing.B) {
	mx := sync.Mutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkDT_MutexLockUnlock(b *testing.B) {
	mx := sync.Mutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()

		go func() {
			mx.Unlock()
		}()

		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkDT_N_MutexLockUnlock(b *testing.B) {
	mx := sync.Mutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()

		go func() {
		}()

		mx.Unlock()
	}
}

func TestRWTMutex(t *testing.T) {

	var mx RWTMutex

	mx.Lock()
	mx.Unlock()

	mx.Lock()
	t1 := mx.RLockD(time.Millisecond)
	if t1 {
		t.Fatal("TestRWTMutex t1 fail R lock duration")
	}

	go func() {
		time.Sleep(5 * time.Millisecond)
		mx.Unlock()
	}()

	t2 := mx.RLockD(10 * time.Millisecond)
	t3 := mx.RLockD(10 * time.Millisecond)

	if !t2 {
		t.Fatal("TestRWTMutex t2 fail R lock duration")
	}
	if !t3 {
		t.Fatal("TestRWTMutex t2 fail R lock duration")
	}

}

// TODO: make normal test
