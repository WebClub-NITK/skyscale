package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bluequbit/faas/control-plane/auth"
	"github.com/bluequbit/faas/control-plane/registry"
	"github.com/bluequbit/faas/control-plane/scheduler"
	"github.com/bluequbit/faas/control-plane/vm"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// APIHandler handles API requests
type APIHandler struct {
	functionRegistry *registry.FunctionRegistry
	vmManager        *vm.VMManager
	scheduler        *scheduler.Scheduler
	authManager      *auth.AuthManager
	logger           *logrus.Logger
}

// FunctionRequest represents a request to register a function
type FunctionRequest struct {
	Name         string `json:"name"`
	Runtime      string `json:"runtime"`
	Memory       int    `json:"memory"`
	Timeout      int    `json:"timeout"`
	Code         string `json:"code"`
	Requirements string `json:"requirements"`
	Config       string `json:"config"`
}

// InvokeRequest represents a request to invoke a function
type InvokeRequest struct {
	Input map[string]interface{} `json:"input"`
	Sync  bool                   `json:"sync"`
}

// APIKeyRequest represents a request to generate an API key
type APIKeyRequest struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	ExpiresIn int64    `json:"expires_in"` // in seconds
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(functionRegistry *registry.FunctionRegistry, vmManager *vm.VMManager, scheduler *scheduler.Scheduler, authManager *auth.AuthManager, logger *logrus.Logger) *APIHandler {
	return &APIHandler{
		functionRegistry: functionRegistry,
		vmManager:        vmManager,
		scheduler:        scheduler,
		authManager:      authManager,
		logger:           logger,
	}
}

// RegisterRoutes registers API routes
func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes
	api.HandleFunc("/health", h.healthHandler).Methods("GET")

	// Auth routes
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/api-key", h.generateAPIKeyHandler).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(h.authManager.Middleware)

	// Function routes
	functions := protected.PathPrefix("/functions").Subrouter()
	functions.HandleFunc("", h.listFunctionsHandler).Methods("GET")
	functions.HandleFunc("", h.registerFunctionHandler).Methods("POST")
	functions.HandleFunc("/{id}", h.getFunctionHandler).Methods("GET")
	functions.HandleFunc("/{id}", h.updateFunctionHandler).Methods("PUT")
	functions.HandleFunc("/{id}", h.deleteFunctionHandler).Methods("DELETE")
	functions.HandleFunc("/{id}/invoke", h.invokeFunctionHandler).Methods("POST")
	functions.HandleFunc("/name/{name}", h.getFunctionByNameHandler).Methods("GET")
	functions.HandleFunc("/name/{name}/invoke", h.invokeFunctionByNameHandler).Methods("POST")

	// Execution routes
	executions := protected.PathPrefix("/executions").Subrouter()
	executions.HandleFunc("/{id}", h.getExecutionHandler).Methods("GET")
	executions.HandleFunc("/function/{id}", h.listExecutionsHandler).Methods("GET")

	// VM routes
	vms := protected.PathPrefix("/vms").Subrouter()
	vms.HandleFunc("", h.listVMsHandler).Methods("GET")
	vms.HandleFunc("/{id}", h.getVMHandler).Methods("GET")
}

// healthHandler handles health check requests
func (h *APIHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// generateAPIKeyHandler handles API key generation requests
func (h *APIHandler) generateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	var req APIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate API key
	key, err := h.authManager.GenerateAPIKey(req.UserID, req.Roles, time.Duration(req.ExpiresIn)*time.Second)
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	// Return API key
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"api_key": key,
	})
}

// registerFunctionHandler handles function registration requests
func (h *APIHandler) registerFunctionHandler(w http.ResponseWriter, r *http.Request) {
	var req FunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Register function
	function, err := h.functionRegistry.RegisterFunction(req.Name, req.Runtime, req.Memory, req.Timeout, req.Code, req.Requirements, req.Config)
	if err != nil {
		http.Error(w, "Failed to register function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return function metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// updateFunctionHandler handles function update requests
func (h *APIHandler) updateFunctionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req FunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update function
	function, err := h.functionRegistry.UpdateFunction(id, req.Code, req.Requirements, req.Config)
	if err != nil {
		http.Error(w, "Failed to update function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return function metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// getFunctionHandler handles function retrieval requests
func (h *APIHandler) getFunctionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get function
	function, err := h.functionRegistry.GetFunction(id)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Return function metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// getFunctionByNameHandler handles function retrieval by name requests
func (h *APIHandler) getFunctionByNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Get function
	function, err := h.functionRegistry.GetFunctionByName(name)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	// Return function metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// listFunctionsHandler handles function listing requests
func (h *APIHandler) listFunctionsHandler(w http.ResponseWriter, r *http.Request) {
	// List functions
	functions, err := h.functionRegistry.ListFunctions()
	if err != nil {
		http.Error(w, "Failed to list functions", http.StatusInternalServerError)
		return
	}

	// Return function list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(functions)
}

// deleteFunctionHandler handles function deletion requests
func (h *APIHandler) deleteFunctionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Delete function
	err := h.functionRegistry.DeleteFunction(id)
	if err != nil {
		http.Error(w, "Failed to delete function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Function deleted"))
}

// invokeFunctionHandler handles function invocation requests
func (h *APIHandler) invokeFunctionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Invoke function
	response, err := h.scheduler.ScheduleExecution(id, req.Input, req.Sync)
	if err != nil {
		http.Error(w, "Failed to invoke function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// invokeFunctionByNameHandler handles function invocation by name requests
func (h *APIHandler) invokeFunctionByNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var req InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Invoke function
	response, err := h.scheduler.ScheduleExecutionByName(name, req.Input, req.Sync)
	if err != nil {
		http.Error(w, "Failed to invoke function: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getExecutionHandler handles execution retrieval requests
func (h *APIHandler) getExecutionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get execution
	execution, err := h.scheduler.GetExecution(id)
	if err != nil {
		http.Error(w, "Execution not found", http.StatusNotFound)
		return
	}

	// Return execution
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

// listExecutionsHandler handles execution listing requests
func (h *APIHandler) listExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// List executions
	executions, err := h.scheduler.ListExecutions(id)
	if err != nil {
		http.Error(w, "Failed to list executions", http.StatusInternalServerError)
		return
	}

	// Return execution list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

// listVMsHandler handles VM listing requests
func (h *APIHandler) listVMsHandler(w http.ResponseWriter, r *http.Request) {
	// List VMs
	vms, err := h.vmManager.ListVMs()
	if err != nil {
		http.Error(w, "Failed to list VMs", http.StatusInternalServerError)
		return
	}

	// Return VM list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vms)
}

// getVMHandler handles VM retrieval requests
func (h *APIHandler) getVMHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get VM
	vm, err := h.vmManager.GetVMByID(id)
	if err != nil {
		http.Error(w, "VM not found", http.StatusNotFound)
		return
	}

	// Return VM
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vm)
}
