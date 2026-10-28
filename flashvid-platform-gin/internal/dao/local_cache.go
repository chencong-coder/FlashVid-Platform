package dao

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

var LocalCache *cache.Cache

// MustInitLocalCache 初始化本地缓存
func MustInitLocalCache(cfg *viper.Viper) {
	if !cfg.GetBool("cache.local_enabled") {
		LocalCache = nil
		return
	}

	ttl := cfg.GetDuration("cache.local_ttl")
	if ttl == 0 {
		ttl = 5 * time.Second // 默认 5 秒
	}

	cleanupInterval := ttl * 2
	LocalCache = cache.New(ttl, cleanupInterval)
}

// LocalCacheGet 从本地缓存获取数据（自动降级）
func LocalCacheGet(key string) (interface{}, bool) {
	if LocalCache == nil {
		return nil, false
	}
	return LocalCache.Get(key)
}

// LocalCacheSet 写入本地缓存（自动降级）
func LocalCacheSet(key string, value interface{}, ttl time.Duration) {
	if LocalCache == nil {
		return
	}
	LocalCache.Set(key, value, ttl)
}

// LocalCacheDelete 删除本地缓存
func LocalCacheDelete(key string) {
	if LocalCache == nil {
		return
	}
	LocalCache.Delete(key)
}
