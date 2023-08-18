package async

import (
	"context"
	"fmt"
	"time"
)

const (
	defaultTimeout = time.Second * 5

	defaultNumWorkers = 10
	defaultPoolSize   = 100
)

type funcChannelData struct {
	ctx context.Context
	fn  HandleFunc
}

type Pool struct {
	funcChannel         chan funcChannelData
	errorHandler        ErrorHandler
	timeoutInsertToPool time.Duration
	contextPropagators  []ContextPropagator
}

type HandleFunc func(ctx context.Context) error

// NewPool creates a new Pool instance. The method initializes n number of workers (10 is the default) that listen for a received function.
//
// When calling Pool.RunAsync with HandleFunc, it adds the function to a channel which consumed by the workers.
//
// Options:
//   - WithTimeoutForGoRoutine: the max time to wait for fn to be finished.
//   - WithErrorHandler: add a custom errorHandler that will be triggered in case of an error.
//   - WithContextInjector
//   - WithNumberOfWorkers: The amount of workers.
//   - WithPoolSize: The size of the pool. When calling Pool.Dispatch when the pool is fool, it will wait until the timeout had reached.
//
// Note - can't close the pool.
//
// TODO - add Close functionality.
func NewPool(options ...PoolOption) *Pool {
	conf := PoolConfig{
		errorHandler:           defaultErrorHandler{},
		timeoutForInsertToPool: defaultTimeout,
		timeoutForFN:           defaultTimeoutForGoRoutine,
		numberOfWorkers:        defaultNumWorkers,
		poolSize:               defaultPoolSize,
	}

	for _, op := range options {
		op(&conf)
	}

	p := &Pool{
		funcChannel:         make(chan funcChannelData, conf.poolSize),
		errorHandler:        conf.errorHandler,
		timeoutInsertToPool: conf.timeoutForInsertToPool,
		contextPropagators:  conf.contextPropagators,
	}

	for i := 0; i < conf.numberOfWorkers; i++ {
		w := worker{
			id:           i,
			funcChannel:  p.funcChannel,
			errorHandler: conf.errorHandler,
			timeout:      conf.timeoutForFN,
		}
		w.startReceivingData()
	}

	return p
}

// RunAsync adds the function into the channel which will be received by a worker.
func (p *Pool) RunAsync(ctx context.Context, fn HandleFunc) {
	data := funcChannelData{
		ctx: asyncContext(ctx, p.contextPropagators),
		fn:  fn,
	}

	go func() {
		select {
		case p.funcChannel <- data:
		case <-time.After(p.timeoutInsertToPool):
			err := fmt.Errorf("pool.Dispatch channel is full, timeout waiting for dispatch")
			p.errorHandler.HandleError(ctx, err)
		}
	}()
}

type worker struct {
	id           int
	funcChannel  chan funcChannelData
	errorHandler ErrorHandler
	timeout      time.Duration
}

func (w *worker) startReceivingData() {
	go func() {
		for data := range w.funcChannel {
			w.handleData(data.ctx, data.fn)
		}
	}()
}

func (w *worker) handleData(ctx context.Context, fn HandleFunc) {
	ctx, cacnelFunc := context.WithTimeout(ctx, w.timeout)
	defer cacnelFunc()

	defer recoverPanic(ctx, w.errorHandler)

	if err := fn(ctx); err != nil {
		err = fmt.Errorf("async handleData: %w", err)
		w.errorHandler.HandleError(ctx, err)
	}
}

//
