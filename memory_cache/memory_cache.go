package memory_cache

import (
	"time"

	"github.com/maypok86/otter"
)

type MemoryCacheConfig struct {
	CacheCount      int  `yaml:"cache_count"`
	CacheTTLSeconds int  `yaml:"cache_ttl_seconds"`
	CacheEnable     bool `yaml:"cache_enable"`
}

type MemoryCache struct {
	memoryCache       *otter.Cache[string, any]
	memoryCacheConfig MemoryCacheConfig
}

func NewMemoryCache(config *MemoryCacheConfig) *MemoryCache {
	ret := &MemoryCache{}

	ret.memoryCacheConfig = *config

	if !ret.memoryCacheConfig.CacheEnable {
		return ret
	}
	builder := otter.MustBuilder[string, any](ret.memoryCacheConfig.CacheCount).WithTTL(time.Duration(ret.memoryCacheConfig.CacheTTLSeconds) * time.Second)
	cache, err := builder.Build()
	if err != nil {
		panic(err)
	}
	ret.memoryCache = &cache
	return ret
}

func (mc *MemoryCache) Get(key string) (any, bool) {
	if !mc.memoryCacheConfig.CacheEnable {
		return nil, false
	}
	return mc.memoryCache.Get(key)
}

func (mc *MemoryCache) Set(key string, value any) {
	if !mc.memoryCacheConfig.CacheEnable {
		return
	}
	mc.memoryCache.Set(key, value)
}

func (mc *MemoryCache) Delete(key string) {
	if !mc.memoryCacheConfig.CacheEnable {
		return
	}
	mc.memoryCache.Delete(key)
}
