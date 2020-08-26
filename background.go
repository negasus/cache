package cache

import (
	"context"
	"sort"
	"sync/atomic"
	"time"
)

func (c *Cache) compact() {
	s := make(map[int]string)
	var t []int

	c.mx.RLock()
	for key, i := range c.storage {
		n := int(i.lu.UnixNano())
		s[n] = key
		t = append(t, n)
	}
	c.mx.RUnlock()

	sort.Ints(t)

	c.mx.Lock()
	defer c.mx.Unlock()
	for _, skey := range t {
		key := s[skey]
		i, ok := c.storage[key]
		if !ok {
			continue
		}
		delete(c.storage, key)
		atomic.AddInt64(&c.size, -int64(len(i.data)))

		if atomic.LoadInt64(&c.size) < c.sizeLimit {
			return
		}
	}
}

func (c *Cache) scanExpired(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		now := time.Now()

		c.expMx.Lock()
		c.mx.RLock()
		for key, i := range c.storage {
			if !i.ttl.IsZero() && i.ttl.Before(now) {
				c.expired[key] = struct{}{}
			}
		}
		c.mx.RUnlock()
		c.expMx.Unlock()

		time.Sleep(c.checkExpireTimout)
	}
}

func (c *Cache) clear(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		c.expMx.Lock()
		c.mx.Lock()
		for key := range c.expired {
			delete(c.expired, key)
			i, ok := c.storage[key]
			if !ok {
				continue
			}
			delete(c.storage, key)
			atomic.AddInt64(&c.size, -int64(len(i.data)))
		}
		c.mx.Unlock()
		c.expMx.Unlock()

		time.Sleep(c.clearExpireTimout)
	}
}
