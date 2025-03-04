package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bluequbit/faas/control-plane/api"
	"github.com/bluequbit/faas/control-plane/auth"
	"github.com/bluequbit/faas/control-plane/registry"
	"github.com/bluequbit/faas/control-plane/scheduler"
	"github.com/bluequbit/faas/control-plane/state"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockFunctionRegistry struct {
	mock.Mock
}

func (m *MockFunctionRegistry) RegisterFunction(name, runtime string, memory, timeout int, code, requirements, config string) (*registry.FunctionMetadata, error) {
	args := m.Called(name, runtime, memory, timeout, code, requirements, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registry.FunctionMetadata), args.Error(1)
}

func (m *MockFunctionRegistry) UpdateFunction(id string, code, requirements, config string) (*registry.FunctionMetadata, error) {
	args := m.Called(id, code, requirements, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registry.FunctionMetadata), args.Error(1)
}

func (m *MockFunctionRegistry) GetFunction(id string) (*registry.FunctionMetadata, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registry.FunctionMetadata), args.Error(1)
}

func (m *MockFunctionRegistry) GetFunctionByName(name string) (*registry.FunctionMetadata, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registry.FunctionMetadata), args.Error(1)
}

func (m *MockFunctionRegistry) GetFunctionCode(id string) (*registry.FunctionCode, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*registry.FunctionCode), args.Error(1)
}

func (m *MockFunctionRegistry) ListFunctions() ([]registry.FunctionMetadata, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]registry.FunctionMetadata), args.Error(1)
}

func (m *MockFunctionRegistry) DeleteFunction(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockVMManager struct {
	mock.Mock
}

func (m *MockVMManager) GetVM() (*state.VM, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*state.VM), args.Error(1)
}

func (m *MockVMManager) ReturnVM(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockVMManager) GetVMStatus(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockVMManager) ListVMs() ([]state.VM, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]state.VM), args.Error(1)
}

func (m *MockVMManager) GetVMByID(id string) (*state.VM, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*state.VM), args.Error(1)
}

func (m *MockVMManager) Cleanup() {
	m.Called()
}

type MockScheduler struct {
	mock.Mock
}

func (m *MockScheduler) ScheduleExecution(functionID string, input map[string]interface{}, sync bool) (*scheduler.ExecutionResponse, error) {
	args := m.Called(functionID, input, sync)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*scheduler.ExecutionResponse), args.Error(1)
}

func (m *MockScheduler) ScheduleExecutionByName(functionName string, input map[string]interface{}, sync bool) (*scheduler.ExecutionResponse, error) {
	args := m.Called(functionName, input, sync)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*scheduler.ExecutionResponse), args.Error(1)
}

func (m *MockScheduler) GetExecution(id string) (*scheduler.ExecutionContext, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*scheduler.ExecutionContext), args.Error(1)
}

func (m *MockScheduler) ListExecutions(functionID string) ([]*scheduler.ExecutionContext, error) {
	args := m.Called(functionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*scheduler.ExecutionContext), args.Error(1)
}

type MockAuthManager struct {
	mock.Mock
}

func (m *MockAuthManager) GenerateAPIKey(userID string, roles []string, expiresIn time.Duration) (string, error) {
	args := m.Called(userID, roles, expiresIn)
	return args.String(0), args.Error(1)
}

func (m *MockAuthManager) ValidateAPIKey(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *MockAuthManager) HasRole(key string, role string) (bool, error) {
	args := m.Called(key, role)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthManager) Middleware(next http.Handler) http.Handler {
	// For testing, we'll just pass through without authentication
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (m *MockAuthManager) RoleMiddleware(role string, next http.Handler) http.Handler {
	// For testing, we'll just pass through without authorization
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Helper function to setup the test environment
func setupTest() (*api.APIHandler, registry.FunctionRegistry, vm.VMManager, scheduler.Scheduler, auth.AuthManager, *mux.Router) {
	// Initialize real implementations instead of mocks
	functionRegistry, err := registry.NewFunctionRegistry(stateManager, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create function registry: %v", err))
	}

	vmManager, err := vm.NewVMManager(stateManager, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create VM manager: %v", err))
	}

	sched, err := scheduler.NewScheduler(functionRegistry, vmManager)
	if err != nil {
		panic(fmt.Sprintf("Failed to create scheduler: %v", err))
	}

	authManager, err := auth.NewAuthManager("test_secret_key")
	if err != nil {
		panic(fmt.Sprintf("Failed to create auth manager: %v", err))
	}

	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil)) // Silence logs during tests

	apiHandler := api.NewAPIHandler(
		functionRegistry,
		vmManager,
		sched,
		authManager,
		logger,
	)

	router := mux.NewRouter()
	apiHandler.RegisterRoutes(router)

	return apiHandler, functionRegistry, vmManager, sched, authManager, router
}

