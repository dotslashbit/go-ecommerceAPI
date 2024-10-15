package product

import (
	"time"

	"github.com/lib/pq"
)

type Product struct {
	ID          int64          `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	Description string         `db:"description" json:"description"`
	Price       float64        `db:"price" json:"price"`
	Categories  pq.StringArray `db:"categories" json:"categories"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

type CreateProductInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Categories  []string `json:"categories"`
}

type UpdateProductInput struct {
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Price       *float64  `json:"price"`
	Categories  *[]string `json:"categories"`
}

type ProductFilter struct {
	CategoryID *int64   `json:"category_id"`
	MinPrice   *float64 `json:"min_price"`
	MaxPrice   *float64 `json:"max_price"`
	Search     *string  `json:"search"`
}

type PaginationParams struct {
	Page  int `json:"page" validate:"required,min=1"`
	Limit int `json:"limit" validate:"required,min=1,max=100"`
}
