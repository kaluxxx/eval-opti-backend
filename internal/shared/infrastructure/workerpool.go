package infrastructure

import (
	"context"
	"sync"
)

// Task représente une tâche à exécuter
type Task func() error

// WorkerPool gère un pool de workers pour traiter des tâches en parallèle
type WorkerPool struct {
	workerCount int
	tasks       chan Task
	errors      chan error
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool crée un nouveau pool de workers
func NewWorkerPool(workerCount int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount: workerCount,
		tasks:       make(chan Task, workerCount*2),
		errors:      make(chan error, workerCount),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start démarre les workers
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker est la routine d'exécution des tâches
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			if err := task(); err != nil {
				select {
				case wp.errors <- err:
				default:
					// Canal d'erreurs plein, on ignore
				}
			}
		}
	}
}

// Submit soumet une tâche au pool
func (wp *WorkerPool) Submit(task Task) {
	select {
	case <-wp.ctx.Done():
		return
	case wp.tasks <- task:
	}
}

// Wait attend que toutes les tâches soient terminées
func (wp *WorkerPool) Wait() {
	close(wp.tasks)
	wp.wg.Wait()
}

// Stop arrête le pool
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
}

// Errors retourne le canal d'erreurs
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errors
}

// BatchProcessor traite des items par batch avec un worker pool
type BatchProcessor struct {
	batchSize int
	pool      *WorkerPool
}

// NewBatchProcessor crée un nouveau processeur par batch
func NewBatchProcessor(workerCount, batchSize int) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		pool:      NewWorkerPool(workerCount),
	}
}

// Process traite une slice d'items par batch
func (bp *BatchProcessor) Process(items []interface{}, processor func(batch []interface{}) error) error {
	bp.pool.Start()
	defer bp.pool.Wait()

	// Découper en batches et soumettre
	for i := 0; i < len(items); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		bp.pool.Submit(func() error {
			return processor(batch)
		})
	}

	// Collecter les erreurs
	bp.pool.Wait()
	close(bp.pool.errors)

	var firstError error
	for err := range bp.pool.errors {
		if firstError == nil {
			firstError = err
		}
	}

	return firstError
}

// Stop arrête le processeur
func (bp *BatchProcessor) Stop() {
	bp.pool.Stop()
}

// ObjectPool pool générique pour la réutilisation d'objets
type ObjectPool struct {
	pool sync.Pool
	new  func() interface{}
}

// NewObjectPool crée un nouveau pool d'objets
func NewObjectPool(newFunc func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: newFunc,
		},
		new: newFunc,
	}
}

// Get récupère un objet du pool
func (op *ObjectPool) Get() interface{} {
	return op.pool.Get()
}

// Put remet un objet dans le pool
func (op *ObjectPool) Put(obj interface{}) {
	op.pool.Put(obj)
}

// SlicePool pool spécialisé pour les slices
type SlicePool struct {
	pool     *ObjectPool
	capacity int
}

// NewSlicePool crée un pool de slices avec une capacité fixe
func NewSlicePool(capacity int) *SlicePool {
	return &SlicePool{
		pool: NewObjectPool(func() interface{} {
			return make([]interface{}, 0, capacity)
		}),
		capacity: capacity,
	}
}

// Get récupère une slice du pool
func (sp *SlicePool) Get() []interface{} {
	slice := sp.pool.Get().([]interface{})
	return slice[:0] // Reset length mais garde capacity
}

// Put remet une slice dans le pool
func (sp *SlicePool) Put(slice []interface{}) {
	if cap(slice) == sp.capacity {
		sp.pool.Put(slice)
	}
}