// Test health endpoint
func TestHealthHandler(t *testing.T) {
	_, _, _, _, _, router := setupTest()

	req, _ := http.NewRequest("GET", "/api/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

// Test API key generation
func TestGenerateAPIKeyHandler(t *testing.T) {
	_, _, _, _, mockAuth, router := setupTest()

	// Setup mock
	mockAuth.On("GenerateAPIKey", "user123", []string{"admin"}, time.Duration(3600)*time.Second).Return("test-api-key", nil)

	// Create request
	reqBody := api.APIKeyRequest{
		UserID:    "user123",
		Roles:     []string{"admin"},
		ExpiresIn: 3600,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth/api-key", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "test-api-key", response["api_key"])

	mockAuth.AssertExpectations(t)
}

// Test function registration
func TestRegisterFunctionHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	functionMeta := &registry.FunctionMetadata{
		ID:      "func123",
		Name:    "test-function",
		Runtime: "python3.9",
		Memory:  128,
		Timeout: 30,
	}
	mockRegistry.On("RegisterFunction", "test-function", "python3.9", 128, 30, "def handler(): return 'hello'", "requests==2.26.0", "timeout: 30").Return(functionMeta, nil)

	// Create request
	reqBody := api.FunctionRequest{
		Name:         "test-function",
		Runtime:      "python3.9",
		Memory:       128,
		Timeout:      30,
		Code:         "def handler(): return 'hello'",
		Requirements: "requests==2.26.0",
		Config:       "timeout: 30",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/functions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response registry.FunctionMetadata
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "func123", response.ID)
	assert.Equal(t, "test-function", response.Name)

	mockRegistry.AssertExpectations(t)
}

// Test function registration failure
func TestRegisterFunctionHandler_Failure(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	mockRegistry.On("RegisterFunction", "test-function", "python3.9", 128, 30, "def handler(): return 'hello'", "requests==2.26.0", "timeout: 30").Return(nil, errors.New("function already exists"))

	// Create request
	reqBody := api.FunctionRequest{
		Name:         "test-function",
		Runtime:      "python3.9",
		Memory:       128,
		Timeout:      30,
		Code:         "def handler(): return 'hello'",
		Requirements: "requests==2.26.0",
		Config:       "timeout: 30",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/functions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Failed to register function")

	mockRegistry.AssertExpectations(t)
}

// Test get function by ID
func TestGetFunctionHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	functionMeta := &registry.FunctionMetadata{
		ID:      "func123",
		Name:    "test-function",
		Runtime: "python3.9",
		Memory:  128,
		Timeout: 30,
	}
	mockRegistry.On("GetFunction", "func123").Return(functionMeta, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/functions/func123", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response registry.FunctionMetadata
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "func123", response.ID)
	assert.Equal(t, "test-function", response.Name)

	mockRegistry.AssertExpectations(t)
}

// Test get function by name
func TestGetFunctionByNameHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	functionMeta := &registry.FunctionMetadata{
		ID:      "func123",
		Name:    "test-function",
		Runtime: "python3.9",
		Memory:  128,
		Timeout: 30,
	}
	mockRegistry.On("GetFunctionByName", "test-function").Return(functionMeta, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/functions/name/test-function", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response registry.FunctionMetadata
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "func123", response.ID)
	assert.Equal(t, "test-function", response.Name)

	mockRegistry.AssertExpectations(t)
}

// Test list functions
func TestListFunctionsHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	functions := []registry.FunctionMetadata{
		{
			ID:      "func123",
			Name:    "test-function-1",
			Runtime: "python3.9",
		},
		{
			ID:      "func456",
			Name:    "test-function-2",
			Runtime: "python3.9",
		},
	}
	mockRegistry.On("ListFunctions").Return(functions, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/functions", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []registry.FunctionMetadata
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "func123", response[0].ID)
	assert.Equal(t, "func456", response[1].ID)

	mockRegistry.AssertExpectations(t)
}

// Test update function
func TestUpdateFunctionHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	functionMeta := &registry.FunctionMetadata{
		ID:      "func123",
		Name:    "test-function",
		Runtime: "python3.9",
		Memory:  128,
		Timeout: 30,
		Version: "1.0.1",
	}
	mockRegistry.On("UpdateFunction", "func123", "def handler(): return 'updated'", "requests==2.27.0", "timeout: 60").Return(functionMeta, nil)

	// Create request
	reqBody := api.FunctionRequest{
		Code:         "def handler(): return 'updated'",
		Requirements: "requests==2.27.0",
		Config:       "timeout: 60",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/functions/func123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response registry.FunctionMetadata
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "func123", response.ID)
	assert.Equal(t, "1.0.1", response.Version)

	mockRegistry.AssertExpectations(t)
}

