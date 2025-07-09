package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	redis    *redis.Client
	memory   *sync.Map
	config   *CacheConfig
	fallback bool
}

type CacheConfig struct {
	RedisURL     string
	DefaultTTL   time.Duration
	MaxMemoryMB  int
	EnableMemory bool
	EnableRedis  bool
}

type CacheItem struct {
	Data      []byte
	ExpiresAt time.Time
}

var (
	cacheService *CacheService
	cacheOnce    sync.Once
)

func InitCacheService(config *CacheConfig) *CacheService {
	cacheOnce.Do(func() {
		cacheService = &CacheService{
			memory:   &sync.Map{},
			config:   config,
			fallback: false,
		}

		// Initialize Redis if enabled
		if config.EnableRedis {
			opt, err := redis.ParseURL(config.RedisURL)
			if err != nil {
				log.Printf("❌ Failed to parse Redis URL: %v", err)
				cacheService.fallback = true
			} else {
				cacheService.redis = redis.NewClient(opt)
				
				// Test Redis connection
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				
				if err := cacheService.redis.Ping(ctx).Err(); err != nil {
					log.Printf("❌ Redis connection failed: %v", err)
					cacheService.fallback = true
				} else {
					log.Printf("✅ Redis cache connected successfully")
				}
			}
		}

		if config.EnableMemory {
			log.Printf("✅ In-memory cache enabled")
		}
	})

	return cacheService
}

func GetCacheService() *CacheService {
	return cacheService
}

func (c *CacheService) Get(key string) ([]byte, error) {
	ctx := context.Background()
	
	// Try Redis first if available
	if c.redis != nil && !c.fallback {
		val, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			return []byte(val), nil
		}
		if err != redis.Nil {
			log.Printf("❌ Redis get error: %v", err)
		}
	}

	// Fallback to in-memory cache
	if c.config.EnableMemory {
		if val, ok := c.memory.Load(key); ok {
			item := val.(CacheItem)
			if time.Now().Before(item.ExpiresAt) {
				return item.Data, nil
			}
			// Remove expired item
			c.memory.Delete(key)
		}
	}

	return nil, fmt.Errorf("cache miss")
}

func (c *CacheService) Set(key string, data []byte, ttl time.Duration) error {
	ctx := context.Background()
	
	// Set in Redis if available
	if c.redis != nil && !c.fallback {
		err := c.redis.Set(ctx, key, data, ttl).Err()
		if err != nil {
			log.Printf("❌ Redis set error: %v", err)
		}
	}

	// Also set in memory cache
	if c.config.EnableMemory {
		item := CacheItem{
			Data:      data,
			ExpiresAt: time.Now().Add(ttl),
		}
		c.memory.Store(key, item)
	}

	return nil
}

func (c *CacheService) Delete(key string) error {
	ctx := context.Background()
	
	// Delete from Redis if available
	if c.redis != nil && !c.fallback {
		err := c.redis.Del(ctx, key).Err()
		if err != nil {
			log.Printf("❌ Redis delete error: %v", err)
		}
	}

	// Delete from memory cache
	if c.config.EnableMemory {
		c.memory.Delete(key)
	}

	return nil
}

func (c *CacheService) DeletePattern(pattern string) error {
	ctx := context.Background()
	
	// Delete from Redis if available
	if c.redis != nil && !c.fallback {
		iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			err := c.redis.Del(ctx, iter.Val()).Err()
			if err != nil {
				log.Printf("❌ Redis delete pattern error: %v", err)
			}
		}
	}

	// Delete from memory cache (simplified pattern matching)
	if c.config.EnableMemory {
		c.memory.Range(func(key, value interface{}) bool {
			keyStr := key.(string)
			// Simple pattern matching (could be enhanced)
			if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
				prefix := pattern[:len(pattern)-1]
				if len(keyStr) >= len(prefix) && keyStr[:len(prefix)] == prefix {
					c.memory.Delete(key)
				}
			}
			return true
		})
	}

	return nil
}

func (c *CacheService) Clear() error {
	ctx := context.Background()
	
	// Clear Redis if available
	if c.redis != nil && !c.fallback {
		err := c.redis.FlushDB(ctx).Err()
		if err != nil {
			log.Printf("❌ Redis clear error: %v", err)
		}
	}

	// Clear memory cache
	if c.config.EnableMemory {
		c.memory.Range(func(key, value interface{}) bool {
			c.memory.Delete(key)
			return true
		})
	}

	return nil
}

func (c *CacheService) GetJSON(key string, dest interface{}) error {
	data, err := c.Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *CacheService) SetJSON(key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Set(key, jsonData, ttl)
}

func (c *CacheService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	if c.redis != nil && !c.fallback {
		ctx := context.Background()
		info, err := c.redis.Info(ctx, "memory").Result()
		if err == nil {
			stats["redis_info"] = info
		}
	}

	// Count memory cache items
	memoryCount := 0
	c.memory.Range(func(key, value interface{}) bool {
		memoryCount++
		return true
	})
	stats["memory_items"] = memoryCount
	stats["fallback_mode"] = c.fallback
	
	return stats
}

func (c *CacheService) HealthCheck() error {
	if c.redis != nil && !c.fallback {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		err := c.redis.Ping(ctx).Err()
		if err != nil {
			return fmt.Errorf("redis health check failed: %v", err)
		}
	}
	return nil
}

func (c *CacheService) Close() error {
	if c.redis != nil {
		return c.redis.Close()
	}
	return nil
}