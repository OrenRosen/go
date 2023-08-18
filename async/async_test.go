package async_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/OrenRosen/go/async"
)

func Test_async_RunAsync(t *testing.T) {
	rep := &errorHandler{}
	asyncer := async.New(
		async.WithErrorHandler(rep),
		async.WithContextPropagation(propagator{}),
	)

	ctx := context.WithValue(context.Background(), "someKey", "someValue")
	ctx = context.WithValue(ctx, "someOtherKey", "someOtherValue")
	ch := make(chan struct{})
	asyncer.RunAsync(ctx, func(ctx context.Context) error {
		defer func() {
			ch <- struct{}{}
		}()

		val, ok := ctx.Value("someKey").(string)
		require.True(t, ok, "didn't find someKey")
		require.Equal(t, "someValue", val)

		_, ok = ctx.Value("someOtherKey").(string)
		require.False(t, ok, "someOtherKey shouldn't persist between contexts")

		return nil
	})

	select {
	case <-ch:
	case <-time.After(time.Millisecond * 100):
		t.Fatal("timeout waiting for channel")
	}

	require.False(t, rep.called)
}

func Test_async_RunAsync_Panic(t *testing.T) {
	rep := &errorHandler{
		errorCh: make(chan struct{}, 1),
	}
	asyncer := async.New(
		async.WithErrorHandler(rep),
	)

	ctx := context.WithValue(context.Background(), "someKey", "someValue")
	ctx = context.WithValue(ctx, "someOtherKey", "someOtherValue")
	asyncer.RunAsync(ctx, func(ctx context.Context) error {
		panic("aaaa")
	})

	select {
	case <-rep.errorCh:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for errorHandler to be called")
	}

	require.True(t, rep.called, "error errorHandler expected to be called")
}

func Test_async_options(t *testing.T) {
	rep := &errorHandler{
		errorCh: make(chan struct{}),
	}

	// start asyncer with limit 1 in guard
	// short timeout for waiting to guard
	// do twice RunAsync that takes a long time
	// expect that errorHandler will be called
	asyncer := async.New(
		async.WithErrorHandler(rep),
		async.WithMaxGoRoutines(1),
		async.WithTimeoutForGuard(time.Millisecond*10),
		async.WithTimeoutForGoRoutine(time.Second*4),
	)

	asyncer.RunAsync(context.Background(), func(ctx context.Context) error {
		time.Sleep(time.Second * 5)
		return nil
	})

	go func() {
		select {
		case <-rep.errorCh:
		case <-time.After(time.Millisecond * 50):
			t.Error("errorHandler should have been called")
			return
		}
	}()

	asyncer.RunAsync(context.Background(), func(ctx context.Context) error {
		time.Sleep(time.Second * 5)
		return nil
	})
}

type errorHandler struct {
	called  bool
	errorCh chan struct{}
}

func (r *errorHandler) HandleError(ctx context.Context, err error) {
	r.called = true
	if r.errorCh != nil {
		r.errorCh <- struct{}{}
	}
}

type propagator struct {
}

func (i propagator) MoveToContext(from, to context.Context) context.Context {
	val := from.Value("someKey")
	return context.WithValue(to, "someKey", val)
}
