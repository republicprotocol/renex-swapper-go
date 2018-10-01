package swapper

import "sync"

type MutexCache struct {
	status map[[32]byte]bool
	mu     *sync.RWMutex
}

func NewMutexCache() MutexCache {
	return MutexCache{
		status: map[[32]byte]bool{},
		mu:     new(sync.RWMutex),
	}
}

func (cache *MutexCache) Lock(id [32]byte) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.status[id] {
		return false
	}
	cache.status[id] = true
	return true
}

func (cache *MutexCache) Unlock(id [32]byte) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.status[id] = false
}