// Test delete function
func TestDeleteFunctionHandler(t *testing.T) {
	_, mockRegistry, _, _, _, router := setupTest()

	// Setup mock
	mockRegistry.On("DeleteFunction", "func123").Return(nil)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/functions/func123", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Function deleted", rr.Body.String())

	mockRegistry.AssertExpectations(t)
}

// Test invoke function
func TestInvokeFunctionHandler(t *testing.T) {
	_, _, _, mockScheduler, _, router := setupTest()

	// Setup mock
	executionResponse := &scheduler.ExecutionResponse{
		ExecutionID: "exec123",
		Status:      "completed",
		Result:      "Hello, World!",
		Duration:    100,
	}
	input := map[string]interface{}{"name": "World"}
	mockScheduler.On("ScheduleExecution", "func123", input, true).Return(executionResponse, nil)

	// Create request
	reqBody := api.InvokeRequest{
		Input: input,
		Sync:  true,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/functions/func123/invoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response scheduler.ExecutionResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "exec123", response.ExecutionID)
	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, "Hello, World!", response.Result)

	mockScheduler.AssertExpectations(t)
}

// Test invoke function by name
func TestInvokeFunctionByNameHandler(t *testing.T) {
	_, _, _, mockScheduler, _, router := setupTest()

	// Setup mock
	executionResponse := &scheduler.ExecutionResponse{
		ExecutionID: "exec123",
		Status:      "completed",
		Result:      "Hello, World!",
		Duration:    100,
	}
	input := map[string]interface{}{"name": "World"}
	mockScheduler.On("ScheduleExecutionByName", "test-function", input, true).Return(executionResponse, nil)

	// Create request
	reqBody := api.InvokeRequest{
		Input: input,
		Sync:  true,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/functions/name/test-function/invoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response scheduler.ExecutionResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "exec123", response.ExecutionID)
	assert.Equal(t, "completed", response.Status)
	assert.Equal(t, "Hello, World!", response.Result)

	mockScheduler.AssertExpectations(t)
}

// Test get execution
func TestGetExecutionHandler(t *testing.T) {
	_, _, _, mockScheduler, _, router := setupTest()

	// Setup mock
	executionContext := &scheduler.ExecutionContext{
		ID:         "exec123",
		FunctionID: "func123",
		Status:     "completed",
		Result:     "Hello, World!",
		Duration:   100,
	}
	mockScheduler.On("GetExecution", "exec123").Return(executionContext, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/executions/exec123", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response scheduler.ExecutionContext
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "exec123", response.ID)
	assert.Equal(t, "func123", response.FunctionID)
	assert.Equal(t, "completed", response.Status)

	mockScheduler.AssertExpectations(t)
}

// Test list executions
func TestListExecutionsHandler(t *testing.T) {
	_, _, _, mockScheduler, _, router := setupTest()

	// Setup mock
	executions := []*scheduler.ExecutionContext{
		{
			ID:         "exec123",
			FunctionID: "func123",
			Status:     "completed",
		},
		{
			ID:         "exec456",
			FunctionID: "func123",
			Status:     "failed",
		},
	}
	mockScheduler.On("ListExecutions", "func123").Return(executions, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/executions/function/func123", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []*scheduler.ExecutionContext
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "exec123", response[0].ID)
	assert.Equal(t, "exec456", response[1].ID)

	mockScheduler.AssertExpectations(t)
}

// Test list VMs
func TestListVMsHandler(t *testing.T) {
	_, _, mockVMManager, _, _, router := setupTest()

	// Setup mock
	vms := []state.VM{
		{
			ID:     "vm123",
			Status: "ready",
			IP:     "172.16.0.2",
		},
		{
			ID:     "vm456",
			Status: "busy",
			IP:     "172.16.0.3",
		},
	}
	mockVMManager.On("ListVMs").Return(vms, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/vms", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response []state.VM
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))
	assert.Equal(t, "vm123", response[0].ID)
	assert.Equal(t, "vm456", response[1].ID)

	mockVMManager.AssertExpectations(t)
}

// Test get VM by ID
func TestGetVMHandler(t *testing.T) {
	_, _, mockVMManager, _, _, router := setupTest()

	// Setup mock
	vm := &state.VM{
		ID:     "vm123",
		Status: "ready",
		IP:     "172.16.0.2",
	}
	mockVMManager.On("GetVMByID", "vm123").Return(vm, nil)

	// Create request
	req, _ := http.NewRequest("GET", "/api/vms/vm123", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response state.VM
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "vm123", response.ID)
	assert.Equal(t, "ready", response.Status)
	assert.Equal(t, "172.16.0.2", response.IP)

	mockVMManager.AssertExpectations(t)
}

// Test error handling for invalid requests
func TestInvalidRequestBody(t *testing.T) {
	_, _, _, _, _, router := setupTest()

	// Create invalid request
	req, _ := http.NewRequest("POST", "/api/functions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid request body")
}
