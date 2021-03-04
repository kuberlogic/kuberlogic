package cache

import "github.com/kuberlogic/operator/modules/apiserver/internal/logging"

type Cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key interface{}, val interface{}, ttlSeconds int) bool
}

func NewCache(log logging.Logger) (Cache, error) {
	// only ristretto cache is available
	cache, err := NewRistrettoCache(log)
	return cache, err
}
