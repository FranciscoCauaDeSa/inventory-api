package main

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("product not found")

type ProductRepository interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id int64) (Product, error)
	Create(ctx context.Context, product Product) (Product, error)
	Update(ctx context.Context, id int64, product Product) (Product, error)
	Delete(ctx context.Context, id int64) error
}
