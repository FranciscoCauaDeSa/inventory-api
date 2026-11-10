package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeProductRepository keeps handler tests independent of PostgreSQL.
type fakeProductRepository struct {
	products map[int64]Product
	nextID   int64
}

func newFakeProductRepository(products ...Product) *fakeProductRepository {
	repo := &fakeProductRepository{products: make(map[int64]Product), nextID: 1}
	for _, product := range products {
		repo.products[product.ID] = product
		if product.ID >= repo.nextID {
			repo.nextID = product.ID + 1
		}
	}
	return repo
}

func (r *fakeProductRepository) List(context.Context) ([]Product, error) {
	products := make([]Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

func (r *fakeProductRepository) Get(_ context.Context, id int64) (Product, error) {
	product, ok := r.products[id]
	if !ok {
		return Product{}, ErrNotFound
	}
	return product, nil
}

func (r *fakeProductRepository) Create(_ context.Context, product Product) (Product, error) {
	product.ID = r.nextID
	r.nextID++
	r.products[product.ID] = product
	return product, nil
}

func (r *fakeProductRepository) Update(_ context.Context, id int64, product Product) (Product, error) {
	if _, ok := r.products[id]; !ok {
		return Product{}, ErrNotFound
	}
	product.ID = id
	r.products[id] = product
	return product, nil
}

func (r *fakeProductRepository) Delete(_ context.Context, id int64) error {
	if _, ok := r.products[id]; !ok {
		return ErrNotFound
	}
	delete(r.products, id)
	return nil
}

func newTestHandler(repo ProductRepository) http.Handler {
	mux := http.NewServeMux()
	NewAPI(repo).RegisterRoutes(mux)
	return mux
}

func decodeResponseProduct(t *testing.T, response *httptest.ResponseRecorder) Product {
	t.Helper()
	var product Product
	if err := json.NewDecoder(response.Body).Decode(&product); err != nil {
		t.Fatalf("decode product response: %v", err)
	}
	return product
}

func decodeResponseProducts(t *testing.T, response *httptest.ResponseRecorder) []Product {
	t.Helper()
	var products []Product
	if err := json.NewDecoder(response.Body).Decode(&products); err != nil {
		t.Fatalf("decode products response: %v", err)
	}
	return products
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, status int, message string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, status, response.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["error"] != message {
		t.Errorf("error = %q, want %q", body["error"], message)
	}
}

func TestListProducts(t *testing.T) {
	want := []Product{
		{ID: 1, Name: "Keyboard", Description: "Mechanical", PriceCents: 7999, Stock: 12},
		{ID: 2, Name: "Mouse", Description: "Wireless", PriceCents: 2999, Stock: 8},
	}
	response := httptest.NewRecorder()
	newTestHandler(newFakeProductRepository(want...)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	got := decodeResponseProducts(t, response)
	if len(got) != len(want) {
		t.Fatalf("got %d products, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("product[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGetProduct(t *testing.T) {
	want := Product{ID: 7, Name: "Monitor", Description: "4K", PriceCents: 25000, Stock: 3}
	response := httptest.NewRecorder()
	newTestHandler(newFakeProductRepository(want)).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products/7", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := decodeResponseProduct(t, response); got != want {
		t.Errorf("product = %+v, want %+v", got, want)
	}
}

func TestGetMissingProduct(t *testing.T) {
	response := httptest.NewRecorder()
	newTestHandler(newFakeProductRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products/7", nil))
	assertErrorResponse(t, response, http.StatusNotFound, "product not found")
}

func TestCreateProduct(t *testing.T) {
	response := httptest.NewRecorder()
	body := `{"name":"Desk","description":"Standing","price_cents":15000,"stock":4}`
	newTestHandler(newFakeProductRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body)))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	want := Product{ID: 1, Name: "Desk", Description: "Standing", PriceCents: 15000, Stock: 4}
	if got := decodeResponseProduct(t, response); got != want {
		t.Errorf("product = %+v, want %+v", got, want)
	}
}

func TestCreateProductInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid JSON", body: `{"name":`},
		{name: "invalid input", body: `{"name":"  ","price_cents":100,"stock":1}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			newTestHandler(newFakeProductRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(test.body)))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
			var body map[string]string
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body["error"] == "" {
				t.Error("expected JSON error message")
			}
		})
	}
}

func TestUpdateMissingProduct(t *testing.T) {
	response := httptest.NewRecorder()
	body := `{"name":"Desk","description":"Standing","price_cents":15000,"stock":4}`
	newTestHandler(newFakeProductRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/products/7", strings.NewReader(body)))
	assertErrorResponse(t, response, http.StatusNotFound, "product not found")
}

func TestDeleteProduct(t *testing.T) {
	response := httptest.NewRecorder()
	newTestHandler(newFakeProductRepository(Product{ID: 7, Name: "Monitor"})).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/products/7", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusNoContent, response.Body.String())
	}
	if response.Body.Len() != 0 {
		t.Errorf("body = %q, want empty body", response.Body.String())
	}
}

func TestDeleteMissingProduct(t *testing.T) {
	response := httptest.NewRecorder()
	newTestHandler(newFakeProductRepository()).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/products/7", nil))
	assertErrorResponse(t, response, http.StatusNotFound, "product not found")
}
