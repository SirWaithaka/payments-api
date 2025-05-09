package daraja

import (
	"reflect"
	"sync"
	"time"
)

// Cache caches the authentication token to avoid unnecessary API calls
type Cache[T any] struct {
	mu     sync.RWMutex
	value  T
	expiry time.Time
}

// NewCache creates a new token cache
func NewCache[T any]() *Cache[T] {
	return &Cache[T]{
		mu: sync.RWMutex{},
	}
}

// Get returns the cached value if it's valid, otherwise returns empty string
func (c *Cache[T]) Get() T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if IsEmpty(c.value) || time.Now().After(c.expiry) {
		var t T
		return t
	}
	return c.value
}

// Set caches a value with its expiry time
func (c *Cache[T]) Set(value T, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = value
	c.expiry = expiry
}

// Clear removes the cached token
func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var t T
	c.value = t
	c.expiry = time.Time{}
}

func IsEmpty(value any) bool {
	// Get the reflect value of the cache value
	v := reflect.ValueOf(value)

	// Check if it's a nil interface or pointer
	if (v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr) && v.IsNil() {
		return true
	}

	// Check if it's a zero value
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}
