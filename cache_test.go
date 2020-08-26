package cache

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCache_New(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())

	c := New(ctx, WithCheckExpireTimeout(time.Second), WithClearExpireTimeout(time.Second), WithSizeLimit(100))
	ctxCancel()
	assert.IsType(t, &Cache{}, c)
	assert.Equal(t, time.Second, c.checkExpireTimout)
	assert.Equal(t, time.Second, c.clearExpireTimout)
	assert.Equal(t, int64(100), c.sizeLimit)

	time.Sleep(time.Millisecond * 50)
}

func TestCache_Get_not_exists(t *testing.T) {
	c := New(context.Background())

	_, err := c.Get("foo")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestCache_Get_expired(t *testing.T) {
	c := New(context.Background())
	c.storage["foo"] = &item{
		data: []byte{0x10},
		ttl:  time.Now().Add(-time.Second),
	}
	_, err := c.Get("foo")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
}

func TestCache_Get(t *testing.T) {
	c := New(context.Background())
	c.storage["foo"] = &item{
		data: []byte{0x10},
		ttl:  time.Now().Add(time.Second),
	}
	c.storage["bar"] = &item{
		data: []byte{0x20},
	}
	data, err := c.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x10}, data)

	data, err = c.Get("bar")
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x20}, data)
}

func TestCache_GetOrNew_exists(t *testing.T) {
	c := New(context.Background())
	c.storage["foo"] = &item{
		data: []byte{0x10},
		ttl:  time.Now().Add(time.Second),
	}
	data, err := c.GetOrNew("foo", nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x10}, data)
}

func TestCache_GetOrNew_not_exists_error_cb(t *testing.T) {
	cb := func(key string) ([]byte, error) {
		return nil, fmt.Errorf("error1")
	}

	c := New(context.Background())

	_, err := c.GetOrNew("foo", cb)
	assert.Error(t, err)
	assert.Equal(t, "error1", err.Error())

	_, ok := c.storage["foo"]
	assert.False(t, ok)
}

func TestCache_GetOrNew_not_exists_call_cb(t *testing.T) {
	cb := func(key string) ([]byte, error) {
		return []byte{0x30}, nil
	}

	c := New(context.Background())

	data, err := c.GetOrNew("foo", cb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x30}, data)

	i, ok := c.storage["foo"]
	assert.True(t, ok)
	assert.Equal(t, []byte{0x30}, i.data)
}

func TestCache_Has_not_exists(t *testing.T) {
	c := New(context.Background())

	ok := c.Has("foo")
	assert.False(t, ok)
}

func TestCache_Has_expired(t *testing.T) {
	c := New(context.Background())

	c.storage["foo"] = &item{
		data: []byte{0x10},
		ttl:  time.Now().Add(-time.Second),
	}

	ok := c.Has("foo")
	assert.False(t, ok)
}

func TestCache_Has_without_ttl(t *testing.T) {
	c := New(context.Background())

	c.storage["foo"] = &item{
		data: []byte{0x10},
	}

	ok := c.Has("foo")
	assert.True(t, ok)
}

func TestCache_Has_not_expired(t *testing.T) {
	c := New(context.Background())

	c.storage["foo"] = &item{
		data: []byte{0x10},
		ttl:  time.Now().Add(time.Second),
	}

	ok := c.Has("foo")
	assert.True(t, ok)
}

func TestCache_Delete(t *testing.T) {
	c := New(context.Background())

	c.storage["foo"] = &item{
		data: []byte{0x10},
	}

	c.Delete("foo")
	_, ok := c.storage["foo"]
	assert.False(t, ok)

	c.Delete("bar") // not exists
}

func TestCache_Put_over_size(t *testing.T) {
	c := New(context.Background())
	c.sizeLimit = 1

	c.Put("foo", []byte{0x10, 0x20})

	_, ok := c.storage["foo"]
	assert.False(t, ok)
}

func TestCache_Put(t *testing.T) {
	c := New(context.Background())

	c.Put("foo", []byte{0x10})

	i, ok := c.storage["foo"]
	assert.True(t, ok)
	assert.Equal(t, []byte{0x10}, i.data)
}

func TestCache_PutWithTTL_over_size(t *testing.T) {
	c := New(context.Background())
	c.sizeLimit = 1

	c.PutWithTTL("foo", []byte{0x10, 0x20}, time.Second)

	_, ok := c.storage["foo"]
	assert.False(t, ok)
}

