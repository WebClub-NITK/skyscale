package main

import (
	"context"
	"flag"
	"net/http"
	pprof "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bluequbit/faas/control-plane/api"
	"github.com/bluequbit/faas/control-plane/auth"
	"github.com/bluequbit/faas/control-plane/registry"
	"github.com/bluequbit/faas/control-plane/scheduler"
	"github.com/bluequbit/faas/control-plane/state"
	"github.com/bluequbit/faas/control-plane/vm"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func main() {
	// Parse command-line flags
	flag.Parse()

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	logger.Info("Starting Skyscale Control Plane")

	// Check if running in test mode
	if TestMode {
		logger.Info("Running in TEST MODE with simulated host VM")
	}

	// Initialize components
	stateManager, err := state.NewStateManager(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize state manager: %v", err)
	}

	functionRegistry, err := registry.NewFunctionRegistry(stateManager, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize function registry: %v", err)
	}

	vmManager, err := vm.NewVMManager(stateManager, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize VM manager: %v", err)
	}

	// Set up test environment if in test mode
	if err := SetupTestEnvironment(vmManager, logger); err != nil {
		logger.Fatalf("Failed to set up test environment: %v", err)
	}

	functionScheduler, err := scheduler.NewScheduler(vmManager, functionRegistry, stateManager, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize scheduler: %v", err)
	}

	authManager, err := auth.NewAuthManager(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize auth manager: %v", err)
	}

	// Create router
	router := mux.NewRouter()
	AttachProfiler(router)

	// Register API routes
	apiHandler := api.NewAPIHandler(functionRegistry, vmManager, functionScheduler, authManager, stateManager, logger)
	apiHandler.RegisterRoutes(router)

	// Add metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// Add health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add test mode endpoint
	if TestMode {
		router.HandleFunc("/test/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"test_mode": true, "host_vm_id": "` + TestHostVMID + `"}`))
		})
		//add a test/invoke endpoint

	}

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Starting HTTP server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server shutdown failed: %v", err)
	}

	// Cleanup resources
	vmManager.Cleanup()
	stateManager.Close()

	logger.Info("Server stopped")
}
