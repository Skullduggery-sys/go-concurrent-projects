package sharded_cache

import (
	"hash/maphash"
	"time"

	"sharded-cache/internal/cache/simple_cache"
)

type ShardedCache struct {
	shards []*simple_cache.Cache
	seed   maphash.Seed
}

func New(shardsNumber int) *ShardedCache {
	shards := make([]*simple_cache.Cache, shardsNumber)

	for k := range shardsNumber {
		shards[k] = simple_cache.New()
	}

	return &ShardedCache{
		shards: shards,
		seed:   maphash.MakeSeed(),
	}
}

func (s *ShardedCache) Get(key string) (interface{}, bool) {
	shard := maphash.Comparable[string](s.seed, key) % uint64(len(s.shards))

	return s.shards[shard].Get(key)
}

func (s *ShardedCache) Set(k string, v interface{}, ttl time.Duration) {
	shard := maphash.Comparable[string](s.seed, k) % uint64(len(s.shards))

	s.shards[shard].Set(k, v, ttl)
}

func (s *ShardedCache) Stop() {
	for _, v := range s.shards {
		v.Stop()
	}
}
