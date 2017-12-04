package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/methane/zerotimecache"
	"golang.org/x/sync/singleflight"
)

type NameRepo struct {
	m      sync.Mutex
	names  []string
	called int64 // GetNames() called
}

func (r *NameRepo) AddName(name string) {
	r.m.Lock()
	r.names = append(r.names, name)
	r.m.Unlock()
}

func (r *NameRepo) GetNames() map[string]bool {
	res := make(map[string]bool)

	r.m.Lock()
	r.called++
	for _, n := range r.names {
		res[n] = true
	}
	r.m.Unlock()
	return res
}

func Check(test string, nr *NameRepo, f func() map[string]bool) {
	const N = 100
	var errorCount int64

	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		name := fmt.Sprintf("worker-%d", i)
		go func(name string) {
			for j := 0; j < 10; j++ {
				n := fmt.Sprintf("%s-%d", name, j)
				nr.AddName(n)
				result := f()
				if !result[n] {
					atomic.AddInt64(&errorCount, 1)
				}
			}
			wg.Done()
		}(name)
	}
	wg.Wait()

	fmt.Printf("%s: called %d times, %d errors\n", test, nr.called, errorCount)
}

func SampleDirect() {
	var nr NameRepo
	Check("direct", &nr, nr.GetNames)
}

func SampleSingleFlight() {
	var nr NameRepo
	var group singleflight.Group

	f := func() map[string]bool {
		v, _, _ := group.Do("", func() (interface{}, error) {
			return nr.GetNames(), nil
		})
		return v.(map[string]bool)
	}
	Check("singleflight", &nr, f)
}

func SampleZTC() {
	var nr NameRepo
	var cache zerotimecache.Cache

	f := func() map[string]bool {
		v, _ := cache.Do(func() (interface{}, error) {
			return nr.GetNames(), nil
		})
		return v.(map[string]bool)
	}
	Check("ZeroTimeCache", &nr, f)
}

func SampleZTCDelay() {
	var nr NameRepo
	var cache zerotimecache.Cache

	f := func() map[string]bool {
		v, _ := cache.DoDelay(time.Millisecond, func() (interface{}, error) {
			return nr.GetNames(), nil
		})
		return v.(map[string]bool)
	}
	Check("ZeroTimeCacheDelay", &nr, f)
}

func main() {
	SampleDirect()
	SampleSingleFlight()
	SampleZTC()
	SampleZTCDelay()
}
