package infrastructure

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// ========================================
// Benchmarks: InMemoryCache (Single Shard)
// ========================================

// BenchmarkInMemoryCache_Get_NoContention teste Get sans contention (single goroutine)
func BenchmarkInMemoryCache_Get_NoContention(b *testing.B) {
	cache := NewInMemoryCache()
	cache.Set("key1", "value1", 5*time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key1")
	}
}

// BenchmarkInMemoryCache_Set_NoContention teste Set sans contention
func BenchmarkInMemoryCache_Set_NoContention(b *testing.B) {
	cache := NewInMemoryCache()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}
}

// BenchmarkInMemoryCache_Get_HighContention teste Get avec haute contention
func BenchmarkInMemoryCache_Get_HighContention(b *testing.B) {
	cache := NewInMemoryCache()
	cache.Set("shared_key", "shared_value", 5*time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("shared_key")
		}
	})
}

// BenchmarkInMemoryCache_Set_HighContention teste Set avec haute contention
func BenchmarkInMemoryCache_Set_HighContention(b *testing.B) {
	cache := NewInMemoryCache()

	b.ResetTimer()
	b.ReportAllocs()

	counter := 0
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			key := counter
			counter++
			mu.Unlock()

			cache.Set(fmt.Sprintf("key%d", key), "value", 5*time.Minute)
		}
	})
}

// BenchmarkInMemoryCache_Mixed_80Read_20Write teste un mix 80% read / 20% write
func BenchmarkInMemoryCache_Mixed_80Read_20Write(b *testing.B) {
	cache := NewInMemoryCache()

	// Pré-remplir le cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	counter := 0
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++

			if localCounter%5 == 0 {
				// 20% writes
				mu.Lock()
				key := counter % 1000
				counter++
				mu.Unlock()

				cache.Set(fmt.Sprintf("key%d", key), "value", 5*time.Minute)
			} else {
				// 80% reads
				_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
			}
		}
	})
}

// ========================================
// Benchmarks: ShardedCache (16 Shards)
// ========================================

// BenchmarkShardedCache_Get_NoContention teste Get sans contention
func BenchmarkShardedCache_Get_NoContention(b *testing.B) {
	cache := NewShardedCache(16)
	cache.Set("key1", "value1", 5*time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("key1")
	}
}

// BenchmarkShardedCache_Set_NoContention teste Set sans contention
func BenchmarkShardedCache_Set_NoContention(b *testing.B) {
	cache := NewShardedCache(16)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}
}

// BenchmarkShardedCache_Get_HighContention teste Get avec haute contention
func BenchmarkShardedCache_Get_HighContention(b *testing.B) {
	cache := NewShardedCache(16)
	cache.Set("shared_key", "shared_value", 5*time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("shared_key")
		}
	})
}

// BenchmarkShardedCache_Set_HighContention teste Set avec haute contention
func BenchmarkShardedCache_Set_HighContention(b *testing.B) {
	cache := NewShardedCache(16)

	b.ResetTimer()
	b.ReportAllocs()

	counter := 0
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			key := counter
			counter++
			mu.Unlock()

			cache.Set(fmt.Sprintf("key%d", key), "value", 5*time.Minute)
		}
	})
}

// BenchmarkShardedCache_Mixed_80Read_20Write teste un mix 80% read / 20% write
func BenchmarkShardedCache_Mixed_80Read_20Write(b *testing.B) {
	cache := NewShardedCache(16)

	// Pré-remplir le cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	counter := 0
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++

			if localCounter%5 == 0 {
				// 20% writes
				mu.Lock()
				key := counter % 1000
				counter++
				mu.Unlock()

				cache.Set(fmt.Sprintf("key%d", key), "value", 5*time.Minute)
			} else {
				// 80% reads
				_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
			}
		}
	})
}

// ========================================
// Benchmarks: Sharding Comparison
// ========================================

// BenchmarkShardedCache_4Shards teste avec 4 shards
func BenchmarkShardedCache_4Shards_Concurrent(b *testing.B) {
	cache := NewShardedCache(4)

	// Pré-remplir
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++
			_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
		}
	})
}

