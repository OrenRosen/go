package async

import (
	"time"
)

// Async options

type AsyncOption func(*Config)

type Config struct {
	errorHandler        ErrorHandler
	timeoutForGoRoutine time.Duration
	contextPropagators  []ContextPropagator
	maxGoRoutines       uint
	timeoutForGuard     time.Duration
}

func WithTimeoutForGuard(t time.Duration) AsyncOption {
	return func(conf *Config) {
		conf.timeoutForGuard = t
	}
}

func WithTimeoutForGoRoutine(t time.Duration) AsyncOption {
	return func(conf *Config) {
		conf.timeoutForGoRoutine = t
	}
}

func WithErrorHandler(errorHandler ErrorHandler) AsyncOption {
	return func(conf *Config) {
		conf.errorHandler = errorHandler
	}
}

func WithMaxGoRoutines(n uint) AsyncOption {
	return func(conf *Config) {
		conf.maxGoRoutines = n
	}
}

// WithContextPropagation adds a context propagator. A propagator is used to move values from the received context to the new context
// Use this in order to preserve needed values between the contexts when initializing a new goroutine.
func WithContextPropagation(contextPropagator ContextPropagator) AsyncOption {
	return func(conf *Config) {
		conf.contextPropagators = append(conf.contextPropagators, contextPropagator)
	}
}

// pool options

type PoolOption func(*PoolConfig)

type PoolConfig struct {
	errorHandler           ErrorHandler
	timeoutForFN           time.Duration
	timeoutForInsertToPool time.Duration
	contextPropagators     []ContextPropagator
	poolSize               uint
	numberOfWorkers        int
}

// WithPoolTimeoutForFN sets the timeout for running the consumer's function.
func WithPoolTimeoutForFN(t time.Duration) PoolOption {
	return func(conf *PoolConfig) {
		conf.timeoutForFN = t
	}
}

// WithPoolTimeoutInsertToPool sets the timeout for trying to insert the new function into the pool.
// This can happen when the pool is already with max number of messages (pool size is full) and it takes long time for a function to return.
func WithPoolTimeoutInsertToPool(t time.Duration) PoolOption {
	return func(conf *PoolConfig) {
		conf.timeoutForInsertToPool = t
	}
}

func WithPoolErrorHandler(errorHandler ErrorHandler) PoolOption {
	return func(conf *PoolConfig) {
		conf.errorHandler = errorHandler
	}
}

// WithPoolNumberOfWorkers limits the workers number. Each worker is listening to a received data on a different goroutine.
func WithPoolNumberOfWorkers(n int) PoolOption {
	return func(conf *PoolConfig) {
		conf.numberOfWorkers = n
	}
}

// WithPoolSize limits the number of messages the channel can receive
func WithPoolSize(n uint) PoolOption {
	return func(conf *PoolConfig) {
		conf.poolSize = n
	}
}

func WithPoolContextPropagator(contextPropagator ContextPropagator) PoolOption {
	return func(conf *PoolConfig) {
		conf.contextPropagators = append(conf.contextPropagators, contextPropagator)
	}
}
