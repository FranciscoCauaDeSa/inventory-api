-- name: ListProducts :many
SELECT id, name, description, price_cents, stock, sku
FROM products
ORDER BY id;

-- name: GetProduct :one
SELECT id, name, description, price_cents, stock, sku
FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (
    name,
    description,
    price_cents,
    stock,
    sku
)
VALUES ($1, $2, $3, $4, sqlc.narg('sku'))
RETURNING id, name, description, price_cents, stock, sku;

-- name: UpdateProduct :one
UPDATE products
SET
    name = $1,
    description = $2,
    price_cents = $3,
    stock = $4,
    sku = sqlc.narg('sku')
WHERE id = $5
RETURNING id, name, description, price_cents, stock, sku;

-- name: DeleteProduct :execrows
DELETE
FROM products
WHERE id = $1;
