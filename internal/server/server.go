// Package server provides shared HTTP setup helpers used by all ZTCV
// binaries (session-svc, identity-svc, risk-chain-svc, chain-adapter).
// It deliberately stays tiny: register your service routes via the
// Builder.Route helpers and call Builder.Run; SIGTERM-graceful shutdown,
// /healthz, /metrics stub, request-ID, CORS, and structured logging are
// all wired automatically.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// Version is the binary version string returned by /healthz. Bumped per
// release; for the MVP scaffold this is fixed at 0.1.0.
const Version = "0.1.0"

// Builder constructs a service binary's HTTP stack. Zero value is not
// usable — use New.
type Builder struct {
	name   string
	port   string
	mux    *http.ServeMux
	server *http.Server

	// metrics — simple per-binary counters exposed by /metrics. The
	// Prometheus-style format is intentional but we don't depend on the
	// Prometheus client library yet.
	requestsTotal atomic.Int64
}

// New creates a Builder for the given service name + resolved port.
// Callers must resolve the port from internal/config — no direct env
// lookup happens here (per WP-01.U4 invariant).
// Health and metrics routes are pre-registered.
func New(name, port string) *Builder {
	b := &Builder{
		name: name,
		port: port,
		mux:  http.NewServeMux(),
	}
	b.mux.HandleFunc("/healthz", b.healthHandler)
	b.mux.HandleFunc("/metrics", b.metricsHandler)
	return b
}

// Route registers an HTTP handler. The handler is wrapped with request-ID,
// CORS, and counter middleware automatically.
func (b *Builder) Route(pattern string, handler http.HandlerFunc) {
	b.mux.HandleFunc(pattern, b.wrap(handler))
}

// RouteRaw is for cases where you've already wrapped the handler (or you
// want to opt out of the standard middleware — rare).
func (b *Builder) RouteRaw(pattern string, handler http.HandlerFunc) {
	b.mux.HandleFunc(pattern, handler)
}

// Run starts the HTTP server and blocks until SIGINT/SIGTERM. It returns
// nil on graceful shutdown, or the underlying server error otherwise.
func (b *Builder) Run() error {
	addr := ":" + b.port
	b.server = &http.Server{
		Addr:              addr,
		Handler:           b.mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("[%s] listening on %s (version=%s)", b.name, addr, Version)
		if err := b.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[%s] server error: %v", b.name, err)
		}
	}()

	// Block until SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("[%s] shutdown signal received, draining…", b.name)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	log.Printf("[%s] shutdown complete", b.name)
	return nil
}

// healthHandler returns the canonical /healthz response.
func (b *Builder) healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"svc":     b.name,
		"version": Version,
	})
}

// metricsHandler returns minimal Prometheus-formatted counters. Real
// Prometheus client integration can land later — this satisfies the
// /metrics endpoint requirement without an extra dependency now.
// Write errors are intentionally swallowed: if the response stream is
// dead, there is nothing useful for the handler to do.
func (b *Builder) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w, "# HELP ztcv_requests_total Total HTTP requests by binary.\n")
	_, _ = fmt.Fprintf(w, "# TYPE ztcv_requests_total counter\n")
	_, _ = fmt.Fprintf(w, "ztcv_requests_total{svc=%q} %d\n", b.name, b.requestsTotal.Load())
}

// wrap layers the standard middleware: request-ID, CORS, counter, logging.
func (b *Builder) wrap(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", reqID)

		// CORS — permissive in dev; tighten per environment later.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-LLM-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		b.requestsTotal.Add(1)
		start := time.Now()
		h(w, r)
		log.Printf("[%s] %s %s req_id=%s dur=%s", b.name, r.Method, r.URL.Path, reqID, time.Since(start))
	}
}

// WriteJSON is a tiny helper shared across binaries.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("WriteJSON encode error: %v", err)
	}
}

// WriteError is a tiny helper for {"error": ..., "reason": ...} responses.
// `reason` is the machine-readable code surfaced to the frontend
// (e.g., "unauthorized_purpose" for Scenario 3).
func WriteError(w http.ResponseWriter, status int, reason, message string) {
	WriteJSON(w, status, map[string]string{
		"error":   http.StatusText(status),
		"reason":  reason,
		"message": message,
	})
}
