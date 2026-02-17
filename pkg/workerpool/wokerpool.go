package workerpool

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Result interface{}   // equivalente a any do typescript

type Job interface {}

type ProcessorFunc func(ctx context.Context, job Job) Result

type WorkerPool interface {
	Start(ctx context.Context,inputCh <-chan Job) (<-chan Result, error)
	Stop() error
	IsRunning() bool
}

type State int

const (
	StateIdle State = iota   // 0 iota is a keyword in Go which is used to declare an enum
	StateRunning   // 1
	StateStopped   // 2
)

type Config struct {
	WorkerCount int
	Logger *slog.Logger
}

func DefaultConfig() Config {
	return Config{
		WorkerCount: 1,
		Logger: slog.Default(),
	}
}

type workerPool struct {
	workerCount int
	logger *slog.Logger
	state State
	processorFunc ProcessorFunc
	stateMutex sync.Mutex
	stopCh chan struct{}
	stopWg sync.WaitGroup
}

func NewWorkerPool(config Config, processorFunc ProcessorFunc) *workerPool {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 1
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	return &workerPool{
		workerCount: config.WorkerCount,
		logger: config.Logger,
		state: StateIdle,
		processorFunc: processorFunc,
		stopCh: make(chan struct{}),
	}

}

func (wp *workerPool) Start(ctx context.Context, inputCh <-chan Job) (<-chan Result, error) {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()
	if wp.state != StateIdle {
		return nil, fmt.Errorf("worker pool is not idle")
	}
	result := make(chan Result)
	wp.state = StateRunning
	wp.stopCh = make(chan struct{})
	wp.stopWg.Add(wp.workerCount)

	for i:= range wp.workerCount {
		go wp.worker(ctx, i, inputCh, result)
	}

	go func() {
		wp.stopWg.Wait()
		close(result)
		wp.stateMutex.Lock()
		wp.state = StateIdle
		wp.stateMutex.Unlock()
	}()

	return result, nil
}

func (wp *workerPool) Stop() error {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()
	if wp.state != StateRunning {
		return fmt.Errorf("worker pool is not running")
	}
	wp.state = StateStopped
	close(wp.stopCh)
	wp.stopWg.Wait()
	wp.state  = StateIdle
	return nil
}

func (wp *workerPool) IsRunning() bool {
	wp.stateMutex.Lock()
	defer wp.stateMutex.Unlock()
	return wp.state == StateRunning
}

func (wp *workerPool) worker(ctx context.Context, workerID int, inputCh <-chan Job, resultCh chan<- Result) {
	wp.logger.Info("Worker %d started","workerId", workerID)

	for{
		select { // select consegue monitorar mais de um canal ao mesmo tempo	
			case <-wp.stopCh:
				wp.logger.Info("Worker interronpid","workerid", workerID)
			case <-ctx.Done():
				wp.logger.Info("Worker cancelado, interrpendo worker ","workerId", workerID)
		    case job, ok := <-inputCh:
				if !ok {
					wp.logger.Info("Worker %d finalizado","workerid", workerID) 
					return
				}
				result := wp.processorFunc(ctx, job)
				select {
					case resultCh <- result:
					case <-wp.stopCh:
						wp.logger.Info("Worker interrompido", "workerId", workerID)
						return
					case <-ctx.Done():
						wp.logger.Info("Worker cancelado, interrompendo worker", "workerId", workerID)
						return	
				}
		}
	}
}