func TestCache_PutWithTTL(t *testing.T) {
	c := New(context.Background())

	c.PutWithTTL("foo", []byte{0x10}, time.Second)

	i, ok := c.storage["foo"]
	assert.True(t, ok)
	assert.Equal(t, []byte{0x10}, i.data)
	assert.InDelta(t, time.Now().Add(time.Second).Unix(), i.ttl.Unix(), 1)
}

func TestCache_GetOrNewWithTTL_exists(t *testing.T) {
	cb := func(key string) ([]byte, error) {
		return []byte{0x30}, nil
	}

	c := New(context.Background())
	c.storage["foo"] = &item{
		data: []byte{0x20},
	}

	data, err := c.GetOrNewWithTTL("foo", time.Second, cb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x20}, data)
}

func TestCache_GetOrNewWithTTL_error_cb(t *testing.T) {
	cb := func(key string) ([]byte, error) {
		return nil, fmt.Errorf("error1")
	}

	c := New(context.Background())

	_, err := c.GetOrNewWithTTL("foo", time.Second, cb)
	assert.Error(t, err)
	assert.Equal(t, "error1", err.Error())
}

func TestCache_GetOrNewWithTTL(t *testing.T) {
	cb := func(key string) ([]byte, error) {
		return []byte{0x30}, nil
	}

	c := New(context.Background())

	data, err := c.GetOrNewWithTTL("foo", time.Second, cb)
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x30}, data)

	i, ok := c.storage["foo"]
	assert.True(t, ok)
	assert.Equal(t, []byte{0x30}, i.data)
	assert.InDelta(t, time.Now().Add(time.Second).Unix(), i.ttl.Unix(), 1)
}

func TestCache_clear_with_get(t *testing.T) {
	c := New(context.Background(), WithClearExpireTimeout(time.Millisecond*100), WithCheckExpireTimeout(time.Hour))

	c.PutWithTTL("foo", []byte{0x10}, time.Millisecond*50)

	time.Sleep(time.Millisecond * 60) // wait for expire

	_, err := c.Get("foo")
	assert.Equal(t, ErrNotFound, err)

	time.Sleep(time.Millisecond * 150) // wait for clear

	_, ok := c.storage["foo"]
	assert.False(t, ok)
}

func TestCache_clear_with_check(t *testing.T) {
	c := New(context.Background(), WithClearExpireTimeout(time.Millisecond*100), WithCheckExpireTimeout(time.Millisecond*100))

	c.PutWithTTL("foo", []byte{0x10}, time.Millisecond*50)

	time.Sleep(time.Millisecond * 250) // wait for clear

	_, ok := c.storage["foo"]
	assert.False(t, ok)
}

func TestPut_compact(t *testing.T) {
	c := New(context.Background())
	c.sizeLimit = 20
	c.size = 15

	now := time.Now()

	c.storage["1"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -1)}
	c.storage["2"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -2)}
	c.storage["3"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -3)}

	c.Put("5", []byte("1234"))

	time.Sleep(time.Millisecond * 50) // time for run 'compact'

	assert.Equal(t, int64(19), c.size)
	assert.Equal(t, 4, len(c.storage))

	c.Put("6", []byte("1234567"))

	time.Sleep(time.Millisecond * 50) // time for run 'compact'

	assert.Equal(t, int64(16), c.size)
	assert.Equal(t, 3, len(c.storage))

	_, ok := c.storage["6"]
	assert.True(t, ok)
	_, ok = c.storage["5"]
	assert.True(t, ok)
	_, ok = c.storage["1"]
	assert.True(t, ok)
}

func TestPutWithTTL_compact(t *testing.T) {
	c := New(context.Background())
	c.sizeLimit = 20
	c.size = 15

	now := time.Now()

	c.storage["1"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -1)}
	c.storage["2"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -2)}
	c.storage["3"] = &item{data: []byte("12345"), lu: now.Add(time.Second * -3)}

	c.PutWithTTL("5", []byte("1234"), time.Hour)

	time.Sleep(time.Millisecond * 50) // time for run 'compact'

	assert.Equal(t, int64(19), c.size)
	assert.Equal(t, 4, len(c.storage))

	c.PutWithTTL("6", []byte("1234567"), time.Hour)

	time.Sleep(time.Millisecond * 50) // time for run 'compact'

	assert.Equal(t, int64(16), c.size)
	assert.Equal(t, 3, len(c.storage))

	_, ok := c.storage["6"]
	assert.True(t, ok)
	_, ok = c.storage["5"]
	assert.True(t, ok)
	_, ok = c.storage["1"]
	assert.True(t, ok)
}
