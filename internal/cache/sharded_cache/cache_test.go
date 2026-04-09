package sharded_cache

import (
	"hash/maphash"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cache := New(10)
	if cache == nil {
		t.Fatal("cache is nil")
	}
	if len(cache.shards) != 10 {
		t.Fatalf("expected 10 shards, got %d", len(cache.shards))
	}
	defer cache.Stop()
}

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	cache := New(10)
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

	cache := New(10)
	defer cache.Stop()

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Fatal("expected key not to exist")
	}
}

func TestSharding(t *testing.T) {
	t.Parallel()

	cache := New(4)
	defer cache.Stop()

	// Test that same key always goes to same shard
	key := "testkey"
	cache.Set(key, "value1", time.Hour)
	value2, ok := cache.Get(key)

	if !ok {
		t.Fatal("key not found after set")
	}
	// The value should be consistent (same shard)
	if value2 != "value1" {
		t.Fatalf("expected value1, got %v", value2)
	}
}

func TestDifferentKeysSharding(t *testing.T) {
	t.Parallel()

	cache := New(4)
	defer cache.Stop()

	// Set many keys and verify they don't all go to the same shard
	keys := make(map[int]bool)
	for i := 0; i < 100; i++ {
		key := "key" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		cache.Set(key, i, time.Hour)
		shard := int(maphash.Comparable[string](cache.seed, key)) % 4
		keys[shard] = true
	}

	// We should have used multiple shards (statistically very likely)
	if len(keys) < 2 {
		t.Log("Warning: only one shard was used, this is unlikely but possible")
	}
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()

	cache := New(10)
	defer cache.Stop()

	var wg sync.WaitGroup
	iterations := 1000

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key" + string(rune(n%10+'0'))
			cache.Set(key, n, time.Second)
			cache.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestStop(t *testing.T) {
	t.Parallel()

	cache := New(10)
	cache.Stop()

	// After stop, cache should handle gracefully
	cache.Set("key1", "value1", time.Second)
}

func TestStopMultipleTimes(t *testing.T) {
	t.Parallel()

	cache := New(10)

	// First stop
	cache.Stop()

	// Second stop should not panic
	cache.Stop()
}

func TestMultipleKeys(t *testing.T) {
	t.Parallel()

	cache := New(10)
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

func TestShardsNumber(t *testing.T) {
	t.Parallel()

	testCases := []int{1, 2, 5, 10, 100}

	for _, n := range testCases {
		cache := New(n)
		if len(cache.shards) != n {
			t.Fatalf("expected %d shards, got %d", n, len(cache.shards))
		}
		cache.Stop()
	}
}
