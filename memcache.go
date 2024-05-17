package gocache

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

const (
	dstMustNotBeNil = "dst must not be nil pointer"
)

// MemCache is a memory cache implementation.
type MemCache struct {
	ch        chan int
	instances []Instance[interface{}]
	maxSize   uint
}

// Stat is a struct with Count, Keys, MaxSize, Size, Usage, and Values.
type Stat struct {
	Count   int                     `json:"count"`
	Keys    []string                `json:"keys"`
	Size    int                     `json:"size"`
	MaxSize uint                    `json:"maxSize"`
	Usage   float64                 `json:"usage"`
	Values  []Instance[interface{}] `json:"values"`
}

// Resolver is a function that returns a value and an error.
type Resolver[T interface{}] func() (T, error)

// Instance is a struct with Key, Size, Value, Resolver, ExpiresAt, and ExpiresIn.
type Instance[T interface{}] struct {
	Key       string        `json:"key"`
	Size      int           `json:"size"`
	Value     T             `json:"value"`
	Resolver  interface{}   `json:"-"`
	ExpiresAt time.Time     `json:"expiresAt"`
	ExpiresIn time.Duration `json:"expiresIn"`
}

// / GetValue returns the value of the instance.
func (i Instance[T]) GetValue() *T {
	if time.Now().After(i.ExpiresAt) {
		return nil
	}

	i.ExpiresAt = time.Now().Add(i.ExpiresIn)

	return &i.Value
}

// IsExpired checks if the instance is expired.
func (i Instance[T]) IsExpired() bool {
	if i.ExpiresAt.IsZero() || i.ExpiresIn == 0 {
		return false
	}

	return i.ExpiresAt.Before(time.Now())
}

// New initializes the memory cache with the provided configuration.
// maxSize is the maximum byte size that can be stored in the cache.
func NewMemCache(maxSize uint) *MemCache {
	return &MemCache{
		ch:        make(chan int),
		instances: make([]Instance[interface{}], 0),
		maxSize:   maxSize,
	}
}

// maxSize returns the maximum size of the MemCache.
//
// m *MemCache
// uint
func maxSize(m *MemCache) uint {
	return m.maxSize
}

// getSize returns the size of the MemCache.
//
// Parameter: m *MemCache
// Return type: int
func getSize(m *MemCache) int {
	return sizeOf(m)
}

// count returns the number of instances in the MemCache.
//
// Parameter:
//
//	m *MemCache: pointer to a MemCache struct
//
// Return type:
//
//	int
func count(m *MemCache) int {
	return len(m.instances)
}

// keys returns the list of keys from the MemCache instance.
//
// m *MemCache - a pointer to the MemCache instance
// []string - a slice of strings containing the keys
func keys(m *MemCache) []string {
	keys := make([]string, 0)

	for _, instance := range m.instances {
		keys = append(keys, instance.Key)
	}

	return keys
}

// values returns the instances stored in the MemCache.
//
// m *MemCache - a pointer to the MemCache
// []Instance[interface{}] - a slice of Instance[interface{}]
func values(m *MemCache) []Instance[interface{}] {
	return m.instances
}

// value retrieves the instance from the MemCache associated with the given key.
//
// Parameters:
// - m: pointer to MemCache
// - key: string
// Returns:
// - pointer to Instance[interface{}]
func value(m *MemCache, key string) *Instance[interface{}] {
	for _, instance := range m.instances {
		if instance.Key == key {
			if instance.IsExpired() {
				delete(m, key)
				return nil
			}

			return &instance
		}
	}

	return nil
}

// exists checks if a key exists in the MemCache.
//
// Parameters:
//
//	m *MemCache - pointer to the MemCache object
//	key string - the key to check for existence
//
// Returns:
//
//	bool - true if the key exists and is not expired, false otherwise
func exists(m *MemCache, key string) bool {
	for _, instance := range m.instances {
		if instance.Key == key {
			if instance.IsExpired() {
				delete(m, instance.Key)
				return false
			}
			return true
		}
	}

	return false
}

// get retrieves a value from MemCache based on a key and stores it into the provided destination pointer.
//
// Parameters:
//   - m: a pointer to the MemCache instance.
//   - key: the key to look up in the MemCache.
//   - dst: a pointer to the destination where the retrieved value will be stored.
func get(m *MemCache, key string, dst interface{}) {
	if reflect.Ptr != reflect.TypeOf(dst).Kind() {
		panic("dst must be a pointer")
	}

	var vPtr *interface{}
	for _, instance := range m.instances {
		if instance.IsExpired() {
			delete(m, key)
			continue
		}

		if instance.Key == key {
			vPtr = instance.GetValue()
			break
		}
	}

	if vPtr == nil {
		return
	}

	dstValue := reflect.ValueOf(dst)
	if dstValue.IsNil() {
		panic(dstMustNotBeNil)
	}

	insType := reflect.TypeOf(*vPtr).Kind()
	dsType := reflect.TypeOf(dst).Elem().Kind()

	if insType == dsType {
		dstValue.Elem().Set(reflect.ValueOf(*vPtr))
	} else if dstValue.Elem().Kind() == reflect.Ptr {
		if dstValue.Elem().IsNil() {
			panic(dstMustNotBeNil)
		}

		ptrDstValue := dstValue.Elem()
		if ptrDstValue.IsNil() {
			panic(dstMustNotBeNil)
		}

		ptrDstValue.Elem().Set(reflect.ValueOf(*vPtr))
	}
}

