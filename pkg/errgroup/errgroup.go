package errgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type eg = errgroup.Group

type egx struct {
	eg  *eg
	ctx context.Context
}

func New(ctx context.Context) *egx {
	g, ctx := errgroup.WithContext(ctx)
	return &egx{
		eg:  g,
		ctx: ctx,
	}
}

func (e *egx) Wait() error {
	return e.eg.Wait()
}

// Go runs f in a new goroutine.
// If f panics, the panic is wrapped in an error and returned from Go.
func (e *egx) Go(f func(ctx context.Context) error) {
	e.eg.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panicb %v", r)
			}
		}()

		err = f(e.ctx)
		return
	})
}
