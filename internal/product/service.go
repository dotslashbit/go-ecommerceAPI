package product

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidInput    = errors.New("invalid input")
)

type Service interface {
	CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error)
	GetProductByID(ctx context.Context, id int64) (*Product, error)
	ListProducts(ctx context.Context, filter ProductFilter, pagination PaginationParams) ([]*Product, int, error)
	UpdateProduct(ctx context.Context, id int64, input UpdateProductInput) error
	DeleteProduct(ctx context.Context, id int64) error
}

type service struct {
	repo      Repository
	validator *validator.Validate
}

func NewService(repo Repository) Service {
	return &service{
		repo:      repo,
		validator: validator.New(),
	}
}

func (s *service) CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, ErrInvalidInput
	}

	product := &Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Categories:  input.Categories,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *service) GetProductByID(ctx context.Context, id int64) (*Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

func (s *service) ListProducts(ctx context.Context, filter ProductFilter, pagination PaginationParams) ([]*Product, int, error) {
	if err := s.validator.Struct(pagination); err != nil {
		return nil, 0, ErrInvalidInput
	}

	return s.repo.List(ctx, filter, pagination)
}

func (s *service) UpdateProduct(ctx context.Context, id int64, input UpdateProductInput) error {
	if err := s.validator.Struct(input); err != nil {
		return ErrInvalidInput
	}

	err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProductNotFound
		}
		return err
	}

	return nil
}

func (s *service) DeleteProduct(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProductNotFound
		}
		return err
	}

	return nil
}
