package atom

import (
	"sync"
	"testing"

	"go.starlark.net/starlark"
	"go.uber.org/atomic"
)

// TestConcurrentFreezeAndSet guards the concurrency-safety the module
// advertises: the frozen flag is read by every mutating method and written by
// Freeze, so it must be an atomic — a plain bool would be a data race under
// `go test -race` (Freeze from one goroutine, a mutator from another). This
// test only makes sense with the race detector, which the module's bar runs.
func TestConcurrentFreezeAndSet(t *testing.T) {
	a := &AtomicInt{val: atomic.NewInt64(0)}
	set := intMethods["set"].BindReceiver(a)
	args := starlark.Tuple{starlark.MakeInt(1)}

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100000; i++ {
			a.Freeze()
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100000; i++ {
			_, _ = intSet(nil, set, args, nil)
		}
	}()
	close(start)
	wg.Wait()
}
