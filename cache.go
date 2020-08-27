package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultClearExpireTimeout = time.Hour
	defaultCheckExpireTimeout = time.Hour
	defaultSizeLimit          = int64(1024 * 1024)
)

var (
	ErrNotFound = errors.New("not found")
)

type GetCallback func(string) ([]byte, error)

type OptionFunc func(c *Cache)

func WithClearExpireTimeout(t time.Duration) OptionFunc {
	return func(c *Cache) {
		c.clearExpireTimout = t
	}
}

func WithCheckExpireTimeout(t time.Duration) OptionFunc {
	return func(c *Cache) {
		c.checkExpireTimout = t
	}
}

func WithSizeLimit(l int64) OptionFunc {
	return func(c *Cache) {
		c.sizeLimit = l
	}
}

type item struct {
	data []byte
	ttl  time.Time
	lu   time.Time
}

type Cache struct {
	mx      sync.RWMutex
	storage map[string]*item

	expMx   sync.RWMutex
	expired map[string]struct{}

	clearExpireTimout time.Duration
	checkExpireTimout time.Duration
	sizeLimit         int64
	size              int64
}

func New(ctx context.Context, options ...OptionFunc) *Cache {
	c := &Cache{
		storage:           map[string]*item{},
		expired:           map[string]struct{}{},
		clearExpireTimout: defaultClearExpireTimeout,
		checkExpireTimout: defaultCheckExpireTimeout,
		sizeLimit:         defaultSizeLimit,
	}

	for _, o := range options {
		o(c)
	}

	go c.clear(ctx)
	go c.scanExpired(ctx)

	return c
}

func (c *Cache) Get(key string) ([]byte, error) {
	c.mx.RLock()
	i, ok := c.storage[key]
	c.mx.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}

	if !i.ttl.IsZero() && i.ttl.Before(time.Now()) {
		c.expMx.Lock()
		c.expired[key] = struct{}{}
		c.expMx.Unlock()
		return nil, ErrNotFound
	}

	i.lu = time.Now()

	return i.data, nil
}

func (c *Cache) GetOrNew(key string, cb GetCallback) ([]byte, error) {
	data, err := c.Get(key)
	if err == nil {
		return data, nil
	}

	data, err = cb(key)
	if err != nil {
		return nil, err
	}

	c.Put(key, data)

	return data, nil
}

func (c *Cache) GetOrNewWithTTL(key string, ttl time.Duration, cb GetCallback) ([]byte, error) {
	data, err := c.Get(key)
	if err == nil {
		return data, nil
	}

	data, err = cb(key)
	if err != nil {
		return nil, err
	}

	c.PutWithTTL(key, data, ttl)

	return data, nil
}

func (c *Cache) Has(key string) bool {
	c.mx.RLock()
	defer c.mx.RUnlock()

	i, ok := c.storage[key]
	if !ok {
		return false
	}

	return i.ttl.IsZero() || i.ttl.After(time.Now())
}

func (c *Cache) Delete(key string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	i, ok := c.storage[key]
	if !ok {
		return
	}

	delete(c.storage, key)

	atomic.AddInt64(&c.size, -int64(len(i.data)))
}

func (c *Cache) Put(key string, data []byte) {
	if int64(len(data)) >= c.sizeLimit {
		return
	}

	c.mx.Lock()
	c.expMx.Lock()
	c.storage[key] = &item{
		data: data,
		lu:   time.Now(),
	}
	delete(c.expired, key)
	c.expMx.Unlock()
	c.mx.Unlock()

	n := atomic.AddInt64(&c.size, int64(len(data)))
	if n > c.sizeLimit {
		go c.compact()
	}
}

func (c *Cache) PutWithTTL(key string, data []byte, ttl time.Duration) {
	if int64(len(data)) >= c.sizeLimit {
		return
	}

	c.mx.Lock()
	c.expMx.Lock()
	c.storage[key] = &item{
		data: data,
		ttl:  time.Now().Add(ttl),
		lu:   time.Now(),
	}
	delete(c.expired, key)
	c.expMx.Unlock()
	c.mx.Unlock()

	n := atomic.AddInt64(&c.size, int64(len(data)))
	if n > c.sizeLimit {
		go c.compact()
	}
}
