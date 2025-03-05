package scheduler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bluequbit/faas/control-plane/registry"
	"github.com/bluequbit/faas/control-plane/state"
	"github.com/bluequbit/faas/control-plane/vm"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Scheduler manages function execution scheduling
type Scheduler struct {
	vmManager        *vm.VMManager
	functionRegistry *registry.FunctionRegistry
	stateManager     *state.StateManager
	logger           *logrus.Logger
	executions       map[string]*ExecutionContext
	mu               sync.Mutex
	queue            chan *ExecutionRequest
	workers          int
}

// ExecutionContext tracks the execution of a function
type ExecutionContext struct {
	ID         string
	FunctionID string
	VMID       string
	Status     string
	StartTime  time.Time
	EndTime    time.Time
	Duration   int64
	Logs       string
	Error      string
	Result     string
}

// ExecutionRequest represents a request to execute a function
type ExecutionRequest struct {
	ExecutionID  string
	FunctionID   string
	FunctionName string
	Input        map[string]interface{}
	Sync         bool
	ResponseChan chan *ExecutionResponse
}

// ExecutionResponse represents the response from a function execution
type ExecutionResponse struct {
	ExecutionID string
	Status      string
	Result      string
	Error       string
	Duration    int64
}

// NewScheduler creates a new scheduler
func NewScheduler(vmManager *vm.VMManager, functionRegistry *registry.FunctionRegistry, stateManager *state.StateManager, logger *logrus.Logger) (*Scheduler, error) {
	scheduler := &Scheduler{
		vmManager:        vmManager,
		functionRegistry: functionRegistry,
		stateManager:     stateManager,
		logger:           logger,
		executions:       make(map[string]*ExecutionContext),
		queue:            make(chan *ExecutionRequest, 100),
		workers:          10, // Default number of workers
	}

	// Start worker pool
	for i := 0; i < scheduler.workers; i++ {
		go scheduler.worker()
	}

	return scheduler, nil
}

// worker processes execution requests from the queue
func (s *Scheduler) worker() {
	for req := range s.queue {
		s.logger.Infof("Processing execution request for function: %s", req.FunctionID)

		// Execute the function
		resp, err := s.executeFunction(req)
		if err != nil {
			s.logger.Errorf("Failed to execute function: %v", err)
			if req.ResponseChan != nil {
				req.ResponseChan <- &ExecutionResponse{
					Status: "error",
					Error:  err.Error(),
				}
			}
			continue
		}

		// Send response if synchronous
		if req.Sync && req.ResponseChan != nil {
			req.ResponseChan <- resp
		}
	}
}

// ScheduleExecution schedules a function for execution
func (s *Scheduler) ScheduleExecution(functionID string, input map[string]interface{}, sync bool) (*ExecutionResponse, error) {
	// Generate execution ID
	executionID := uuid.New().String()

	// Create execution request
	req := &ExecutionRequest{
		ExecutionID: executionID,
		FunctionID:  functionID,
		Input:       input,
		Sync:        sync,
	}

	// If synchronous, create response channel
	if sync {
		req.ResponseChan = make(chan *ExecutionResponse, 1)
	}

	// Add request to queue
	s.queue <- req

	// If synchronous, wait for response
	if sync {
		resp := <-req.ResponseChan
		return resp, nil
	}

	// If asynchronous, return immediately with the execution ID
	return &ExecutionResponse{
		ExecutionID: executionID,
		Status:      "scheduled",
	}, nil
}

// ScheduleExecutionByName schedules a function for execution by name
func (s *Scheduler) ScheduleExecutionByName(functionName string, input map[string]interface{}, sync bool) (*ExecutionResponse, error) {
	// Get function by name
	function, err := s.functionRegistry.GetFunctionByName(functionName)
	if err != nil {
		return nil, err
	}

	// Schedule execution
	return s.ScheduleExecution(function.ID, input, sync)
}

