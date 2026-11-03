package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type API struct {
	products *ProductStore
}

type ProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Stock       int64  `json:"stock"`
}

func (api *API) RegisterRoutes(*http.ServeMux) {

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
	products := api.products.List()

	writeJSON(w, http.StatusOK, products)
}

func (api *API) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := parseProductID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	product, exists := api.products.Get(id)
	if !exists {
		writeError(w, http.StatusBadRequest, "product not found")
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

	product := api.products.Create(Product{
		Name:        input.Name,
		Description: input.Description,
		PriceCents:  input.PriceCents,
		Stock:       int64(input.Stock),
	})

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

	product, exists := api.products.Update(Product{
		Name:        input.Name,
		Description: input.Description,
		PriceCents:  input.PriceCents,
		Stock:       input.Stock,
	}, id)

	if !exists {
		writeError(w, http.StatusNotFound, "product not found")
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

	deleted := api.products.Delete(id)

	if !deleted {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
