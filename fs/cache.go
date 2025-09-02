package fs

import (
	"errors"
	"sync"
	"time"
)

const CACHE_DURATION int64 = 1000 // in ms

type StaticFileCacheItem struct {
	CreatedAt int64 // in ms
	Binary    []byte
}

type StaticFileCache struct {
	Cache map[string]StaticFileCacheItem // Maps a resource location to a cache item
	mu    sync.RWMutex
}

var GlobalStaticFileCache StaticFileCache

func NewStaticFileCache() StaticFileCache {
	return StaticFileCache{
		Cache: make(map[string]StaticFileCacheItem),
	}
}

func (fcache *StaticFileCache) _Set(key string, value []byte) error {
	fcache.mu.Lock()
	defer fcache.mu.Unlock()

	cache_item := StaticFileCacheItem{
		CreatedAt: time.Now().UnixMilli(),
		Binary:    value,
	}

	fcache.Cache[key] = cache_item

	return nil
}

func (fcache *StaticFileCache) _Get(key string) ([]byte, error) {
	fcache.mu.RLock()
	defer fcache.mu.RUnlock()

	cached, ok := fcache.Cache[key]

	if !ok {
		// Cache Miss
		return []byte(""), errors.New("Cache Miss")
	}

	// Check if expired
	now := time.Now().UnixMilli()

	if now >= (cached.CreatedAt + CACHE_DURATION) {
		// Expired

		return []byte(""), errors.New("Expired")
	}

	return cached.Binary, nil
}

func (fcache *StaticFileCache) Get(key string) ([]byte, error) {
	data, err := fcache._Get(key)

	if err != nil {
		// Refresh cache
		body, err := LoadFile(key)

		if err != nil {
			return []byte(""), err
		}

		err = fcache._Set(key, body)

		if err != nil {
			return []byte(""), err
		}

		data = body
	}

	return data, nil
}
