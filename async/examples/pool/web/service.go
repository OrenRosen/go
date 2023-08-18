package main

import (
	"context"
	"fmt"

	"github.com/OrenRosen/go/async"
)

type service struct {
	pool *async.Pool
}

func NewService() *service {
	return &service{
		pool: async.NewPool(),
	}
}

func (s *service) DoWorkAsync(ctx context.Context, i int) {
	s.pool.RunAsync(ctx, func(ctx context.Context) error {
		fmt.Printf("service is working on data %d\n", i)
		return nil
	})
}
