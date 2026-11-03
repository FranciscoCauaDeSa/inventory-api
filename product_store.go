package main

import (
	"sort"
	"sync"
)

type ProductStore struct {
	mu       sync.RWMutex
	products map[int64]Product
	nextID   int64
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[int64]Product),
		nextID:   1,
	}
}

func (s *ProductStore) Get(id int64) (Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[id]

	return product, exists
}

func (s *ProductStore) Create(product Product) Product {
	s.mu.Lock()
	defer s.mu.Unlock()

	product.ID = s.nextID
	s.nextID++

	s.products[product.ID] = product

	return product
}

func (s *ProductStore) Update(product Product, id int64) (Product, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return Product{}, false
	}

	product.ID = id
	s.products[id] = product

	return product, true
}

func (s *ProductStore) Delete(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		return false
	}

	delete(s.products, id)
	return true
}

func (s *ProductStore) List() []Product {
	s.mu.RLock()
	defer s.mu.RUnlock()

	products := make([]Product, 0, len(s.products))

	for _, product := range s.products {
		products = append(products, product)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})

	return products
}
