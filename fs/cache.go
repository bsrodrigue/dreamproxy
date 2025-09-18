package fs

import (
	"dreamproxy/contentencoder"
	"dreamproxy/format"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const CACHE_DURATION_MS int64 = 5 * 60 * 1000 // in ms (5mins)

var mimeCacheDurationMs = map[string]int64{
	".html":  60 * 1000, // 1min
	".css":   CACHE_DURATION_MS,
	".js":    CACHE_DURATION_MS,
	".json":  CACHE_DURATION_MS,
	".png":   CACHE_DURATION_MS,
	".jpg":   CACHE_DURATION_MS,
	".jpeg":  CACHE_DURATION_MS,
	".gif":   CACHE_DURATION_MS,
	".svg":   CACHE_DURATION_MS,
	".ico":   CACHE_DURATION_MS,
	".txt":   CACHE_DURATION_MS,
	".woff":  CACHE_DURATION_MS,
	".woff2": CACHE_DURATION_MS,
	".ttf":   CACHE_DURATION_MS,
}

var mimeCacheImmutability = map[string]string{
	".html":  "", // 1min
	".css":   "immutable",
	".js":    "immutable",
	".json":  "immutable",
	".png":   "immutable",
	".jpg":   "immutable",
	".jpeg":  "immutable",
	".gif":   "immutable",
	".svg":   "immutable",
	".ico":   "immutable",
	".txt":   "immutable",
	".woff":  "immutable",
	".woff2": "immutable",
	".ttf":   "immutable",
}

func GetMimeCacheDurationMs(mime_type string) int64 {
	duration, ok := mimeCacheDurationMs[mime_type]

	if !ok {
		duration = 0
	}

	return duration
}

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
		return nil, errors.New("Cache Miss")
	}

	now := time.Now().UnixMilli()
	ext := filepath.Ext(key)

	if ext == "" {
		ext = ".html"
	}

	cache_duration_ms := GetMimeCacheDurationMs(ext)

	if now >= (cached.CreatedAt + cache_duration_ms) { // Check if expired
		// Expired
		return nil, errors.New("Expired")
	}

	return cached.Binary, nil
}

func (fcache *StaticFileCache) Get(key string) ([]byte, error) {
	data, err := fcache._Get(key)

	if err != nil {
		// Refresh cache
		body, err := LoadFile(key)

		if err != nil {
			return nil, err
		}

		err = fcache._Set(key, body)

		if err != nil {
			return nil, err
		}

		data = body

		contentencoder.EncodeGzip(data)
	}

	return data, nil
}

func GenerateETag(fileStat os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, fileStat.Size(), fileStat.ModTime().Unix())
}

func GetMimeCacheDurationByKey(key string) int64 {
	ext := filepath.Ext(key)

	if ext == "" {
		ext = ".html"
	}

	cache_duration_ms := GetMimeCacheDurationMs(ext)

	return cache_duration_ms
}

func GetMimeCacheImmutabilityByKey(key string) string {
	ext := filepath.Ext(key)

	if ext == "" {
		ext = ".html"
	}

	immutability := mimeCacheImmutability[ext]

	return immutability
}

func (fcache *StaticFileCache) GetExpirationDateTime(key string) (string, error) {
	cacheDurationMs := GetMimeCacheDurationByKey(key)
	exp := time.Now().Add(time.Duration(cacheDurationMs) * time.Millisecond).UTC()
	return format.TimeToGMT(exp), nil
}

func GenerateCacheControl(key string) string {
	maxAge := GetMimeCacheDurationByKey(key) / 1000
	immutability := GetMimeCacheImmutabilityByKey(key)

	if immutability != "" {
		return fmt.Sprintf("public, max-age=%d, %s", maxAge, immutability)
	}
	return fmt.Sprintf("public, max-age=%d", maxAge)
}
