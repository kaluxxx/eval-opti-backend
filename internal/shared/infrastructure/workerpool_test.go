package infrastructure

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ========================================
// Benchmarks: Worker Pool with Different Worker Counts
// ========================================

// BenchmarkWorkerPool_1Worker teste avec 1 seul worker
func BenchmarkWorkerPool_1Worker_FastTasks(b *testing.B) {
	wp := NewWorkerPool(1)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			// Tâche rapide
			_ = 1 + 1
			return nil
		})
	}
}

// BenchmarkWorkerPool_2Workers teste avec 2 workers
func BenchmarkWorkerPool_2Workers_FastTasks(b *testing.B) {
	wp := NewWorkerPool(2)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			_ = 1 + 1
			return nil
		})
	}
}

// BenchmarkWorkerPool_4Workers teste avec 4 workers (défaut dans le projet)
func BenchmarkWorkerPool_4Workers_FastTasks(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			_ = 1 + 1
			return nil
		})
	}
}

// BenchmarkWorkerPool_8Workers teste avec 8 workers
func BenchmarkWorkerPool_8Workers_FastTasks(b *testing.B) {
	wp := NewWorkerPool(8)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			_ = 1 + 1
			return nil
		})
	}
}

// BenchmarkWorkerPool_16Workers teste avec 16 workers
func BenchmarkWorkerPool_16Workers_FastTasks(b *testing.B) {
	wp := NewWorkerPool(16)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			_ = 1 + 1
			return nil
		})
	}
}

// ========================================
// Benchmarks: Task Duration Variations
// ========================================

// BenchmarkWorkerPool_FastTasks teste avec tâches très rapides (<1µs)
func BenchmarkWorkerPool_FastTasks(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			sum := 0
			for j := 0; j < 10; j++ {
				sum += j
			}
			return nil
		})
	}
}

// BenchmarkWorkerPool_MediumTasks teste avec tâches moyennes (~10µs)
func BenchmarkWorkerPool_MediumTasks(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			sum := 0
			for j := 0; j < 1000; j++ {
				sum += j
			}
			return nil
		})
	}
}

// BenchmarkWorkerPool_SlowTasks teste avec tâches lentes (~100µs)
func BenchmarkWorkerPool_SlowTasks(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(func() error {
			sum := 0
			for j := 0; j < 10000; j++ {
				sum += j
			}
			return nil
		})
	}
}

// ========================================
// Benchmarks: Throughput Measurement
// ========================================

