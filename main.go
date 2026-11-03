package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	store := NewProductStore()

	api := &API{
		products: store,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /products", api.listProducts)
	mux.HandleFunc("GET /products/{id}", api.getProduct)
	mux.HandleFunc("POST /products", api.createProduct)
	mux.HandleFunc("PUT /products/{id}", api.updateProduct)
	mux.HandleFunc("DELETE /products/{id}", api.deleteProduct)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "route not found")
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")

	log.Fatal(server.ListenAndServe())
}
