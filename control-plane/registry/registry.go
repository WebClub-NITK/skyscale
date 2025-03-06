
// Package registry provides functionality for managing function metadata and code.
//
// The FunctionRegistry manages the registration, updating, and retrieval of functions.
// It also handles the storage and retrieval of function code and metadata.
//
// The registry is used by the scheduler to allocate VMs for function execution,

package registry

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/bluequbit/faas/control-plane/state"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FunctionRegistry manages the serverless functions
type FunctionRegistry struct {
	stateManager *state.StateManager
	logger       *logrus.Logger
	storageDir   string
}

// FunctionMetadata contains metadata about a function
type FunctionMetadata struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Runtime   string    `json:"runtime"`
	Memory    int       `json:"memory"`
	Timeout   int       `json:"timeout"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
	Version   string    `json:"version"`
}

// FunctionCode contains the code and requirements for a function
type FunctionCode struct {
	Code         string `json:"code"`
	Requirements string `json:"requirements"`
	Config       string `json:"config"`
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry(stateManager *state.StateManager, logger *logrus.Logger) (*FunctionRegistry, error) {
	// Create storage directory if it doesn't exist
	storageDir := "function-storage"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, err
	}

	return &FunctionRegistry{
		stateManager: stateManager,
		logger:       logger,
		storageDir:   storageDir,
	}, nil
}

// RegisterFunction registers a new function
func (r *FunctionRegistry) RegisterFunction(name, runtime string, memory, timeout int, code, requirements, config string) (*FunctionMetadata, error) {
	// Check if function with the same name already exists
	_, err := r.stateManager.GetFunctionByName(name)
	if err == nil {
		return nil, errors.New("function with this name already exists")
	}

	// Create function ID
	id := uuid.New().String()

	// Create function directory
	functionDir := filepath.Join(r.storageDir, id)
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		return nil, err
	}

	// Write function code
	if err := ioutil.WriteFile(filepath.Join(functionDir, "handler.py"), []byte(code), 0644); err != nil {
		return nil, err
	}

	// Write requirements.txt
	if err := ioutil.WriteFile(filepath.Join(functionDir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		return nil, err
	}

	// Write skyscale.yaml
	if err := ioutil.WriteFile(filepath.Join(functionDir, "skyscale.yaml"), []byte(config), 0644); err != nil {
		return nil, err
	}

	// Create function in state manager
	now := time.Now()
	function := &state.Function{
		ID:        id,
		Name:      name,
		Runtime:   runtime,
		Memory:    memory,
		Timeout:   timeout,
		CreatedAt: now,
		UpdatedAt: now,
		Status:    "ready",
		Version:   "1.0.0",
		Code:      code,
	}

	if err := r.stateManager.SaveFunction(function); err != nil {
		// Cleanup on failure
		os.RemoveAll(functionDir)
		return nil, err
	}

	return &FunctionMetadata{
		ID:        function.ID,
		Name:      function.Name,
		Runtime:   function.Runtime,
		Memory:    function.Memory,
		Timeout:   function.Timeout,
		CreatedAt: function.CreatedAt,
		UpdatedAt: function.UpdatedAt,
		Status:    function.Status,
		Version:   function.Version,
	}, nil
}

// UpdateFunction updates an existing function
func (r *FunctionRegistry) UpdateFunction(id string, code, requirements, config string) (*FunctionMetadata, error) {
	// Get function from state manager
	function, err := r.stateManager.GetFunction(id)
	if err != nil {
		return nil, err
	}

	// Update function directory
	functionDir := filepath.Join(r.storageDir, id)

	// Write function code
	if err := ioutil.WriteFile(filepath.Join(functionDir, "handler.py"), []byte(code), 0644); err != nil {
		return nil, err
	}

	// Write requirements.txt
	if err := ioutil.WriteFile(filepath.Join(functionDir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		return nil, err
	}

	// Write skyscale.yaml
	if err := ioutil.WriteFile(filepath.Join(functionDir, "skyscale.yaml"), []byte(config), 0644); err != nil {
		return nil, err
	}

	// Update function in state manager
	function.UpdatedAt = time.Now()
	function.Code = code
	function.Version = incrementVersion(function.Version)

	if err := r.stateManager.SaveFunction(function); err != nil {
		return nil, err
	}

	return &FunctionMetadata{
		ID:        function.ID,
		Name:      function.Name,
		Runtime:   function.Runtime,
		Memory:    function.Memory,
		Timeout:   function.Timeout,
		CreatedAt: function.CreatedAt,
		UpdatedAt: function.UpdatedAt,
		Status:    function.Status,
		Version:   function.Version,
	}, nil
}

// GetFunction retrieves a function by ID
func (r *FunctionRegistry) GetFunction(id string) (*FunctionMetadata, error) {
	function, err := r.stateManager.GetFunction(id)
	if err != nil {
		return nil, err
	}

	return &FunctionMetadata{
		ID:        function.ID,
		Name:      function.Name,
		Runtime:   function.Runtime,
		Memory:    function.Memory,
		Timeout:   function.Timeout,
		CreatedAt: function.CreatedAt,
		UpdatedAt: function.UpdatedAt,
		Status:    function.Status,
		Version:   function.Version,
	}, nil
}

// GetFunctionByName retrieves a function by name
func (r *FunctionRegistry) GetFunctionByName(name string) (*FunctionMetadata, error) {
	function, err := r.stateManager.GetFunctionByName(name)
	if err != nil {
		return nil, err
	}

	return &FunctionMetadata{
		ID:        function.ID,
		Name:      function.Name,
		Runtime:   function.Runtime,
		Memory:    function.Memory,
		Timeout:   function.Timeout,
		CreatedAt: function.CreatedAt,
		UpdatedAt: function.UpdatedAt,
		Status:    function.Status,
		Version:   function.Version,
	}, nil
}

// GetFunctionCode retrieves the code for a function
func (r *FunctionRegistry) GetFunctionCode(id string) (*FunctionCode, error) {
	// Get function from state manager
	_, err := r.stateManager.GetFunction(id)
	if err != nil {
		return nil, err
	}

	// Read function code
	functionDir := filepath.Join(r.storageDir, id)
	code, err := ioutil.ReadFile(filepath.Join(functionDir, "handler.py"))
	if err != nil {
		return nil, err
	}

	// Read requirements.txt
	requirements, err := ioutil.ReadFile(filepath.Join(functionDir, "requirements.txt"))
	if err != nil {
		return nil, err
	}

	// Read skyscale.yaml
	config, err := ioutil.ReadFile(filepath.Join(functionDir, "skyscale.yaml"))
	if err != nil {
		return nil, err
	}

	return &FunctionCode{
		Code:         string(code),
		Requirements: string(requirements),
		Config:       string(config),
	}, nil
}

// ListFunctions lists all functions
func (r *FunctionRegistry) ListFunctions() ([]FunctionMetadata, error) {
	functions, err := r.stateManager.ListFunctions()
	if err != nil {
		return nil, err
	}

	result := make([]FunctionMetadata, len(functions))
	for i, function := range functions {
		result[i] = FunctionMetadata{
			ID:        function.ID,
			Name:      function.Name,
			Runtime:   function.Runtime,
			Memory:    function.Memory,
			Timeout:   function.Timeout,
			CreatedAt: function.CreatedAt,
			UpdatedAt: function.UpdatedAt,
			Status:    function.Status,
			Version:   function.Version,
		}
	}

	return result, nil
}

// DeleteFunction deletes a function
func (r *FunctionRegistry) DeleteFunction(id string) error {
	// Get function from state manager
	function, err := r.stateManager.GetFunction(id)
	if err != nil {
		return err
	}

	// Delete function directory
	functionDir := filepath.Join(r.storageDir, id)
	if err := os.RemoveAll(functionDir); err != nil {
		return err
	}

	// Delete function from state manager
	return r.stateManager.DeleteFunction(function.ID)
}

// incrementVersion increments the version number
func incrementVersion(version string) string {
	var major, minor, patch int
	fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	patch++
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
