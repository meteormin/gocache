package gocache

import (
	"time"
)

var memCache *MemCache

// New initializes a new MemCache with the given maxSize and starts a goroutine to periodically check for expired instances.
//
// Parameter: maxSize uint - the maximum size of the MemCache
// Returns: *MemCache - a pointer to the newly created MemCache
func New(maxSize uint) *MemCache {
	memCache = NewMemCache(maxSize)

	go func() {
		for {
			deleteExired(memCache, 10)

			time.Sleep(time.Second)

			if <-memCache.ch {
				return
			}
		}
	}()

	return memCache
}

// MaxSize returns the maximum size of the memCache.
//
// uint.
func MaxSize() uint {
	return maxSize(memCache)
}

// Size returns the size by calling getSize on memCache.
//
// Returns an integer.
func Size() int {
	return getSize(memCache)
}

// Count returns the count of the given parameter.
//
// Returns an integer.
func Count() int {
	return count(memCache)
}

// Keys returns the keys of the memCache.
//
// Returns a slice of strings.
func Keys() []string {
	return keys(memCache)
}

// Values returns the values of the memCache.
//
// Returns a slice of Instance[interface{}].
func Values() []Instance[interface{}] {
	return values(memCache)
}

// Value retrieves the value associated with the given key from the memory cache.
//
// key string
// *Instance[interface{}]
func Value(key string) *Instance[interface{}] {
	return value(memCache, key)
}

// Exists checks if a key exists in the memory cache.
//
// It takes a string key as a parameter and returns a boolean value.
func Exists(key string) bool {
	return exists(memCache, key)
}

// Get retrieves a value from the memory cache using the provided key and stores it in the dst interface{}.
//
// key string, dst interface{}
func Get(key string, dst interface{}) {
	get(memCache, key, dst)
}

// Set sets a value in the memory cache.
//
// key: the key to set in the cache
// exp: the expiration time duration for the key
// src: the value to set in the cache
// error: an error if the operation fails
func Set(key string, exp time.Duration, src interface{}) error {
	return set(memCache, key, exp, src)
}

// Delete Deletes a key from the memCache.
//
// Parameter:
//
//	key - the key to be Deleted
//
// Return type:
//
//	bool
func Delete(key string) bool {
	return delete(memCache, key)
}

// Clear clears the instances in the memory cache.
func Clear() {
	clear(memCache)
}

// Resolve resolves the value for the given key using the provided resolver function.
//
// key string, exp time.Duration, resolver[T]
// (T, error)
func Resolve[T interface{}](key string, exp time.Duration, resolver Resolver[T]) (T, error) {
	return resolve(memCache, key, exp, resolver)
}

// GetStat returns a Stat struct with Count, Keys, MaxSize, Size, Usage, and Values.
//
// Returns a Stat struct.
func GetStat() Stat {
	return getStat(memCache)
}

// Close closes the memory cache by sending a signal through the memCache channel.
//
// No parameters.
// No return types.
func Close() {
	memCache.ch <- true
}

func IsRunning() bool {
	return memCache != nil
}
