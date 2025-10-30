package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"distributed-saga-coordinator/internal/saga"
)

// HTTPServer exposes saga execution endpoints over HTTP.
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer builds a new HTTP server bound to the provided address.
func NewHTTPServer(addr string, coordinator *saga.Coordinator) *HTTPServer {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /sagas/execute", func(w http.ResponseWriter, r *http.Request) {
		var req saga.ExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		result, err := coordinator.Execute(ctx, req)
		switch {
		case errors.Is(err, saga.ErrSagaNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		case err != nil:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	})

	return &HTTPServer{
		server: &http.Server{
			Addr:              addr,
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           mux,
		},
	}
}

// Run starts the HTTP server. It blocks until the server exits.
func (s *HTTPServer) Run() error {
	return s.server.ListenAndServe()
}

// Shutdown attempts a graceful shutdown within the provided context.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
