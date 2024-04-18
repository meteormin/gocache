package gocache

import (
	"time"
)

var m *MemCache

// New initializes a new MemCache with the given maxSize and starts a goroutine to periodically check for expired instances.
//
// Parameter: maxSize uint - the maximum size of the MemCache
// Returns: *MemCache - a pointer to the newly created MemCache
func New(maxSize uint) *MemCache {
	m = NewMemCache(maxSize)

	go func() {
		for {
			if len(m.instances) <= 10 {
				for _, instance := range m.instances {
					if instance.IsExpired() {
						Delete(instance.Key)
					}
				}
			} else {
				tmp := make([]int, 0, 10)
				for i := 0; i < 10; i++ {
					for j := 0; j < 10; j++ {
						randomNumber := randResolver(&tmp)
						if m.instances[randomNumber].IsExpired() {
							Delete(m.instances[randomNumber].Key)
						}
					}
				}
			}

			time.Sleep(time.Second)

			if <-m.ch {
				return
			}
		}
	}()

	return m
}

// MaxSize returns the maximum size of the
//
// uint.
func MaxSize() uint {
	return maxSize(m)
}

// Size returns the size by calling getSize on
//
// Returns an integer.
func Size() int {
	return getSize(m)
}

// Count returns the count of the given parameter.
//
// Returns an integer.
func Count() int {
	return count(m)
}

// Keys returns the keys of the
//
// Returns a slice of strings.
func Keys() []string {
	return keys(m)
}

// Values returns the values of the
//
// Returns a slice of Instance[interface{}].
func Values() []Instance[interface{}] {
	return values(m)
}

// Value retrieves the value associated with the given key from the memory cache.
//
// key string
// *Instance[interface{}]
func Value(key string) *Instance[interface{}] {
	return value(m, key)
}

// Exists checks if a key exists in the memory cache.
//
// It takes a string key as a parameter and returns a boolean value.
func Exists(key string) bool {
	return exists(m, key)
}

// Get retrieves a value from the memory cache using the provided key and stores it in the dst interface{}.
//
// key string, dst interface{}
func Get(key string, dst interface{}) {
	get(m, key, dst)
}

// Set sets a value in the memory cache.
//
// key: the key to set in the cache
// exp: the expiration time duration for the key
// src: the value to set in the cache
// error: an error if the operation fails
func Set(key string, exp time.Duration, src interface{}) error {
	return set(m, key, exp, src)
}

// Delete Deletes a key from the
//
// Parameter:
//
//	key - the key to be Deleted
//
// Return type:
//
//	bool
func Delete(key string) bool {
	return delete(m, key)
}

// Clear clears the instances in the memory cache.
func Clear() {
	clear(m)
}

// Resolve resolves the value for the given key using the provided resolver function.
//
// key string, exp time.Duration, resolver[T]
// (T, error)
func Resolve[T interface{}](key string, exp time.Duration, resolver Resolver[T]) (T, error) {
	return resolve(m, key, exp, resolver)
}

// GetStat returns a Stat struct with Count, Keys, MaxSize, Size, Usage, and Values.
//
// Returns a Stat struct.
func GetStat() Stat {
	return getStat(m)
}

// Close closes the memory cache by sending a signal through the memCache channel.
//
// No parameters.
// No return types.
func Close() {
	m.ch <- true
}

func IsRunning() bool {
	return m != nil
}
