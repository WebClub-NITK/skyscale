// Package state provides functionality for managing the state of the control plane.
//
// The StateManager manages the database and cache for the control plane.
// It provides methods for saving, retrieving, and deleting functions, executions, and VMs.
// It also provides a method for tracking active executions.

package state

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// StateManager handles the state management for the control plane
type StateManager struct {
	db          *gorm.DB
	cache       *redis.Client
	logger      *logrus.Logger
	activeExecs sync.Map // Map to track active executions
	mu          sync.Mutex
}

// Function represents a serverless function
type Function struct {
	ID        string `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	Runtime   string
	Memory    int
	Timeout   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Status    string
	Version   string
	Code      string
}

// Execution represents a function execution
type Execution struct {
	ID         string `gorm:"primaryKey"`
	FunctionID string
	Status     string
	StartTime  time.Time
	EndTime    time.Time
	Duration   int64
	VMID       string
	Logs       string
	Error      string
}

// VM represents a Firecracker micro-VM
type VM struct {
	ID        string `gorm:"primaryKey"`
	Status    string
	IP        string
	CreatedAt time.Time
	LastUsed  time.Time
	Memory    int
	CPU       int
	IsWarm    bool
}

// NewStateManager creates a new state manager
func NewStateManager(logger *logrus.Logger) (*StateManager, error) {
	// Initialize SQLite database
	db, err := gorm.Open(sqlite.Open("skyscale.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&Function{}, &Execution{}, &VM{})
	if err != nil {
		return nil, err
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test Redis connection
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		logger.Warnf("Redis not available, continuing without cache: %v", err)
		rdb = nil
	}

	return &StateManager{
		db:     db,
		cache:  rdb,
		logger: logger,
	}, nil
}

// SaveFunction saves a function to the database
func (s *StateManager) SaveFunction(function *Function) error {
	return s.db.Save(function).Error
}

// GetFunction retrieves a function by ID
func (s *StateManager) GetFunction(id string) (*Function, error) {
	var function Function
	err := s.db.First(&function, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &function, nil
}

// GetFunctionByName retrieves a function by name
func (s *StateManager) GetFunctionByName(name string) (*Function, error) {
	var function Function
	err := s.db.First(&function, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &function, nil
}

// ListFunctions retrieves all functions
func (s *StateManager) ListFunctions() ([]Function, error) {
	var functions []Function
	err := s.db.Find(&functions).Error
	return functions, err
}

// DeleteFunction deletes a function by ID
func (s *StateManager) DeleteFunction(id string) error {
	return s.db.Delete(&Function{}, "id = ?", id).Error
}

// SaveExecution saves an execution to the database
func (s *StateManager) SaveExecution(execution *Execution) error {
	return s.db.Save(execution).Error
}

// GetExecution retrieves an execution by ID
func (s *StateManager) GetExecution(id string) (*Execution, error) {
	var execution Execution
	err := s.db.First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// ListExecutions retrieves all executions for a function
func (s *StateManager) ListExecutions(functionID string) ([]Execution, error) {
	var executions []Execution
	err := s.db.Find(&executions, "function_id = ?", functionID).Error
	return executions, err
}

// SaveVM saves a VM to the database
func (s *StateManager) SaveVM(vm *VM) error {
	return s.db.Save(vm).Error
}

// GetVM retrieves a VM by ID
func (s *StateManager) GetVM(id string) (*VM, error) {
	var vm VM
	err := s.db.First(&vm, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

// ListVMs retrieves all VMs
func (s *StateManager) ListVMs() ([]VM, error) {
	var vms []VM
	err := s.db.Find(&vms).Error
	return vms, err
}

// ListWarmVMs retrieves all warm VMs
func (s *StateManager) ListWarmVMs() ([]VM, error) {
	var vms []VM
	err := s.db.Find(&vms, "is_warm = ? AND status = ?", true, "ready").Error
	return vms, err
}

// DeleteVM deletes a VM by ID
func (s *StateManager) DeleteVM(id string) error {
	return s.db.Delete(&VM{}, "id = ?", id).Error
}

// TrackActiveExecution adds an execution to the active executions map
func (s *StateManager) TrackActiveExecution(executionID string, vmID string) {
	s.activeExecs.Store(executionID, vmID)
}

// UntrackActiveExecution removes an execution from the active executions map
func (s *StateManager) UntrackActiveExecution(executionID string) {
	s.activeExecs.Delete(executionID)
}

// GetActiveExecutions returns all active executions
func (s *StateManager) GetActiveExecutions() map[string]string {
	result := make(map[string]string)
	s.activeExecs.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(string)
		return true
	})
	return result
}

// Close closes the state manager
func (s *StateManager) Close() {
	if s.cache != nil {
		s.cache.Close()
	}
}