// executeFunction executes a function on a VM
func (s *Scheduler) executeFunction(req *ExecutionRequest) (*ExecutionResponse, error) {
	// Use the execution ID from the request
	executionID := req.ExecutionID

	// Create execution context
	ctx := &ExecutionContext{
		ID:         executionID,
		FunctionID: req.FunctionID,
		Status:     "pending",
		StartTime:  time.Now(),
	}

	// Store execution context
	s.mu.Lock()
	s.executions[executionID] = ctx
	s.mu.Unlock()

	// Get function
	function, err := s.functionRegistry.GetFunction(req.FunctionID)
	if err != nil {
		ctx.Status = "error"
		ctx.Error = fmt.Sprintf("Function not found: %v", err)
		return s.finalizeExecution(ctx)
	}

	// Get function code
	functionCode, err := s.functionRegistry.GetFunctionCode(req.FunctionID)
	if err != nil {
		ctx.Status = "error"
		ctx.Error = fmt.Sprintf("Failed to get function code: %v", err)
		return s.finalizeExecution(ctx)
	}

	// Get VM
	vm, err := s.vmManager.GetVM()
	if err != nil {
		ctx.Status = "error"
		ctx.Error = fmt.Sprintf("Failed to get VM: %v", err)
		return s.finalizeExecution(ctx)
	}

	// Update execution context
	ctx.VMID = vm.ID
	ctx.Status = "running"

	// Track active execution
	s.stateManager.TrackActiveExecution(executionID, vm.ID)

	// Create execution in state manager
	execution := &state.Execution{
		ID:         executionID,
		FunctionID: req.FunctionID,
		Status:     "running",
		StartTime:  ctx.StartTime,
		VMID:       vm.ID,
	}
	if err := s.stateManager.SaveExecution(execution); err != nil {
		s.logger.Errorf("Failed to save execution: %v", err)
	}

	// Execute function on VM
	result, err := s.executeOnVM(vm, function, functionCode, req.Input)
	if err != nil {
		ctx.Status = "error"
		ctx.Error = fmt.Sprintf("Execution failed: %v", err)

		// Return VM to pool
		if err := s.vmManager.ReturnVM(vm.ID); err != nil {
			s.logger.Errorf("Failed to return VM to pool: %v", err)
		}

		return s.finalizeExecution(ctx)
	}

	// Update execution context
	ctx.Status = "completed"
	ctx.Result = result
	ctx.EndTime = time.Now()
	ctx.Duration = ctx.EndTime.Sub(ctx.StartTime).Milliseconds()

	// Return VM to pool
	if err := s.vmManager.ReturnVM(vm.ID); err != nil {
		s.logger.Errorf("Failed to return VM to pool: %v", err)
	}

	// Untrack active execution
	s.stateManager.UntrackActiveExecution(executionID)

	// Finalize execution
	return s.finalizeExecution(ctx)
}

// executeOnVM executes a function on a VM
func (s *Scheduler) executeOnVM(vm *state.VM, function *registry.FunctionMetadata, code *registry.FunctionCode, input map[string]interface{}) (string, error) {
	// In a real implementation, this would communicate with the VM agent
	// For now, we'll simulate execution with a simple HTTP request

	// Construct URL
	url := fmt.Sprintf("http://%s:8080/api/execute", vm.IP)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(function.Timeout) * time.Second,
	}

	// Create request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Function-ID", function.ID)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("execution failed with status: %s", resp.Status)
	}

	// For now, return a simulated result
	return "Function executed successfully", nil
}

// finalizeExecution finalizes an execution and returns the response
func (s *Scheduler) finalizeExecution(ctx *ExecutionContext) (*ExecutionResponse, error) {
	// Update execution in state manager
	execution := &state.Execution{
		ID:         ctx.ID,
		FunctionID: ctx.FunctionID,
		Status:     ctx.Status,
		StartTime:  ctx.StartTime,
		EndTime:    ctx.EndTime,
		Duration:   ctx.Duration,
		VMID:       ctx.VMID,
		Logs:       ctx.Logs,
		Error:      ctx.Error,
	}
	if err := s.stateManager.SaveExecution(execution); err != nil {
		s.logger.Errorf("Failed to save execution: %v", err)
	}

	// Return response
	return &ExecutionResponse{
		ExecutionID: ctx.ID,
		Status:      ctx.Status,
		Result:      ctx.Result,
		Error:       ctx.Error,
		Duration:    ctx.Duration,
	}, nil
}

// GetExecution gets an execution by ID
func (s *Scheduler) GetExecution(id string) (*ExecutionContext, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, exists := s.executions[id]
	if !exists {
		// Try to get from state manager
		execution, err := s.stateManager.GetExecution(id)
		if err != nil {
			return nil, err
		}

		ctx = &ExecutionContext{
			ID:         execution.ID,
			FunctionID: execution.FunctionID,
			VMID:       execution.VMID,
			Status:     execution.Status,
			StartTime:  execution.StartTime,
			EndTime:    execution.EndTime,
			Duration:   execution.Duration,
			Logs:       execution.Logs,
			Error:      execution.Error,
		}
	}

	return ctx, nil
}

// ListExecutions lists all executions for a function
func (s *Scheduler) ListExecutions(functionID string) ([]*ExecutionContext, error) {
	executions, err := s.stateManager.ListExecutions(functionID)
	if err != nil {
		return nil, err
	}

	result := make([]*ExecutionContext, len(executions))
	for i, execution := range executions {
		result[i] = &ExecutionContext{
			ID:         execution.ID,
			FunctionID: execution.FunctionID,
			VMID:       execution.VMID,
			Status:     execution.Status,
			StartTime:  execution.StartTime,
			EndTime:    execution.EndTime,
			Duration:   execution.Duration,
			Logs:       execution.Logs,
			Error:      execution.Error,
		}
	}

	return result, nil
}
