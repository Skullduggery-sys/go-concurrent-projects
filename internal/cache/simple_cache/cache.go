package simple_cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type metrics struct {
	Hits      atomic.Int64 // атомарные счетчики
	Misses    atomic.Int64
	Sets      atomic.Int64
	Evictions atomic.Int64 // сколько ключей удалено из-за TTL
}

type Cache struct {
	mutex     sync.RWMutex
	stopMutex sync.Mutex
	wg        sync.WaitGroup
	isClosed  bool
	data      map[string]interface{}
	ttl       map[string]time.Time
	ticker    *time.Ticker
	stopCh    chan struct{}
	metrics   metrics
}

func New() *Cache {
	cache := &Cache{
		data:   make(map[string]interface{}),
		ttl:    make(map[string]time.Time),
		stopCh: make(chan struct{}),
		ticker: time.NewTicker(time.Second * 5),
	}

	cache.wg.Add(1)
	go cache.cleanUp()

	return cache
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value

	if ttl <= time.Duration(0) {
		c.ttl[key] = time.Time{}
	} else {
		c.ttl[key] = time.Now().Add(ttl)
	}

	c.metrics.Sets.Add(1)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.data[key]
	if !ok {
		c.metrics.Misses.Add(1)

		return nil, false
	}

	isTtlExpired := time.Now().After(c.ttl[key])
	if isTtlExpired {
		c.metrics.Misses.Add(1)

		return nil, false
	}

	c.metrics.Hits.Add(1)

	return value, true
}

func (c *Cache) Stop() {
	c.stopMutex.Lock()
	defer c.stopMutex.Unlock()

	if c.isClosed {
		return
	}

	close(c.stopCh)

	c.wg.Wait()

	c.ticker.Stop()

	c.isClosed = true
}

func (c *Cache) cleanUp() {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopCh:
			return
		case <-c.ticker.C:
			c.cleanUpFunc()
		}
	}
}

func (c *Cache) cleanUpFunc() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	toDelete := make([]string, 0)

	for i := range c.ttl {
		if !c.ttl[i].IsZero() && time.Now().After(c.ttl[i]) {
			toDelete = append(toDelete, i)
		}
	}

	for _, i := range toDelete {
		delete(c.data, i)
		delete(c.ttl, i)
		c.metrics.Evictions.Add(1)
	}
}
