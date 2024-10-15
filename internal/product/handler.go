package product

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/products", h.CreateProduct)
	router.GET("/products/:id", h.GetProduct)
	router.GET("/products", h.ListProducts)
	router.PUT("/products/:id", h.UpdateProduct)
	router.DELETE("/products/:id", h.DeleteProduct)
}
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var input CreateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode create product input", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	product, err := h.service.CreateProduct(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		if err == ErrInvalidInput {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		h.logger.Error("Invalid product ID", zap.Error(err))
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get product", zap.Error(err))
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var filter ProductFilter
	var pagination PaginationParams

	// Parse query parameters
	if err := r.ParseForm(); err != nil {
		h.logger.Error("Failed to parse form data", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Parse filter parameters
	if categoryID := r.Form.Get("category_id"); categoryID != "" {
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err == nil {
			filter.CategoryID = &id
		}
	}
	if minPrice := r.Form.Get("min_price"); minPrice != "" {
		price, err := strconv.ParseFloat(minPrice, 64)
		if err == nil {
			filter.MinPrice = &price
		}
	}
	if maxPrice := r.Form.Get("max_price"); maxPrice != "" {
		price, err := strconv.ParseFloat(maxPrice, 64)
		if err == nil {
			filter.MaxPrice = &price
		}
	}
	if search := r.Form.Get("search"); search != "" {
		filter.Search = &search
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.Form.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.Form.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	pagination = PaginationParams{
		Page:  page,
		Limit: limit,
	}

	products, totalCount, err := h.service.ListProducts(r.Context(), filter, pagination)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Products   []*Product `json:"products"`
		TotalCount int        `json:"total_count"`
		Page       int        `json:"page"`
		Limit      int        `json:"limit"`
	}{
		Products:   products,
		TotalCount: totalCount,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		h.logger.Error("Invalid product ID", zap.Error(err))
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var input UpdateProductInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("Failed to decode update product input", zap.Error(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err = h.service.UpdateProduct(r.Context(), id, input)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Error(err))
		switch err {
		case ErrProductNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		h.logger.Error("Invalid product ID", zap.Error(err))
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to delete product", zap.Error(err))
		if err == ErrProductNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
