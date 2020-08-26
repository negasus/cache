# Bytes Cache with TTL and MaxSize option

## Install

```
go get -u github.com/negasus/cache
```

## Usage example

```
c := cache.New(ctx, cache.WithSizeLimit(1024*1024*5)

c.Put("foo", []byte("bar"))

value, err := c.Get("foo")
```

## Constructor

### `New(ctx context.Context, options ...OptionFunc) *Cache`

You should use a `New` function for create the Cache instance.

You can use `WithCancel` context for stop goroutines, which the Cache starts after creating.

You can use options for change some default values 

## Options

### `WithClearExpireTimeout(t time.Duration)` 

> default: `time.Hour`

The time interval for clear expired values 

### `WithCheckExpireTimeout(t time.Duration)` 

> default: `time.Hour`

The time interval for full scan cache storage for expired values and mark it for delete 

### `WithSizeLimit(l int64)` 

> default: `1048576` (1 Mb)

Max cache size. If reached, `compacting` will be running 

## Compacting

Scan the cache storage, sort items by `last used` field and delete items while storage size greater than `cache.sizeLimit` 

## API

### `Get(key string) ([]byte, error)`

Get a data from the cache.

If a data not found or expired, `ErrNotFound` will be returned

### `GetOrNew(key string, cb GetCallback) ([]byte, error)`

Get a data from the cache or create new cache item with callback function (and return a result)

### `GetOrNewWithTTL(key string, ttl time.Duration, cb GetCallback) ([]byte, error)`

Get a data from the cache or create new cache item with callback function (and return a result)

### `Has(key string) bool`

Check key exists

### `Delete(key string)`

Delete an item from the cache

### `Put(key string, data []byte)`

Put the data to the cache

If data length greater or equal `cache.sizeLimit`, data will not be stored.

If the cache storage reached the `cache.sizeLimit`, `compacting` will be running.

### `PutWithTTL(key string, data []byte, ttl time.Duration)`

Put the data to the cache

If data length greater or equal `cache.sizeLimit`, data will not be stored.

If the cache storage reached the `cache.sizeLimit`, `compacting` will be running.



