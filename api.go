package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type API struct {
	products ProductRepository
}

type ProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int64  `json:"stock"`
}

func NewAPI(products ProductRepository) *API {
	return &API{
		products: products,
	}
}

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /products", api.listProducts)
	mux.HandleFunc("GET /products/{id}", api.getProduct)
	mux.HandleFunc("POST /products", api.createProduct)
	mux.HandleFunc("PUT /products/{id}", api.updateProduct)
	mux.HandleFunc("DELETE /products/{id}", api.deleteProduct)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "route not found")
	})
}

func internalServerError(w http.ResponseWriter, operation string, err error) {
	log.Printf("%s: %v", operation, err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func validateProductInput(input *ProductInput) string {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return "product name is required"
	}

	if input.PriceCents < 0 {
		return "price cannot be negative"
	}

	if input.Stock < 0 {
		return "stock cannot be negative"
	}

	return ""
}

func parseProductID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil || id <= 0 {
		return 0, errors.New("invalid product ID")
	}

	return id, nil
}

func (api *API) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := api.products.List(r.Context())
	if err != nil {
		internalServerError(w, "list products", err)
		return
	}

	writeJSON(w, http.StatusOK, products)
}

func (api *API) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, err := api.products.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		internalServerError(w, "get product", err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (api *API) createProduct(w http.ResponseWriter, r *http.Request) {
	var input ProductInput

	if err := decodeJSON(w, r, &input); err != nil {
		decodeError(w, err)
		return
	}

	if message := validateProductInput(&input); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	product, err := api.products.Create(r.Context(), Product{
		Name:        input.Name,
		Description: input.Description,
		PriceCents:  input.PriceCents,
		Stock:       int64(input.Stock),
	})

	if err != nil {
		internalServerError(w, "create product", err)
		return
	}

	w.Header().Set("Location", "/products/"+strconv.FormatInt(product.ID, 10))

	writeJSON(w, http.StatusCreated, product)
}

func (api *API) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input ProductInput

	if err := decodeJSON(w, r, &input); err != nil {
		decodeError(w, err)
		return
	}

	if message := validateProductInput(&input); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	product, err := api.products.Update(r.Context(), id, Product{
		Name:        input.Name,
		Description: input.Description,
		PriceCents:  input.PriceCents,
		Stock:       input.Stock,
	})

	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		internalServerError(w, "update product", err)
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (api *API) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = api.products.Delete(r.Context(), id)

	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		internalServerError(w, "delete product", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
