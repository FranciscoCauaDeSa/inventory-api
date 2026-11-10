package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlc "github.com/FranciscoCauaDeSa/inventory-api/db/sqlc"
)

// Check if PostgresProductRepository implements
// every ProductRepository method
var _ ProductRepository = (*PostgresProductRepository)(nil)

type PostgresProductRepository struct {
	queries *sqlc.Queries
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{
		queries: sqlc.New(pool),
	}
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]Product, error) {
	rows, err := r.queries.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	products := make([]Product, 0)

	for _, row := range rows {
		products = append(products, Product{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			PriceCents:  row.PriceCents,
			Stock:       row.Stock,
		})
	}

	return products, nil
}

func (r *PostgresProductRepository) Get(ctx context.Context, id int64) (Product, error) {
	row, err := r.queries.GetProduct(ctx, id)

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}

	if err != nil {
		return Product{}, err
	}

	return Product{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PriceCents:  row.PriceCents,
		Stock:       row.Stock,
	}, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, product Product) (Product, error) {
	row, err := r.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Stock:       product.Stock,
	})

	if err != nil {
		return Product{}, err
	}

	return Product{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PriceCents:  row.PriceCents,
		Stock:       row.Stock,
	}, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, id int64, product Product) (Product, error) {
	row, err := r.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:          id,
		Name:        product.Name,
		Description: product.Description,
		PriceCents:  product.PriceCents,
		Stock:       product.Stock,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}

	if err != nil {
		return Product{}, err
	}

	return Product{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		PriceCents:  row.PriceCents,
		Stock:       row.Stock,
	}, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id int64) error {
	rowsAffected, err := r.queries.DeleteProduct(ctx, id)

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func toPGText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid: true,
	}
}

func fromPGText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}
