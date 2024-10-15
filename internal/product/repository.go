package product

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Repository defines the interface for product data operations
type Repository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id int64) (*Product, error)
	List(ctx context.Context, filter ProductFilter, pagination PaginationParams) ([]*Product, int, error)
	Update(ctx context.Context, id int64, input UpdateProductInput) error
	Delete(ctx context.Context, id int64) error
}

// repository is the SQL implementation of the Repository interface
type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new instance of the SQL repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// Create adds a new product to the database
func (r *repository) Create(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (name, description, price, categories)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		product.Name, product.Description, product.Price, product.Categories).
		StructScan(product)

	if err != nil {
		return fmt.Errorf("error creating product: %w", err)
	}

	return nil
}

// GetByID retrieves a single product by its ID
func (r *repository) GetByID(ctx context.Context, id int64) (*Product, error) {
	var product Product
	query := `SELECT * FROM products WHERE id = $1`
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("error getting product: %w", err)
	}
	return &product, nil
}

// List retrieves a list of products, applying filters and pagination
func (r *repository) List(ctx context.Context, filter ProductFilter, pagination PaginationParams) ([]*Product, int, error) {
	query := `SELECT * FROM products`
	countQuery := `SELECT COUNT(*) FROM products`
	whereClause := []string{}
	args := []interface{}{}
	argID := 1

	if filter.CategoryID != nil && *filter.CategoryID != "" {
		whereClause = append(whereClause, fmt.Sprintf(`EXISTS (SELECT 1 FROM unnest(categories) category WHERE category ILIKE $%d)`, argID))
		args = append(args, "%"+*filter.CategoryID+"%")
		argID++
	}
	if filter.MinPrice != nil {
		whereClause = append(whereClause, fmt.Sprintf("price >= $%d", argID))
		args = append(args, *filter.MinPrice)
		argID++
	}
	if filter.MaxPrice != nil {
		whereClause = append(whereClause, fmt.Sprintf("price <= $%d", argID))
		args = append(args, *filter.MaxPrice)
		argID++
	}
	if filter.Search != nil && *filter.Search != "" {
		whereClause = append(whereClause, fmt.Sprintf("(to_tsvector('english', name) @@ plainto_tsquery('english', $%d) OR to_tsvector('english', description) @@ plainto_tsquery('english', $%d))", argID, argID))
		args = append(args, *filter.Search)
		argID++
	}

	if len(whereClause) > 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
		countQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, pagination.Limit, (pagination.Page-1)*pagination.Limit)

	var products []*Product
	err := r.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing products: %w", err)
	}

	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	return products, totalCount, nil
}

// Update modifies an existing product
func (r *repository) Update(ctx context.Context, id int64, input UpdateProductInput) error {
	query := `UPDATE products SET `
	args := []interface{}{}
	argID := 1

	if input.Name != nil {
		query += fmt.Sprintf("name = $%d, ", argID)
		args = append(args, *input.Name)
		argID++
	}
	if input.Description != nil {
		query += fmt.Sprintf("description = $%d, ", argID)
		args = append(args, *input.Description)
		argID++
	}
	if input.Price != nil {
		query += fmt.Sprintf("price = $%d, ", argID)
		args = append(args, *input.Price)
		argID++
	}
	if input.Categories != nil {
		query += fmt.Sprintf("categories = $%d, ", argID)
		args = append(args, *input.Categories)
		argID++
	}

	query = strings.TrimSuffix(query, ", ")
	query += fmt.Sprintf(", updated_at = NOW() WHERE id = $%d", argID)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Delete removes a product from the database
func (r *repository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
