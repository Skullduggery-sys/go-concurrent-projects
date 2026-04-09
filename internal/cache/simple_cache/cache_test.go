package simple_cache

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cache := New()
	if cache == nil {
		t.Fatal("cache is nil")
	}
	if cache.data == nil {
		t.Fatal("data map is nil")
	}
	if cache.ttl == nil {
		t.Fatal("ttl map is nil")
	}
	defer cache.Stop()
}

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	cache.Set("key1", "value1", time.Second*10)
	value, ok := cache.Get("key1")

	if !ok {
		t.Fatal("key not found")
	}
	if value != "value1" {
		t.Fatalf("expected value1, got %v", value)
	}
}

func TestGetNotExists(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Fatal("expected key not to exist")
	}
}

func TestSetWithZeroTTL(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	cache.Set("key1", "value1", 0)
	// Zero/negative TTL is treated as already expired
	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("key with zero TTL should be treated as expired")
	}
}

func TestSetWithNegativeTTL(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	cache.Set("key1", "value1", -time.Second)
	// Zero/negative TTL is treated as already expired
	_, ok := cache.Get("key1")
	if ok {
		t.Fatal("key with negative TTL should be treated as expired")
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	var wg sync.WaitGroup
	iterations := 1000

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set("key", n, time.Second)
			cache.Get("key")
		}(i)
	}

	wg.Wait()
}

func TestStop(t *testing.T) {
	t.Parallel()

	cache := New()
	cache.Stop()

	// After stop, cache should not panic on Set/Get operations
	cache.Set("key1", "value1", time.Second)
	cache.Get("key1")
}

func TestStopMultipleTimes(t *testing.T) {
	t.Parallel()

	cache := New()

	// First stop
	cache.Stop()

	// Second stop should not panic
	cache.Stop()
}

func TestMultipleKeys(t *testing.T) {
	t.Parallel()

	cache := New()
	defer cache.Stop()

	keys := []string{"a", "b", "c", "d", "e"}
	values := []interface{}{1, 2.5, "string", []int{1, 2, 3}, true}

	for i, key := range keys {
		cache.Set(key, values[i], time.Hour)
	}

	for i, key := range keys {
		value, ok := cache.Get(key)
		if !ok {
			t.Fatalf("key %s not found", key)
		}
		if !reflect.DeepEqual(value, values[i]) {
			t.Fatalf("expected %v, got %v", values[i], value)
		}
	}
}
