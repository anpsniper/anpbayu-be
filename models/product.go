package models

import (
	"time"
)

// Product represents a product record in the database.
// This struct will be used for storing and retrieving product data.
type Product struct {
	ID          string    `json:"id"`          // Unique identifier for the product
	Name        string    `json:"name"`        // Name of the product
	Description string    `json:"description"` // Description of the product
	Price       float64   `json:"price"`       // Price of the product
	Stock       int       `json:"stock"`       // Current stock quantity
	CreatedAt   time.Time `json:"created_at"`  // Timestamp when the product was created
	UpdatedAt   time.Time `json:"updated_at"`  // Timestamp when the product record was last updated
	// Add other product-related fields as needed (e.g., Category, ImageURL, etc.)
}

// ProductCreateRequest represents the expected payload for creating a new product.
// This would be used in a product creation controller.
type ProductCreateRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

// ProductResponse represents the structure of a product object returned in API responses.
// It typically mirrors the Product struct but can be customized if certain fields
// should be omitted or transformed for public consumption.
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
