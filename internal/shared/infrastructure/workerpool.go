package infrastructure

import (
	"context"
	"fmt"
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

// Start démarre les workers
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// Submit soumet une tâche au pool
func (wp *WorkerPool) Submit(task Task) error {
	select {
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	case wp.tasks <- task:
		return nil
	}
}

// Wait attend que toutes les tâches soient terminées et ferme le canal de tâches
func (wp *WorkerPool) Wait() {
	close(wp.tasks)
	wp.wg.Wait()
}

// Stop arrête le pool immédiatement
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
}

// Errors retourne le canal d'erreurs
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errors
}
