package gocache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type memCacheTestStruct struct {
	Key   string
	Value string
}

func TestGet(t *testing.T) {
	New(1024 * 1024)

	err := Set("key", 1000*time.Second, &memCacheTestStruct{
		Key:   "key",
		Value: "value",
	})

	assert.Nil(t, err)

	var data memCacheTestStruct
	Get("key", &data)

	assert.NotNil(t, data)
}

func TestResolve(t *testing.T) {
	New(1024 * 1024)

	testClosure := func() (*memCacheTestStruct, error) {
		return Resolve("test", time.Second*1000, func() (*memCacheTestStruct, error) {
			t.Logf("execute resolver")
			return &memCacheTestStruct{
				Key:   "key.",
				Value: "value.",
			}, nil
		})
	}

	// print "execute resolver"
	s, err := testClosure()
	assert.Nil(t, err)

	t.Logf("called resolver! %+v", s)

	// print nothing
	s, err = testClosure()
	assert.Nil(t, err)

	t.Logf("cached value!%+v", s)

	testClosure = func() (*memCacheTestStruct, error) {
		return Resolve("test2", time.Second, func() (*memCacheTestStruct, error) {
			t.Logf("execute resolver")
			return &memCacheTestStruct{
				Key:   "key.",
				Value: "value.",
			}, nil
		})
	}

	// print "execute resolver"
	s, err = testClosure()
	assert.Nil(t, err)

	t.Logf("called resolver! %+v", s)

	// expire cache "test"
	time.Sleep(time.Second)

	// print "execute resolver"
	s, err = testClosure()
	assert.Nil(t, err)
	t.Logf("cached value!%+v", s)

	// Test Result:
	// memcache_test.go:32: execute resolver
	// memcache_test.go:46: called resolver! &{Key:key. Value:value.}
	// memcache_test.go:53: cached value!&{Key:key. Value:value.}
	// memcache_test.go:57: execute resolver
	// memcache_test.go:71: called resolver! &{Key:key. Value:value.}
	// memcache_test.go:57: execute resolver
	// memcache_test.go:81: cached value!&{Key:key. Value:value.}

	t.Log("current cache size:", Size())
}

func TestMaxSize(t *testing.T) {
	tests := []struct {
		name       string
		size       int
		max        int
		isOver     bool
		withStruct bool
	}{
		{
			name:   "10MB - not exceed",
			size:   10 * 1024 * 1024,
			max:    11 * 1024 * 1024,
			isOver: false,
		},
		{
			name:   "10MB - exceed",
			size:   10 * 1024 * 1024,
			max:    10 * 1024 * 1024, // 내부 구조를 위해 사용되는 구조체 떄문에 실제 값은 10MB를 초과하게 된다.
			isOver: true,
		},
		{
			name:       "100MB - not exceed with struct",
			size:       1024,
			max:        101 * 1024 * 1024,
			isOver:     false,
			withStruct: true,
		},
	}

	testClosure := func(count int) ([]byte, error) {
		return Resolve("test", time.Second*1000, func() ([]byte, error) {
			t.Logf("execute resolver")
			tmp := make([]byte, 0, count)
			for i := 0; i < count; i++ {
				tmp = append(tmp, 1)
			}

			return tmp, nil
		})
	}

	withStruct := func(count int) ([]memCacheTestStruct, error) {
		return Resolve("test", time.Second*1000, func() ([]memCacheTestStruct, error) {
			t.Logf("execute resolver")
			tmp := make([]memCacheTestStruct, 0, count)
			for i := 0; i < count; i++ {
				tmp = append(tmp, memCacheTestStruct{
					Key:   "key.",
					Value: "value.",
				})
			}
			return tmp, nil
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New(uint(tt.max))
			if tt.withStruct {
				objList, err := withStruct(tt.size)
				t.Log(err)
				if tt.isOver {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}

				assert.Equal(t, tt.size, len(objList))
			} else {
				b, err := testClosure(tt.size)
				t.Log(err)
				if tt.isOver {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}

				assert.Equal(t, tt.size, len(b))
			}

			t.Log("current cache size:", Size())
		})
	}

}

func TestExpired(t *testing.T) {
	New(0)

	err := Set("test", time.Second, "test string")
	assert.Nil(t, err)
	assert.NotZero(t, Count())

	tt := Value("test")
	t.Log(tt.Value)

	time.Sleep(time.Second * 2)

	t.Log("current cache size:", Size())
	var str string
	Get("test", &str)

	assert.NotEqual(t, "test string", str)
	assert.Zero(t, Count())
}
