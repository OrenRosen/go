package async

import (
	"context"
	"fmt"
	"time"
)

type errorTimeout error

type ErrorHandler interface {
	HandleError(ctx context.Context, err error)
}

type ErrorFunc func(ctx context.Context, err error)

func (f ErrorFunc) HandleError(ctx context.Context, err error) {
	f(ctx, err)
}

// ContextPropagator is used for moving values from the ctx into the new context.
// This is in order to preserve needed values between the context when initializing a new go routine.
type ContextPropagator interface {
	MoveToContext(from, to context.Context) context.Context
}

// The ContextPropagatorFunc type is an adapter to allow the use of ordinary functions as context propagators.
// If f is a function with the appropriate signature, ContextPropagatorFunc(f) is a propagator that calls f.
type ContextPropagatorFunc func(from, to context.Context) context.Context

func (f ContextPropagatorFunc) MoveToContext(from, to context.Context) context.Context {
	return f(from, to)
}

const (
	defaultMaxGoRoutines       = 100
	defaultTimeoutForGuard     = time.Second * 5
	defaultTimeoutForGoRoutine = time.Second * 5
)

type Async struct {
	traceServiceName    string
	guard               chan struct{}
	errorHandler        ErrorHandler
	timeoutForGuard     time.Duration
	timeoutForGoRoutine time.Duration
	contextPropagators  []ContextPropagator
}

func New(options ...AsyncOption) *Async {
	conf := Config{
		errorHandler:        defaultErrorHandler{},
		maxGoRoutines:       defaultMaxGoRoutines,
		timeoutForGuard:     defaultTimeoutForGuard,
		timeoutForGoRoutine: defaultTimeoutForGoRoutine,
	}

	for _, op := range options {
		op(&conf)
	}

	return &Async{
		guard:               make(chan struct{}, conf.maxGoRoutines),
		errorHandler:        conf.errorHandler,
		timeoutForGuard:     conf.timeoutForGuard,
		timeoutForGoRoutine: conf.timeoutForGoRoutine,
		contextPropagators:  conf.contextPropagators,
	}
}

func (a *Async) RunAsync(ctx context.Context, fn HandleFunc) {
	ctx = asyncContext(ctx, a.contextPropagators)

	select {
	case a.guard <- struct{}{}:
		go func() {
			ctx, cacnelFunc := context.WithTimeout(ctx, a.timeoutForGoRoutine)

			var err error
			defer func() {
				cacnelFunc()
				<-a.guard
			}()

			defer recoverPanic(ctx, a.errorHandler)

			if err = fn(ctx); err != nil {
				a.errorHandler.HandleError(ctx, fmt.Errorf("async func failed: %w", err))
			}
		}()

	case <-time.After(a.timeoutForGuard):
		a.errorHandler.HandleError(ctx, errorTimeout(fmt.Errorf("async timeout while waiting to guard")))
	}
}

func asyncContext(ctx context.Context, contextPropagators []ContextPropagator) context.Context {
	newCtx := context.Background()

	for _, propagator := range contextPropagators {
		newCtx = propagator.MoveToContext(ctx, newCtx)
	}

	return newCtx
}

func recoverPanic(ctx context.Context, errorHandler ErrorHandler) {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}

		errorHandler.HandleError(ctx, fmt.Errorf("async recoverPanic: %w", err))
	}
}

type defaultErrorHandler struct {
}

func (_ defaultErrorHandler) HandleError(ctx context.Context, err error) {
	fmt.Printf("async error: %+v\n", err)
}

///