// BenchmarkShardedCache_8Shards teste avec 8 shards
func BenchmarkShardedCache_8Shards_Concurrent(b *testing.B) {
	cache := NewShardedCache(8)

	// Pré-remplir
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++
			_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
		}
	})
}

// BenchmarkShardedCache_16Shards teste avec 16 shards (défaut dans le projet)
func BenchmarkShardedCache_16Shards_Concurrent(b *testing.B) {
	cache := NewShardedCache(16)

	// Pré-remplir
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++
			_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
		}
	})
}

// BenchmarkShardedCache_32Shards teste avec 32 shards
func BenchmarkShardedCache_32Shards_Concurrent(b *testing.B) {
	cache := NewShardedCache(32)

	// Pré-remplir
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++
			_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
		}
	})
}

// ========================================
// Benchmarks: InMemoryCache vs ShardedCache
// ========================================

// BenchmarkComparison_InMemory_vs_Sharded_Sequential compare en séquentiel
func BenchmarkComparison_InMemory_vs_Sharded_Sequential(b *testing.B) {
	b.Run("InMemoryCache", func(b *testing.B) {
		cache := NewInMemoryCache()
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(fmt.Sprintf("key%d", i%100))
		}
	})

	b.Run("ShardedCache_16", func(b *testing.B) {
		cache := NewShardedCache(16)
		for i := 0; i < 100; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(fmt.Sprintf("key%d", i%100))
		}
	})
}

// BenchmarkComparison_InMemory_vs_Sharded_Concurrent compare en concurrence
func BenchmarkComparison_InMemory_vs_Sharded_Concurrent(b *testing.B) {
	b.Run("InMemoryCache", func(b *testing.B) {
		cache := NewInMemoryCache()
		for i := 0; i < 1000; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
		}

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			localCounter := 0
			for pb.Next() {
				localCounter++
				_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
			}
		})
	})

	b.Run("ShardedCache_16", func(b *testing.B) {
		cache := NewShardedCache(16)
		for i := 0; i < 1000; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
		}

		b.ResetTimer()
		b.ReportAllocs()

		b.RunParallel(func(pb *testing.PB) {
			localCounter := 0
			for pb.Next() {
				localCounter++
				_, _ = cache.Get(fmt.Sprintf("key%d", localCounter%1000))
			}
		})
	})
}

// ========================================
// Benchmarks: Hash Function (FNV-1a)
// ========================================

// BenchmarkFNV32_ShortKey teste la performance du hash FNV-1a avec clé courte
func BenchmarkFNV32_ShortKey(b *testing.B) {
	key := "stats:v2:365"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fnv32(key)
	}
}

// BenchmarkFNV32_MediumKey teste avec clé moyenne
func BenchmarkFNV32_MediumKey(b *testing.B) {
	key := "analytics:dashboard:user:12345:period:last_30_days"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fnv32(key)
	}
}

// BenchmarkFNV32_LongKey teste avec clé longue
func BenchmarkFNV32_LongKey(b *testing.B) {
	key := "export:sales:csv:customer:98765:store:downtown:products:laptop,mouse,keyboard:daterange:2024-01-01:2024-12-31:format:compressed"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fnv32(key)
	}
}

// ========================================
// Benchmarks: Cache Operations
// ========================================

// BenchmarkCache_Has teste la méthode Has()
func BenchmarkCache_Has(b *testing.B) {
	cache := NewShardedCache(16)
	cache.Set("existing_key", "value", 5*time.Minute)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cache.Has("existing_key")
	}
}

// BenchmarkCache_Delete teste la méthode Delete()
func BenchmarkCache_Delete(b *testing.B) {
	cache := NewShardedCache(16)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		key := fmt.Sprintf("key%d", i)
		cache.Set(key, "value", 5*time.Minute)
		b.StartTimer()

		cache.Delete(key)
	}
}

// BenchmarkCache_Clear teste la méthode Clear()
func BenchmarkCache_Clear(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cache := NewShardedCache(16)
		for j := 0; j < 1000; j++ {
			cache.Set(fmt.Sprintf("key%d", j), "value", 5*time.Minute)
		}
		b.StartTimer()

		cache.Clear()
	}
}

// ========================================
// Benchmarks: TTL and Expiration
// ========================================

