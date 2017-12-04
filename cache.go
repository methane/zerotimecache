package zerotimecache

import (
	"sync"
	"time"
)

type Cache struct {
	m   sync.Mutex
	t   time.Time
	res interface{}
	err error
}

// DoDelay calls f() unless cached value which is created after calling time is available.
// If delay>0, sleep while it before calling f().  You can use it for sharing result
// value from more concurrent callers.
func (c *Cache) DoDelay(delay time.Duration, f func() (interface{}, error)) (v interface{}, err error) {
	t0 := time.Now()
	c.m.Lock()
	defer c.m.Unlock()

	// If c.t is newer, return cached value
	// We can't use `>=` because some system may produce exactly same time for multiple times.
	if c.t.Sub(t0) > 0 {
		return c.res, c.err
	}
	if delay > 0 {
		time.Sleep(delay)
	}

	c.t = time.Now()
	c.res, c.err = f()
	return c.res, c.err
}

// Do calls f() unless cached value which is created after calling time is available.
func (c *Cache) Do(f func() (interface{}, error)) (v interface{}, err error) {
	return c.DoDelay(0, f)
}
