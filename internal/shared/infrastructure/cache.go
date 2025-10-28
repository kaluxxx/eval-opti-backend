package infrastructure

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry représente une entrée de cache avec expiration
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// IsExpired vérifie si l'entrée est expirée
func (e CacheEntry) IsExpired() bool {
	return time.Now().After(e.Expiration)
}

// Cache interface pour l'abstraction du cache
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Has(key string) bool
}

// InMemoryCache implémentation en mémoire du cache avec TTL
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// NewInMemoryCache crée un nouveau cache en mémoire
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		entries: make(map[string]CacheEntry),
	}
	// Lancer le nettoyage périodique
	go cache.cleanupExpired()
	return cache
}

// Get récupère une valeur du cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if entry.IsExpired() {
		return nil, false
	}

	return entry.Value, true
}

// Set ajoute ou met à jour une valeur dans le cache
func (c *InMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Delete supprime une entrée du cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear vide complètement le cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]CacheEntry)
}

// Has vérifie si une clé existe et n'est pas expirée
func (c *InMemoryCache) Has(key string) bool {
	_, exists := c.Get(key)
	return exists
}

// cleanupExpired supprime périodiquement les entrées expirées
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if entry.IsExpired() {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// ShardedCache cache avec sharding pour réduire la contention
type ShardedCache struct {
	shards    []*InMemoryCache
	shardMask uint32
}

// NewShardedCache crée un cache avec sharding
func NewShardedCache(shardCount int) *ShardedCache {
	if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
		panic("shardCount must be a power of 2")
	}

	shards := make([]*InMemoryCache, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = NewInMemoryCache()
	}

	return &ShardedCache{
		shards:    shards,
		shardMask: uint32(shardCount - 1),
	}
}

// getShard retourne le shard approprié pour une clé
func (sc *ShardedCache) getShard(key string) *InMemoryCache {
	hash := fnv32(key)
	return sc.shards[hash&sc.shardMask]
}

// Get récupère une valeur du cache
func (sc *ShardedCache) Get(key string) (interface{}, bool) {
	return sc.getShard(key).Get(key)
}

// Set ajoute ou met à jour une valeur dans le cache
func (sc *ShardedCache) Set(key string, value interface{}, ttl time.Duration) {
	sc.getShard(key).Set(key, value, ttl)
}

// Delete supprime une entrée du cache
func (sc *ShardedCache) Delete(key string) {
	sc.getShard(key).Delete(key)
}

// Clear vide tous les shards
func (sc *ShardedCache) Clear() {
	for _, shard := range sc.shards {
		shard.Clear()
	}
}

// Has vérifie si une clé existe
func (sc *ShardedCache) Has(key string) bool {
	return sc.getShard(key).Has(key)
}

// fnv32 calcule un hash FNV-1a 32-bit pour le sharding
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}

// CacheKeyBuilder aide à construire des clés de cache cohérentes
type CacheKeyBuilder struct {
	parts []string
}

// NewCacheKeyBuilder crée un nouveau builder de clé
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{
		parts: make([]string, 0, 4),
	}
}

// Add ajoute une partie à la clé
func (b *CacheKeyBuilder) Add(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// AddInt ajoute un entier à la clé
func (b *CacheKeyBuilder) AddInt(value int) *CacheKeyBuilder {
	b.parts = append(b.parts, fmt.Sprintf("%d", value))
	return b
}

// Build construit la clé finale
func (b *CacheKeyBuilder) Build() string {
	result := ""
	for i, part := range b.parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}