// BenchmarkCache_SetWithTTL teste Set avec différents TTL
func BenchmarkCache_SetWithTTL(b *testing.B) {
	cache := NewShardedCache(16)

	b.Run("1_minute", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 1*time.Minute)
		}
	})

	b.Run("5_minutes", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
		}
	})

	b.Run("1_hour", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			cache.Set(fmt.Sprintf("key%d", i), "value", 1*time.Hour)
		}
	})
}

// BenchmarkCache_GetExpired teste Get avec entrée expirée
func BenchmarkCache_GetExpired(b *testing.B) {
	cache := NewShardedCache(16)

	// Créer des entrées expirées
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("key%d", i), "value", -1*time.Second) // Déjà expiré
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(fmt.Sprintf("key%d", i%100))
	}
}

// ========================================
// Benchmarks: Cache Key Builder
// ========================================

// BenchmarkCacheKeyBuilder_Simple teste la construction de clé simple
func BenchmarkCacheKeyBuilder_Simple(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewCacheKeyBuilder().
			Add("stats").
			Add("v2").
			AddInt(365).
			Build()
	}
}

// BenchmarkCacheKeyBuilder_Complex teste la construction de clé complexe
func BenchmarkCacheKeyBuilder_Complex(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewCacheKeyBuilder().
			Add("export").
			Add("sales").
			Add("csv").
			AddInt(12345).
			Add("store").
			AddInt(67).
			AddInt(2024).
			Build()
	}
}

// BenchmarkCacheKeyBuilder_vs_Sprintf compare le builder avec fmt.Sprintf
func BenchmarkCacheKeyBuilder_vs_Sprintf(b *testing.B) {
	b.Run("CacheKeyBuilder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewCacheKeyBuilder().
				Add("stats").
				Add("v2").
				AddInt(365).
				Build()
		}
	})

	b.Run("fmt_Sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("stats:v2:%d", 365)
		}
	})

	b.Run("string_concat", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = "stats" + ":" + "v2" + ":" + fmt.Sprintf("%d", 365)
		}
	})
}

// ========================================
// Benchmarks: Real-World Scenarios
// ========================================

// BenchmarkCache_RealWorld_StatsService simule l'utilisation dans StatsService
func BenchmarkCache_RealWorld_StatsService(b *testing.B) {
	cache := NewShardedCache(16)

	// Simule plusieurs requêtes avec différentes périodes
	periods := []int{7, 30, 90, 365}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++
			period := periods[localCounter%len(periods)]

			key := NewCacheKeyBuilder().
				Add("stats").
				Add("v2").
				AddInt(period).
				Build()

			// 70% cache hit, 30% cache miss
			if localCounter%10 < 7 {
				// Cache hit simulation: Get + trouve
				if _, found := cache.Get(key); !found {
					// Cache miss: Set
					mockStats := map[string]interface{}{
						"revenue": 125000.50,
						"orders":  1463,
					}
					cache.Set(key, mockStats, 5*time.Minute)
				}
			} else {
				// Cache miss forcé: nouvelle clé
				key := NewCacheKeyBuilder().
					Add("stats").
					Add("v2").
					AddInt(localCounter).
					Build()

				mockStats := map[string]interface{}{
					"revenue": 125000.50,
					"orders":  1463,
				}
				cache.Set(key, mockStats, 5*time.Minute)
			}
		}
	})
}

// BenchmarkCache_RealWorld_HighThroughput simule haute charge (1000 req/s)
func BenchmarkCache_RealWorld_HighThroughput(b *testing.B) {
	cache := NewShardedCache(16)

	// Pré-remplir avec des données typiques
	for i := 0; i < 10000; i++ {
		cache.Set(fmt.Sprintf("key:%d", i), "large_value_data", 5*time.Minute)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		localCounter := 0
		for pb.Next() {
			localCounter++

			// 90% reads, 10% writes (ratio typique)
			if localCounter%10 == 0 {
				cache.Set(fmt.Sprintf("key:%d", localCounter%10000), "new_value", 5*time.Minute)
			} else {
				_, _ = cache.Get(fmt.Sprintf("key:%d", localCounter%10000))
			}
		}
	})
}
