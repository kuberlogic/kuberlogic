package cache

import (
	rst "github.com/dgraph-io/ristretto"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"reflect"
	"time"
)

const (
	rstCacheNumCounters      = 1e4
	risrettoCacheCapacity    = 1e7 // can hold up to capacity / cost objects
	risrettoCacheBufferItems = 64
	defaultCost              = 1000
)

type rstCache struct {
	cache *rst.Cache
	log   logging.Logger
}

func (c *rstCache) Set(key, value interface{}, ttlSeconds int) bool {
	cost := valueCost(value)

	c.log.Debugf("setting cache entry with key %s cost %d ttl %d", key, cost, ttlSeconds)
	added := c.cache.SetWithTTL(key, value, cost, time.Duration(ttlSeconds)*time.Second)
	c.log.Debugf("cache entry with key %s is set: %v", key, added)

	return added
}

func (c *rstCache) Get(key interface{}) (interface{}, bool) {
	value, found := c.cache.Get(key)
	c.log.Debugf("found cache entry %s : %v", key, found)
	return value, found
}

func NewRistrettoCache(log logging.Logger) (*rstCache, error) {
	cache, err := rst.NewCache(&rst.Config{
		NumCounters: rstCacheNumCounters,
		MaxCost:     risrettoCacheCapacity,
		BufferItems: risrettoCacheBufferItems,
		Cost:        valueCost,
	})

	rstCache := &rstCache{
		cache: cache,
		log:   log,
	}
	return rstCache, err
}

func valueCost(value interface{}) int64 {
	if reflect.TypeOf(value).Kind() == reflect.String {
		return int64(len(value.(string)))
	}

	return defaultCost
}