// BenchmarkWorkerPool_Throughput_1000Tasks mesure le throughput avec 1000 tâches
func BenchmarkWorkerPool_Throughput_1000Tasks(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()
		b.StartTimer()

		for j := 0; j < 1000; j++ {
			_ = wp.Submit(func() error {
				sum := 0
				for k := 0; k < 100; k++ {
					sum += k
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// BenchmarkWorkerPool_Throughput_10000Tasks mesure le throughput avec 10000 tâches
func BenchmarkWorkerPool_Throughput_10000Tasks(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()
		b.StartTimer()

		for j := 0; j < 10000; j++ {
			_ = wp.Submit(func() error {
				sum := 0
				for k := 0; k < 100; k++ {
					sum += k
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// ========================================
// Benchmarks: WorkerPool vs Direct Goroutines
// ========================================

// BenchmarkComparison_WorkerPool_vs_Goroutines compare les deux approches
func BenchmarkComparison_WorkerPool_vs_Goroutines(b *testing.B) {
	b.Run("WorkerPool_4Workers", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			b.StopTimer()
			wp := NewWorkerPool(4)
			wp.Start()
			b.StartTimer()

			for j := 0; j < 100; j++ {
				_ = wp.Submit(func() error {
					sum := 0
					for k := 0; k < 100; k++ {
						sum += k
					}
					return nil
				})
			}

			b.StopTimer()
			wp.Wait()
			b.StartTimer()
		}
	})

	b.Run("DirectGoroutines", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup

			for j := 0; j < 100; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					sum := 0
					for k := 0; k < 100; k++ {
						sum += k
					}
				}()
			}

			wg.Wait()
		}
	})

	b.Run("DirectGoroutines_Limited", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var wg sync.WaitGroup
			sem := make(chan struct{}, 4) // Limite à 4 goroutines concurrentes

			for j := 0; j < 100; j++ {
				wg.Add(1)
				sem <- struct{}{}
				go func() {
					defer wg.Done()
					defer func() { <-sem }()

					sum := 0
					for k := 0; k < 100; k++ {
						sum += k
					}
				}()
			}

			wg.Wait()
		}
	})
}

// ========================================
// Benchmarks: Submission Overhead
// ========================================

// BenchmarkWorkerPool_SubmitOnly mesure uniquement l'overhead de Submit()
func BenchmarkWorkerPool_SubmitOnly(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	task := func() error {
		return nil
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wp.Submit(task)
	}
}

// BenchmarkWorkerPool_SubmitAndWait mesure Submit + Wait
func BenchmarkWorkerPool_SubmitAndWait(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()
		b.StartTimer()

		for j := 0; j < 100; j++ {
			_ = wp.Submit(func() error {
				return nil
			})
		}

		wp.Wait()
	}
}

// ========================================
// Benchmarks: Error Handling
// ========================================

// BenchmarkWorkerPool_WithErrors teste la gestion des erreurs
func BenchmarkWorkerPool_WithErrors(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()
		b.StartTimer()

		// 50% des tâches retournent des erreurs
		for j := 0; j < 100; j++ {
			jCopy := j
			_ = wp.Submit(func() error {
				if jCopy%2 == 0 {
					return nil
				}
				return nil // Simulation: pas d'erreur réelle pour le benchmark
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// BenchmarkWorkerPool_ErrorChannel teste la lecture du canal d'erreurs
func BenchmarkWorkerPool_ErrorChannel(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()

		// Soumettre des tâches
		for j := 0; j < 100; j++ {
			_ = wp.Submit(func() error {
				return nil
			})
		}

		wp.Wait()
		b.StartTimer()

		// Lire les erreurs (s'il y en a)
		close(wp.errors)
		for range wp.errors {
			// Consomme les erreurs
		}
	}
}

// ========================================
// Benchmarks: Channel Buffer Size Impact
// ========================================

// BenchmarkWorkerPool_ChannelBuffer_Small teste avec petit buffer
func BenchmarkWorkerPool_ChannelBuffer_Small(b *testing.B) {
	// Note: Le buffer actuel est workerCount*2
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(2) // Petit pool, donc petit buffer (4)
		wp.Start()
		b.StartTimer()

		// Soumettre plus de tâches que la capacité du buffer
		for j := 0; j < 100; j++ {
			_ = wp.Submit(func() error {
				sum := 0
				for k := 0; k < 100; k++ {
					sum += k
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// BenchmarkWorkerPool_ChannelBuffer_Large teste avec grand buffer
func BenchmarkWorkerPool_ChannelBuffer_Large(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(16) // Grand pool, donc grand buffer (32)
		wp.Start()
		b.StartTimer()

		for j := 0; j < 100; j++ {
			_ = wp.Submit(func() error {
				sum := 0
				for k := 0; k < 100; k++ {
					sum += k
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// ========================================
// Benchmarks: Real-World Scenarios
// ========================================

// BenchmarkWorkerPool_RealWorld_DataProcessing simule traitement de données par batch
func BenchmarkWorkerPool_RealWorld_DataProcessing(b *testing.B) {
	// Simule le traitement de 1000 lignes de données en batches de 100
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()

		data := make([]int, 1000)
		for idx := range data {
			data[idx] = idx
		}
		b.StartTimer()

		// Diviser en batches
		batchSize := 100
		for j := 0; j < len(data); j += batchSize {
			end := j + batchSize
			if end > len(data) {
				end = len(data)
			}

			batch := data[j:end]
			_ = wp.Submit(func() error {
				// Traitement du batch
				sum := 0
				for _, val := range batch {
					sum += val * 2
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// BenchmarkWorkerPool_RealWorld_ParquetExport simule l'export Parquet du projet
func BenchmarkWorkerPool_RealWorld_ParquetExport(b *testing.B) {
	// Simule l'export de 3000 lignes avec worker pool (comme export_service_v2.go)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		wp := NewWorkerPool(4)
		wp.Start()

		totalRows := 3000
		batchSize := 1000
		numBatches := (totalRows + batchSize - 1) / batchSize
		b.StartTimer()

		for j := 0; j < numBatches; j++ {
			batchStart := j * batchSize
			batchEnd := batchStart + batchSize
			if batchEnd > totalRows {
				batchEnd = totalRows
			}

			_ = wp.Submit(func() error {
				// Simule construction de strings (comme dans export_service_v2.go ligne 215-229)
				result := ""
				for k := batchStart; k < batchEnd; k++ {
					result += "Order: 1001 | Product: Laptop | Qty: 2 | Amount: 2599.98\n"
				}
				return nil
			})
		}

		b.StopTimer()
		wp.Wait()
		b.StartTimer()
	}
}

// ========================================
// Benchmarks: Concurrent Submission
// ========================================

// BenchmarkWorkerPool_ConcurrentSubmit teste soumission concurrente par plusieurs goroutines
func BenchmarkWorkerPool_ConcurrentSubmit(b *testing.B) {
	wp := NewWorkerPool(4)
	wp.Start()
	defer wp.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = wp.Submit(func() error {
				atomic.AddInt64(&counter, 1)
				return nil
			})
		}
	})

	// Attendre que toutes les tâches se terminent
	time.Sleep(100 * time.Millisecond)
}

// ========================================
// Benchmarks: Start/Stop Overhead
// ========================================

// BenchmarkWorkerPool_StartStop mesure l'overhead de Start et Stop
func BenchmarkWorkerPool_StartStop(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wp := NewWorkerPool(4)
		wp.Start()
		wp.Stop()
	}
}

// BenchmarkWorkerPool_StartWaitStop mesure Start + quelques tâches + Wait
func BenchmarkWorkerPool_StartWaitStop(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wp := NewWorkerPool(4)
		wp.Start()

		for j := 0; j < 10; j++ {
			_ = wp.Submit(func() error {
				return nil
			})
		}

		wp.Wait()
	}
}

// ========================================
// Benchmarks: Scalability
// ========================================

// BenchmarkWorkerPool_Scalability teste la scalabilité avec charge croissante
func BenchmarkWorkerPool_Scalability(b *testing.B) {
	tasks := []int{10, 100, 1000, 10000}

	for _, taskCount := range tasks {
		b.Run(fmt.Sprintf("%d_tasks", taskCount), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				wp := NewWorkerPool(4)
				wp.Start()
				b.StartTimer()

				for j := 0; j < taskCount; j++ {
					_ = wp.Submit(func() error {
						sum := 0
						for k := 0; k < 100; k++ {
							sum += k
						}
						return nil
					})
				}

				b.StopTimer()
				wp.Wait()
				b.StartTimer()
			}
		})
	}
}
