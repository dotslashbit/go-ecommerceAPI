package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type Server struct {
	Router *httprouter.Router
	DB     *sqlx.DB
	Logger *zap.Logger
	server *http.Server
}

func NewServer(db *sqlx.DB, logger *zap.Logger) *Server {
	router := httprouter.New()

	s := &Server{
		Router: router,
		DB:     db,
		Logger: logger,
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.Router.GET("/health", s.HandleHealth())
}

func (s *Server) HandleHealth() httprouter.Handle {
	type HealthResponse struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.Logger.Info("Health check requested")

		// Check database connection
		err := s.DB.Ping()
		if err != nil {
			s.Logger.Error("Database health check failed", zap.Error(err))
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		response := HealthResponse{
			Status:    "OK",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.Router,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		s.Logger.Info("API listening", zap.String("addr", s.server.Addr))
		serverErrors <- s.server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return err

	case <-shutdown:
		s.Logger.Info("Start shutdown")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := s.server.Shutdown(ctx); err != nil {
			s.Logger.Error("Graceful shutdown did not complete", zap.Error(err))
			if err := s.server.Close(); err != nil {
				s.Logger.Error("Could not stop http server", zap.Error(err))
			}
		}
	}

	return nil
}
