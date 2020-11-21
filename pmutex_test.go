package mfs

import (
	"context"
	"sync"
	"testing"
	"time"
)

func BenchmarkPMutexLockUnlock(b *testing.B) {
	mx := PMutex{}

	for i := 0; i < b.N; i++ {
		mx.Lock()
		mx.Unlock()
	}
}

func BenchmarkPMutexRLockRUnlock(b *testing.B) {
	mx := PMutex{}

	for i := 0; i < b.N; i++ {
		mx.RLock()
		mx.RUnlock()
	}
}

func BenchmarkPMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := PMutex{}

	for i := 0; i < b.N; i++ {
		mx.TryLock(ctx)
		mx.Unlock()
	}
}

func BenchmarkDT_PMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := PMutex{}

	for i := 0; i < b.N; i++ {
		mx.TryLock(ctx)

		go func() {
			mx.Unlock()
		}()

		mx.TryLock(ctx)
		mx.Unlock()
	}
}

func BenchmarkNT_PMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := PMutex{}

	k := 1000

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(k)

		mx.TryLock(ctx)
		for j := 0; j < k; j++ {
			go func() {
				mx.TryLock(ctx)

				go func() {
					mx.Unlock()
					wg.Done()
				}()
			}()
		}
		mx.Unlock()

		wg.Wait()
	}
}

func BenchmarkN0T_PMutexTryLockUnlock(b *testing.B) {
	ctx := context.Background()
	mx := PMutex{}

	k := 1000

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(k)

		mx.TryLock(ctx)
		for j := 0; j < k; j++ {
			go func() {
				if k%2 == 0 {
					mx.TryLock(ctx)
				} else {
					mx.TryLockF(ctx)
				}

				mx.Unlock()
				wg.Done()
			}()
		}
		mx.Unlock()

		wg.Wait()
	}
}
func BenchmarkPMutexTryRLockRUnlock(b *testing.B) {
	ctx := context.Background()
	mx := PMutex{}

	for i := 0; i < b.N; i++ {
		mx.RTryLock(ctx)
		mx.RUnlock()
	}
}

func TestPMutex(t *testing.T) {

	var mx PMutex

	mx.Lock()
	mx.Unlock()

	mx.Lock()
	t1 := mx.RLockD(time.Millisecond)
	if t1 {
		t.Fatal("TestPMutex t1 fail R lock duration")
	}

	go func() {
		time.Sleep(5 * time.Millisecond)
		mx.Unlock()
	}()

	t2 := mx.RLockD(10 * time.Millisecond)
	t3 := mx.RLockD(10 * time.Millisecond)

	if !t2 {
		t.Fatal("TestPMutex t2 fail R lock duration")
	}
	if !t3 {
		t.Fatal("TestPMutex t2 fail R lock duration")
	}

}

func TestPMutexPromote(t *testing.T) {

	var mx PMutex

	mx.RLock()
	mx.Promote()
	mx.Unlock()

	mx.Lock()
	mx.Reduce()
	mx.RUnlock()

	mx.RLock()
	mx.RLock()
	go func() {
		time.Sleep(5 * time.Millisecond)
		mx.RUnlock()
	}()

	t2 := mx.PromoteD(10 * time.Millisecond)
	if !t2 {
		t.Fatal("TestPMutex t2 fail Promote duration")
	}

	mx.Unlock()

}

// TODO: make normal test
