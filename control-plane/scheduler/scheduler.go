// Package scheduler provides functionality for managing function execution in the FaaS platform.
//
// The scheduler is responsible for:
// - Managing function execution requests (both synchronous and asynchronous)
// - Allocating VMs for function execution
// - Tracking execution state and results
// - Handling timeouts and errors
// - Providing an interface to retrieve execution results
//
// It works with the VMManager to allocate resources, the FunctionRegistry to retrieve
// function metadata and code, and the StateManager to persist execution state.
//
// The scheduler implements a worker pool pattern for handling asynchronous requests
// and includes monitoring capabilities to detect and handle stalled executions.

package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
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
	asyncQueue       chan *ExecutionRequest
	mu               sync.Mutex
	activeExecutions map[string]*ExecutionContext
}

// ExecutionRequest represents a request to execute a function
type ExecutionRequest struct {
	FunctionID   string
	FunctionName string
	Input        map[string]interface{}
	Sync         bool
	RequestID    string
}

// ExecutionContext tracks the context of a function execution
type ExecutionContext struct {
	RequestID  string
	FunctionID string
	VMID       string
	StartTime  time.Time
	Sync       bool
	Result     chan *ExecutionResult
}

// ExecutionResult represents the result of a function execution
type ExecutionResult struct {
	RequestID    string                 `json:"request_id"`
	FunctionID   string                 `json:"function_id"`
	StatusCode   int                    `json:"status_code"`
	Output       map[string]interface{} `json:"output,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Duration     int64                  `json:"duration_ms"`
	MemoryUsage  int64                  `json:"memory_usage_kb,omitempty"`
}

// NewScheduler creates a new function scheduler
func NewScheduler(vmManager *vm.VMManager, functionRegistry *registry.FunctionRegistry, stateManager *state.StateManager, logger *logrus.Logger) (*Scheduler, error) {
	scheduler := &Scheduler{
		vmManager:        vmManager,
		functionRegistry: functionRegistry,
		stateManager:     stateManager,
		logger:           logger,
		asyncQueue:       make(chan *ExecutionRequest, 100), // Buffer size of 100
		activeExecutions: make(map[string]*ExecutionContext),
	}

	// Start the async worker pool
	for i := 0; i < 5; i++ { // Start 5 worker goroutines
		go scheduler.asyncWorker()
	}

	// Start the execution monitor
	go scheduler.monitorExecutions()

	return scheduler, nil
}

// ScheduleExecution schedules a function for execution by ID
func (s *Scheduler) ScheduleExecution(functionID string, input map[string]interface{}, sync bool) (*ExecutionResult, error) {
	// Validate function exists
	_, err := s.functionRegistry.GetFunction(functionID)
	if err != nil {
		return nil, fmt.Errorf("function not found: %v", err)
	}

	// Create execution request
	requestID := uuid.New().String()
	request := &ExecutionRequest{
		FunctionID: functionID,
		Input:      input,
		Sync:       sync,
		RequestID:  requestID,
	}

	// Handle based on sync/async mode
	if sync {
		// For synchronous requests, execute directly and wait for result
		return s.executeFunction(request)
	} else {
		// For asynchronous requests, queue the execution and return immediately
		select {
		case s.asyncQueue <- request:
			// Successfully queued
			return &ExecutionResult{
				RequestID:  requestID,
				FunctionID: functionID,
				StatusCode: 202, // Accepted
			}, nil
		default:
			// Queue is full
			return nil, errors.New("execution queue is full, try again later")
		}
	}
}

// ScheduleExecutionByName schedules a function for execution by name
func (s *Scheduler) ScheduleExecutionByName(functionName string, input map[string]interface{}, sync bool) (*ExecutionResult, error) {
	// Validate function exists
	function, err := s.functionRegistry.GetFunctionByName(functionName)
	if err != nil {
		return nil, fmt.Errorf("function not found: %v", err)
	}

	// Create execution request
	requestID := uuid.New().String()
	request := &ExecutionRequest{
		FunctionID:   function.ID,
		FunctionName: functionName,
		Input:        input,
		Sync:         sync,
		RequestID:    requestID,
	}

	// Handle based on sync/async mode
	if sync {
		// For synchronous requests, execute directly and wait for result
		return s.executeFunction(request)
	} else {
		// For asynchronous requests, queue the execution and return immediately
		select {
		case s.asyncQueue <- request:
			// Successfully queued
			return &ExecutionResult{
				RequestID:  requestID,
				FunctionID: function.ID,
				StatusCode: 202, // Accepted
			}, nil
		default:
			// Queue is full
			return nil, errors.New("execution queue is full, try again later")
		}
	}
}

// GetExecutionResult retrieves the result of an asynchronous execution
func (s *Scheduler) GetExecutionResult(requestID string) (*ExecutionResult, error) {
	// Check if execution is still active
	s.mu.Lock()
	_, active := s.activeExecutions[requestID]
	s.mu.Unlock()

	if active {
		// Execution is still in progress
		return &ExecutionResult{
			RequestID:  requestID,
			StatusCode: 102, // Processing
		}, nil
	}

	// Check if execution result is in the database
	execution, err := s.stateManager.GetExecution(requestID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %v", err)
	}

	// Parse the output
	var output map[string]interface{}
	if execution.Logs != "" {
		if err := json.Unmarshal([]byte(execution.Logs), &output); err != nil {
			s.logger.Warnf("Failed to parse execution output: %v", err)
		}
	}

	// Return the result
	return &ExecutionResult{
		RequestID:    requestID,
		FunctionID:   execution.FunctionID,
		StatusCode:   200,
		Output:       output,
		ErrorMessage: execution.Error,
		Duration:     execution.Duration,
	}, nil
}

// executeFunction executes a function on a VM
func (s *Scheduler) executeFunction(request *ExecutionRequest) (*ExecutionResult, error) {
	// Get function metadata
	function, err := s.functionRegistry.GetFunction(request.FunctionID)
	if err != nil {
		return nil, fmt.Errorf("function not found: %v", err)
	}

	// Get function code
	code, err := s.functionRegistry.GetFunctionCode(request.FunctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get function code: %v", err)
	}

	// Create execution record
	execution := &state.Execution{
		ID:         request.RequestID,
		FunctionID: request.FunctionID,
		Status:     "pending",
		StartTime:  time.Now(),
	}
	if err := s.stateManager.SaveExecution(execution); err != nil {
		s.logger.Errorf("Failed to save execution record: %v", err)
	}

	// Allocate a VM for execution
	vmInstance, err := s.vmManager.GetVM()
	if err != nil {
		execution.Status = "failed"
		execution.Error = fmt.Sprintf("Failed to allocate VM: %v", err)
		execution.EndTime = time.Now()
		s.stateManager.SaveExecution(execution)
		return nil, fmt.Errorf("failed to allocate VM: %v", err)
	}

	// Track the execution
	resultChan := make(chan *ExecutionResult, 1)
	context := &ExecutionContext{
		RequestID:  request.RequestID,
		FunctionID: request.FunctionID,
		VMID:       vmInstance.ID,
		StartTime:  time.Now(),
		Sync:       request.Sync,
		Result:     resultChan,
	}

	s.mu.Lock()
	s.activeExecutions[request.RequestID] = context
	s.mu.Unlock()

	// Track in state manager
	s.stateManager.TrackActiveExecution(request.RequestID, vmInstance.ID)

	// Execute the function on the VM
	go func() {
		defer func() {
			// Cleanup
			s.mu.Lock()
			delete(s.activeExecutions, request.RequestID)
			s.mu.Unlock()
			s.stateManager.UntrackActiveExecution(request.RequestID)
			close(resultChan)
		}()

		// Update execution status
		execution.Status = "running"
		execution.VMID = vmInstance.ID
		s.stateManager.SaveExecution(execution)

		// Create payload for daemon
		payload := map[string]interface{}{
			"function_id":  request.FunctionID,
			"name":         function.Name,
			"code":         code.Code,
			"requirements": code.Requirements,
			"config":       code.Config,
			"runtime":      function.Runtime,
			"entry_point":  "handler.handler", // Default entry point
			"environment":  map[string]string{},
			"request_id":   request.RequestID,
			"timeout":      function.Timeout,
			"memory":       function.Memory,
			"version":      function.Version,
			"input":        request.Input,
		}

		// Convert payload to JSON
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			s.logger.Errorf("Failed to marshal function payload: %v", err)

			// Create error result
			errorResult := &ExecutionResult{
				RequestID:    request.RequestID,
				FunctionID:   request.FunctionID,
				StatusCode:   500,
				ErrorMessage: fmt.Sprintf("Failed to marshal function payload: %v", err),
				Duration:     time.Since(context.StartTime).Milliseconds(),
			}

			// Update execution record
			execution.Status = "failed"
			execution.Error = errorResult.ErrorMessage
			execution.EndTime = time.Now()
			execution.Duration = errorResult.Duration
			s.stateManager.SaveExecution(execution)

			// Return VM to pool
			if err := s.vmManager.ReturnVM(vmInstance.ID); err != nil {
				s.logger.Errorf("Failed to return VM to pool: %v", err)
			}

			// Send result to channel
			resultChan <- errorResult
			return
		}

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: time.Duration(function.Timeout+5) * time.Second, // Add 5 seconds buffer
		}

		// Construct daemon URL
		daemonURL := fmt.Sprintf("http://%s:8081/execute", vmInstance.IP)
		s.logger.Infof("Sending execution request to daemon at %s", daemonURL)

		// Send request to daemon
		resp, err := client.Post(daemonURL, "application/json", bytes.NewBuffer(payloadJSON))

		if err != nil {
			s.logger.Errorf("Failed to send request to daemon: %v", err)

			// Create error result
			errorResult := &ExecutionResult{
				RequestID:    request.RequestID,
				FunctionID:   request.FunctionID,
				StatusCode:   500,
				ErrorMessage: fmt.Sprintf("Failed to send request to daemon: %v", err),
				Duration:     time.Since(context.StartTime).Milliseconds(),
			}

			// Update execution record
			execution.Status = "failed"
			execution.Error = errorResult.ErrorMessage
			execution.EndTime = time.Now()
			execution.Duration = errorResult.Duration
			s.stateManager.SaveExecution(execution)

			// Return VM to pool
			// if err := s.vmManager.ReturnVM(vmInstance.ID); err != nil {
			// 	s.logger.Errorf("Failed to return VM to pool: %v", err)
			// }

			// Send result to channel
			resultChan <- errorResult
			return
		}
		defer resp.Body.Close()

		// For synchronous requests, we need to wait for the result
		if request.Sync {
			// The daemon will send the result to the control plane via a callback
			// We need to poll for the result
			maxRetries := 30 // Maximum number of retries
			retryInterval := 500 * time.Millisecond

			for i := 0; i < maxRetries; i++ {
				// Wait before checking
				time.Sleep(retryInterval)

				// Check if execution is complete
				execResult, err := s.stateManager.GetExecution(request.RequestID)
				if err != nil {
					continue
				}

				if execResult.Status == "completed" || execResult.Status == "failed" {
					// Execution is complete, parse the result
					var output map[string]interface{}
					if execResult.Logs != "" {
						if err := json.Unmarshal([]byte(execResult.Logs), &output); err != nil {
							s.logger.Warnf("Failed to parse execution output: %v", err)
						}
					}

					// Create result
					result := &ExecutionResult{
						RequestID:    request.RequestID,
						FunctionID:   request.FunctionID,
						StatusCode:   200,
						Output:       output,
						ErrorMessage: execResult.Error,
						Duration:     execResult.Duration,
					}

					if execResult.Status == "failed" {
						result.StatusCode = 500
					}

					// Return VM to pool
					if err := s.vmManager.ReturnVM(vmInstance.ID); err != nil {
						s.logger.Errorf("Failed to return VM to pool: %v", err)
					}

					// Send result to channel
					resultChan <- result
					return
				}
			}

			// If we get here, the execution timed out
			s.logger.Warnf("Execution timed out after %d retries", maxRetries)

			// Create timeout result
			timeoutResult := &ExecutionResult{
				RequestID:    request.RequestID,
				FunctionID:   request.FunctionID,
				StatusCode:   504, // Gateway Timeout
				ErrorMessage: "Execution timed out waiting for result",
				Duration:     time.Since(context.StartTime).Milliseconds(),
			}

			// Update execution record
			execution.Status = "timeout"
			execution.Error = timeoutResult.ErrorMessage
			execution.EndTime = time.Now()
			execution.Duration = timeoutResult.Duration
			s.stateManager.SaveExecution(execution)

			// Return VM to pool
			if err := s.vmManager.ReturnVM(vmInstance.ID); err != nil {
				s.logger.Errorf("Failed to return VM to pool: %v", err)
			}

			// Send result to channel
			resultChan <- timeoutResult
			return
		} else {
			// For asynchronous requests, we just acknowledge that the execution has started
			// The daemon will send the result to the control plane via a callback

			// Create accepted result
			acceptedResult := &ExecutionResult{
				RequestID:  request.RequestID,
				FunctionID: request.FunctionID,
				StatusCode: 202, // Accepted
			}

			// Send result to channel
			resultChan <- acceptedResult
		}
	}()

	// For synchronous requests, wait for the result
	if request.Sync {
		result := <-resultChan
		return result, nil
	}

	// For asynchronous requests, return immediately
	return &ExecutionResult{
		RequestID:  request.RequestID,
		FunctionID: request.FunctionID,
		StatusCode: 202, // Accepted
	}, nil
}

// asyncWorker processes asynchronous execution requests
func (s *Scheduler) asyncWorker() {
	for request := range s.asyncQueue {
		s.logger.Infof("Processing async request %s for function %s", request.RequestID, request.FunctionID)
		_, err := s.executeFunction(request)
		if err != nil {
			s.logger.Errorf("Failed to execute async function: %v", err)
		}
	}
}

// monitorExecutions monitors active executions for timeouts
func (s *Scheduler) monitorExecutions() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		s.mu.Lock()
		now := time.Now()
		for requestID, context := range s.activeExecutions {
			// Check if execution has been running for too long (more than 5 minutes)
			if now.Sub(context.StartTime) > 5*time.Minute {
				s.logger.Warnf("Execution %s has been running for too long, marking as timed out", requestID)

				// Get the execution from the state manager
				execution, err := s.stateManager.GetExecution(requestID)
				if err != nil {
					s.logger.Errorf("Failed to get execution %s: %v", requestID, err)
					continue
				}

				// Update execution status
				execution.Status = "timeout"
				execution.Error = "Execution timed out"
				execution.EndTime = now
				execution.Duration = now.Sub(context.StartTime).Milliseconds()
				s.stateManager.SaveExecution(execution)

				// Clean up the VM - since terminateVM is unexported, we'll use ReturnVM instead
				// This isn't ideal but will work until a proper public termination method is available
				if err := s.vmManager.ReturnVM(context.VMID); err != nil {
					s.logger.Errorf("Failed to clean up VM %s: %v", context.VMID, err)
				}

				// Remove from active executions
				delete(s.activeExecutions, requestID)
				s.stateManager.UntrackActiveExecution(requestID)
			}
		}
		s.mu.Unlock()
	}
}
