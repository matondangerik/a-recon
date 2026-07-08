package errgroup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitReturnsNilWhenAllGoroutinesSucceed(t *testing.T) {
	g := New(context.Background())

	done := make(chan struct{}, 2)
	g.Go(func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	})

	err := g.Wait()

	assert.NoError(t, err)
	assert.Len(t, done, 2)
}

func TestWaitReturnsGoroutineError(t *testing.T) {
	g := New(context.Background())
	expectedErr := errors.New("boom")

	g.Go(func(ctx context.Context) error {
		return expectedErr
	})

	err := g.Wait()

	assert.ErrorIs(t, err, expectedErr)
}

func TestGoWrapsPanicAsError(t *testing.T) {
	g := New(context.Background())

	g.Go(func(ctx context.Context) error {
		panic("unexpected")
	})

	err := g.Wait()

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "panicb unexpected")
	}
}

func TestGoPassesDerivedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	g := New(ctx)

	g.Go(func(ctx context.Context) error {
		cancel()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			t.Fatal("expected derived context cancellation")
			return nil
		}
	})

	err := g.Wait()

	assert.ErrorIs(t, err, context.Canceled)
}
