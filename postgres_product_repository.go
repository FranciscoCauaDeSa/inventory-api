package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	db *pgxpool.Pool
}

func NewPostgresProductRepository(db *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (r *PostgresProductRepository) List(ctx context.Context) ([]Product, error) {
	query := `
		SELECT id, name, description, price_cents, stock
		FROM products
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0)

	for rows.Next() {
		var product Product

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.PriceCents,
			&product.Stock,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *PostgresProductRepository) Get(ctx context.Context, id int64) (Product, error) {
	query := `
		SELECT id, name, description, price_cents, stock
		FROM products
		WHERE id = $1
	`

	var product Product

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.Stock,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}

	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func (r *PostgresProductRepository) Create(ctx context.Context, product Product) (Product, error) {
	query := `
		INSERT INTO products (
			name,
			description,
			price_cents,
			stock
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, price_cents, stock
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.Name,
		product.Description,
		product.PriceCents,
		product.Stock,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.Stock,
	)

	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, id int64, product Product) (Product, error) {
	query := `
		UPDATE products
		SET
			name = $1,
			description = $2,
			price_cents = $3,
			stock = $4
		WHERE id = $5
		RETURNING id, name, description, price_cents, stock
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.Name,
		product.Description,
		product.PriceCents,
		product.Stock,
		id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.PriceCents,
		&product.Stock,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}

	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE
		FROM products
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
