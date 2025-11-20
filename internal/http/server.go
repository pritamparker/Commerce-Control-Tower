package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	stor "ecommerce/internal/store"
)

// Server bundles the HTTP router with business logic dependencies.
type Server struct {
	Store *stor.MemoryStore
	mux   chi.Router
}

// NewServer wires up routes and middleware.
func NewServer(store *stor.MemoryStore) *Server {
	s := &Server{Store: store}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		api.Route("/cart/{userID}", func(cart chi.Router) {
			cart.Post("/items", s.handleAddItem)
			cart.Get("/items", s.handleViewCart)
			cart.Post("/checkout", s.handleCheckout)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Post("/discounts/generate", s.handleGenerateDiscount)
			admin.Get("/stats", s.handleStats)
		})
	})

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/*", fileServer)

	s.mux = r
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleAddItem(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var req struct {
		SKU      string  `json:"sku"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if req.SKU == "" || req.Quantity <= 0 || req.Price <= 0 {
		writeError(w, http.StatusBadRequest, "sku, price and quantity are required")
		return
	}
	cart := s.Store.AddItem(userID, stor.CartItem{
		SKU:      req.SKU,
		Name:     req.Name,
		Price:    req.Price,
		Quantity: req.Quantity,
	})
	writeJSON(w, http.StatusCreated, cart)
}

func (s *Server) handleViewCart(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	cart := s.Store.ViewCart(userID)
	writeJSON(w, http.StatusOK, cart)
}

func (s *Server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var req struct {
		DiscountCode string `json:"discountCode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	order, err := s.Store.Checkout(userID, req.DiscountCode)
	if err != nil {
		switch err {
		case stor.ErrCartEmpty:
			writeError(w, http.StatusBadRequest, err.Error())
		case stor.ErrDiscountNotActive, stor.ErrDiscountAlreadyUsed, stor.ErrDiscountMismatch:
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) handleGenerateDiscount(w http.ResponseWriter, r *http.Request) {
	code, err := s.Store.GenerateDiscount()
	if err != nil {
		status := http.StatusBadRequest
		if err == stor.ErrDiscountNotEligible {
			status = http.StatusConflict
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, code)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.Store.Stats()
	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