// set sets a value in the MemCache with the given key and expiration time.
//
// Parameters:
//   - m: pointer to the MemCache where the value will be set
//   - key: the key to identify the value
func set(m *MemCache, key string, exp time.Duration, src interface{}) error {
	instance := Instance[interface{}]{
		Key:       key,
		Size:      sizeOf(src),
		ExpiresIn: exp,
		ExpiresAt: time.Now().Add(exp),
	}

	refSrcValue := reflect.ValueOf(src)
	if refSrcValue.Kind() == reflect.Ptr {

		if refSrcValue.IsNil() {
			panic("src cannot be nil")
		}

		instance.Value = refSrcValue.Elem().Interface()
	} else {
		instance.Value = src
	}

	delete(m, key)

	if isMaxSize(m, instance) {
		return maxSizeError(m, instance)
	}

	m.instances = append(m.instances, instance)

	return nil
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
func delete(m *MemCache, key string) bool {
	for i, instance := range m.instances {
		if instance.Key == key {
			m.instances = append(m.instances[:i], m.instances[i+1:]...)
			return true
		}
	}

	return false
}

// Clear clears the instances in the memory cache.
func clear(m *MemCache) {
	m.instances = make([]Instance[interface{}], 0)
}

// Resolve resolves the value for the given key using the provided resolver function.
//
// key string, exp time.Duration, resolver[T]
// (T, error)
func resolve[T interface{}](m *MemCache, key string, exp time.Duration, resolver Resolver[T]) (T, error) {
	if resolver == nil {
		panic("resolver cannot be nil")
	}

	if reflect.TypeOf(resolver).Kind() != reflect.Func {
		panic("resolver must be a function")
	}

	if reflect.TypeOf(resolver).NumIn() != 0 {
		panic("resolver must have no input parameters")
	}

	for _, instance := range m.instances {
		if instance.Key == key {
			vPtr := instance.GetValue()
			if vPtr != nil {
				v := *vPtr
				return v.(T), nil
			} else {
				return instance.Resolver.(Resolver[T])()
			}
		}
	}

	v, err := resolver()
	if err != nil {
		return v, err
	}

	instance := Instance[interface{}]{
		Key:       key,
		Size:      sizeOf(v),
		Value:     v,
		Resolver:  resolver,
		ExpiresIn: exp,
		ExpiresAt: time.Now().Add(exp),
	}

	if isMaxSize(m, instance) {
		return v, maxSizeError(m, instance)
	}

	m.instances = append(m.instances, instance)

	return v, nil
}

// GetStat returns a Stat struct with Count, Keys, MaxSize, Size, Usage, and Values.
//
// Returns a Stat struct.
func getStat(m *MemCache) Stat {
	return Stat{
		Count:   count(m),
		Keys:    keys(m),
		MaxSize: maxSize(m),
		Size:    getSize(m),
		Usage:   float64(getSize(m)) / float64(maxSize(m)) * 100.0,
		Values:  values(m),
	}
}

// isMaxSize checks if the size of the given instance plus the size of the memCache.instances exceeds the maximum Size
// if the size exceeds the maximum size, it returns true, otherwise it returns false.
// if the size is 0 then unlimited cache Size
//
// Parameters:
// - instance: an Instance of interface{}.
//
// Returns:
// - bool: true if the total size exceeds the maximum size, false otherwise.
func isMaxSize(m *MemCache, value interface{}) bool {
	if m.maxSize <= 0 {
		return false
	}

	s := sizeOf(value)
	ss := sizeOf(m.instances)
	tot := s + ss
	return tot >= int(m.maxSize)
}

// maxSizeError returns an error indicating that the maximum size has been exceeded.
//
// It takes an integer parameter `size` which represents the size that exceeded the maximum Size
// The function returns an error of type `error`.
func maxSizeError(m *MemCache, value interface{}) error {
	return fmt.Errorf("max size exceeded, max size: %d, current size: %d, instance size: %d",
		m.maxSize, sizeOf(m.instances), sizeOf(value))
}

// randResolver generates a random number that is not already present in the input slice.
//
// Parameters:
// tmp *[]int - a pointer to a slice of integers.
//
// Return type:
// int - the generated random number.
func randResolver(tmp *[]int) {
	if cap(*tmp) == len(*tmp) {
		return
	}

	randSource := rand.NewSource(time.Now().UnixNano())
	randNew := rand.New(randSource)
	randomNumber := randNew.Intn(cap(*tmp))
	for _, t := range *tmp {
		if randomNumber == t {
			randResolver(tmp)
		}
	}

	*tmp = append(*tmp, randomNumber)
}

func deleteExiredAll(m *MemCache) int {
	deleted := 0
	for _, instance := range m.instances {
		if instance.IsExpired() {
			if delete(m, instance.Key) {
				deleted++
			}
		}
	}

	return deleted
}

// deleteExired deletes expired instances from the MemCache based on the given size.
//
// Parameters:
// m *MemCache - a pointer to the MemCache object.
// size int - the size parameter for deletion.
func deleteExired(m *MemCache, size int) {
	deleted := 0
	if count(m) <= size {
		m.ch <- deleteExiredAll(m)
		return
	}

	tmp := make([]int, 0, size)
	for len(tmp) < size {
		randResolver(&tmp)
	}

	for _, randNumber := range tmp {
		if m.instances[randNumber].IsExpired() {
			if delete(m, m.instances[randNumber].Key) {
				deleted++
			}
		}
	}

	m.ch <- deleted
}
