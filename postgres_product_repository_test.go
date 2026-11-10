package main

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupTestRepository(t *testing.T) *PostgresProductRepository {
	t.Helper()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to connect to test database: %v", err)
	}

	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to reset products table: %v", err)
	}

	return NewPostgresProductRepository(pool)
}

func assertProductEqual(t *testing.T, got, want Product) {
	t.Helper()

	if got != want {
		t.Errorf("products don't math:\ngot:  %+v\nwant: %+v", got, want)
	}
}

func TestPostgresProductRepositoryCRUD(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	// Test an empty list.
	products, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List returned an error: %v", err)
	}

	if len(products) != 0 {
		t.Fatalf("expected empty product list, got %d products", len(products))
	}

	// Test Create
	input := Product{
		Name:        "Mechanical Keyboard",
		Description: "RGB keyboard",
		PriceCents:  12999,
		Stock:       15,
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected generated product ID")
	}

	expected := input

	expected.ID = created.ID

	assertProductEqual(t, created, expected)

	// Test Get
	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	assertProductEqual(t, got, expected)

	// Test List.
	products, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List returned an error: %v", err)
	}

	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	// Test Update.
	update := Product{
		Name:        "Mechanical Keyboard Pro",
		Description: "Updated keyboard",
		PriceCents:  14999,
		Stock:       20,
	}

	updated, err := repo.Update(ctx, created.ID, update)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	update.ID = created.ID
	assertProductEqual(t, updated, update)

	// Test Delete.
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}

	// Verify the product no longer exists.
	_, err = repo.Get(ctx, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestPostgresProductRepositoryNotFound(t *testing.T) {
	repo := setupTestRepository(t)
	ctx := context.Background()

	const missingID int64 = 999999

	// Get nonexistent product.
	_, err := repo.Get(ctx, missingID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get: expected ErrNotFound, got %v", err)
	}

	// Update nonexistent product.
	_, err = repo.Update(ctx, missingID, Product{
		Name:       "Does not exist",
		PriceCents: 1000,
		Stock:      1,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Update: expected ErrNotFound, got %v", err)
	}

	// Delete nonexistent product.
	err = repo.Delete(ctx, missingID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete: expected ErrNotFound, got %v", err)
	}
}
