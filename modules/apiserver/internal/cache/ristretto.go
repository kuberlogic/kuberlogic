/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cache

import (
	rst "github.com/dgraph-io/ristretto"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
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

	c.log.Debugw("setting cache entry", "key", key, "cost", cost, "ttl", ttlSeconds)
	added := c.cache.SetWithTTL(key, value, cost, time.Duration(ttlSeconds)*time.Second)
	c.log.Debugw("cache entry with key is set up", "key", key, "entry", added)

	return added
}

func (c *rstCache) Get(key interface{}) (interface{}, bool) {
	value, found := c.cache.Get(key)
	c.log.Debugw("found cache entry", "key", key, "entry", found)
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
