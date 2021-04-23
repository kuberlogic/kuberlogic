package controllers

import (
	"k8s.io/apimachinery/pkg/types"
	"sync"
)

var (
	syncMus = sync.Map{}
)

func getMutex(name types.NamespacedName) sync.Mutex {
	v, ok := syncMus.Load(name)
	if !ok {
		v = sync.Mutex{}
		syncMus.Store(name, v)
	}
	return v.(sync.Mutex)
}
