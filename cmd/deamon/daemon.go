package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Configuration
	controlPlaneURL = "http://localhost:8080" // Control plane URL (host machine)
	daemonPort      = "8081"                  // Port for the daemon to listen on
	codeDir         = "/tmp/faas/code"
	logDir          = "/var/log/faas"

	// Endpoints
	functionEndpoint = "/api/functions"
	resultEndpoint   = "/api/results"
	registerEndpoint = "/api/vms/register"
)

// FunctionPayload represents the code and metadata to be executed
type FunctionPayload struct {
	FunctionID   string            `json:"function_id"`
	Name         string            `json:"name"`
	Code         string            `json:"code"`         // Function code
	Requirements string            `json:"requirements"` // Python requirements
	Config       string            `json:"config"`       // Function configuration
	Runtime      string            `json:"runtime"`      // e.g., "python3.9"
	EntryPoint   string            `json:"entry_point"`  // e.g., "handler.handler"
	Environment  map[string]string `json:"environment"`  // Environment variables
	RequestID    string            `json:"request_id"`   // Unique ID for this execution request
	Timeout      int               `json:"timeout"`      // Execution timeout in seconds
	Memory       int               `json:"memory"`       // Memory limit in MB
	Version      string            `json:"version"`      // Function version
}

// ExecutionResult represents the result of function execution
type ExecutionResult struct {
	RequestID    string `json:"request_id"`
	FunctionID   string `json:"function_id"`
	StatusCode   int    `json:"status_code"`
	Output       string `json:"output"`
	ErrorMessage string `json:"error_message,omitempty"`
	Duration     int64  `json:"duration_ms"`
	MemoryUsage  int64  `json:"memory_usage_kb,omitempty"`
}

// VMInfo contains information about this VM instance
type VMInfo struct {
	VMID        string `json:"vm_id"`
	IPAddress   string `json:"ip_address"`
	MachineName string `json:"machine_name"`
	Status      string `json:"status"`
}

var vmInfo VMInfo
var httpClient *http.Client

func init() {
	// Create necessary directories
	os.MkdirAll(codeDir, 0755)
	os.MkdirAll(logDir, 0755)

	// Initialize VM info
	hostname, _ := os.Hostname()
	vmInfo = VMInfo{
		VMID:        os.Getenv("VM_ID"),
		IPAddress:   os.Getenv("VM_IP"),
		MachineName: hostname,
		Status:      "ready",
	}

	// Set up logging
	logFile, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	// Configure HTTP client
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For development only
			},
		},
	}
}

func main() {
	log.Printf("Starting FaaS daemon on %s (ID: %s)", vmInfo.MachineName, vmInfo.VMID)

	// Register VM with control plane
	// if err := registerVM(); err != nil {
	// 	log.Fatalf("Failed to register VM with control plane: %v", err)
	// }

	// Set up HTTP server for receiving function execution requests
	http.HandleFunc("/execute", handleExecuteRequest)
	http.HandleFunc("/health", handleHealthCheck)

	// Start HTTP server
	log.Printf("Starting HTTP server on port %s", daemonPort)
	if err := http.ListenAndServe(":"+daemonPort, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

// registerVM registers this VM with the control plane
func registerVM() error {
	data, err := json.Marshal(vmInfo)
	if err != nil {
		return fmt.Errorf("error marshaling VM info: %v", err)
	}

	resp, err := httpClient.Post(
		fmt.Sprintf("%s%s", controlPlaneURL, registerEndpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("VM registered successfully with control plane")
	return nil
}

// handleHealthCheck handles health check requests
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleExecuteRequest handles function execution requests
func handleExecuteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var payload FunctionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Received function execution request: %s (ID: %s)", payload.Name, payload.RequestID)

	// Update VM status
	vmInfo.Status = "busy"

	// Execute the function asynchronously
	go func() {
		// Execute the function
		result := executeFunction(&payload)

		// Send the result back to the control plane
		if err := sendResult(httpClient, result); err != nil {
			log.Printf("Error sending result: %v", err)
		}

		// Mark VM as ready again
		vmInfo.Status = "ready"

		// Report VM status back to control plane
		if err := reportVMStatus(); err != nil {
			log.Printf("Error reporting VM status: %v", err)
		}
	}()

	// Respond immediately to indicate the request was accepted
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Function execution started"))
}

// reportVMStatus reports the current VM status to the control plane
func reportVMStatus() error {
	data, err := json.Marshal(vmInfo)
	if err != nil {
		return fmt.Errorf("error marshaling VM info: %v", err)
	}

	resp, err := httpClient.Post(
		fmt.Sprintf("%s%s", controlPlaneURL, registerEndpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// executeFunction prepares and executes the function code
func executeFunction(payload *FunctionPayload) *ExecutionResult {
	startTime := time.Now()
	result := &ExecutionResult{
		RequestID:  payload.RequestID,
		FunctionID: payload.FunctionID,
		StatusCode: 500, // Default to error
	}

	log.Printf("Starting execution of function %s (ID: %s)", payload.Name, payload.RequestID)

	// Create a directory for this execution
	execDir := filepath.Join(codeDir, payload.RequestID)
	if err := os.MkdirAll(execDir, 0755); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create execution directory: %v", err)
		return result
	}
	defer os.RemoveAll(execDir) // Clean up after execution

	// Write function code and requirements
	if err := prepareFunction(payload, execDir); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to prepare function: %v", err)
		return result
	}

	// Execute the function
	output, err := runFunction(payload, execDir)
	duration := time.Since(startTime).Milliseconds()

	result.Duration = duration
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Execution error: %v", err)
		result.Output = output // Include any partial output
		log.Printf("Function execution failed: %v", err)
	} else {
		result.StatusCode = 200
		result.Output = output
		log.Printf("Function execution completed successfully in %d ms", duration)
	}

	// Track memory usage if available
	// This is a placeholder - in a real implementation, you would measure actual memory usage
	result.MemoryUsage = 0

	return result
}

// prepareFunction writes the function code and requirements to disk
func prepareFunction(payload *FunctionPayload, execDir string) error {
	// Write handler.py
	if err := os.WriteFile(filepath.Join(execDir, "handler.py"), []byte(payload.Code), 0644); err != nil {
		return fmt.Errorf("failed to write handler.py: %v", err)
	}

	// Write requirements.txt
	if err := os.WriteFile(filepath.Join(execDir, "requirements.txt"), []byte(payload.Requirements), 0644); err != nil {
		return fmt.Errorf("failed to write requirements.txt: %v", err)
	}

	// Write config file
	if err := os.WriteFile(filepath.Join(execDir, "faas.yaml"), []byte(payload.Config), 0644); err != nil {
		return fmt.Errorf("failed to write faas.yaml: %v", err)
	}

	// Install requirements if any
	if payload.Requirements != "" {
		// Create a virtual environment
		venvPath := filepath.Join(execDir, "venv")
		createVenvCmd := exec.Command("python3", "-m", "venv", venvPath)
		createVenvCmd.Dir = execDir
		if output, err := createVenvCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to create virtual environment: %v, output: %s", err, output)
		}

		// Install requirements in the virtual environment
		pipPath := filepath.Join(venvPath, "bin", "pip")
		cmd := exec.Command(pipPath, "install", "-r", filepath.Join(execDir, "requirements.txt"))
		cmd.Dir = execDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to install requirements: %v, output: %s", err, output)
		}
	}

	return nil
}

// runFunction executes the function with the specified runtime
func runFunction(payload *FunctionPayload, execDir string) (string, error) {
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(payload.Timeout)*time.Second)
	defer cancel()

	switch payload.Runtime {
	case "python3", "python3.9", "python3.10":
		// Parse entry point (format: "file.function")
		entryPoint := "handler.handler"
		if payload.EntryPoint != "" {
			entryPoint = payload.EntryPoint
		}

		parts := strings.Split(entryPoint, ".")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid entry point format: %s", entryPoint)
		}

		file, function := parts[0], parts[1]

		// Create Python script to execute the function
		executorCode := fmt.Sprintf(`
import sys
import json
import traceback
import %s

try:
    # Set up environment variables
    %s
    
    # Execute function
    result = %s.%s()
    print(result)
    sys.exit(0)
except Exception as e:
    print("Error:", str(e))
    traceback.print_exc()
    sys.exit(1)
`, file, generateEnvSetup(payload.Environment), file, function)

		// Write executor script
		if err := os.WriteFile(filepath.Join(execDir, "executor.py"), []byte(executorCode), 0644); err != nil {
			return "", fmt.Errorf("failed to write executor.py: %v", err)
		}

		// Determine which Python interpreter to use
		pythonInterpreter := "python3"
		if payload.Requirements != "" {
			// Use the virtual environment's Python interpreter if we created one
			venvPath := filepath.Join(execDir, "venv")
			pythonInterpreter = filepath.Join(venvPath, "bin", "python")
		}

		// Execute the function
		cmd = exec.CommandContext(ctx, pythonInterpreter, filepath.Join(execDir, "executor.py"))
	default:
		return "", fmt.Errorf("unsupported runtime: %s", payload.Runtime)
	}

	// Set working directory
	cmd.Dir = execDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		return output, fmt.Errorf("execution failed: %v, stderr: %s", err, stderr.String())
	}

	return output, nil
}

// generateEnvSetup generates Python code to set environment variables
func generateEnvSetup(env map[string]string) string {
	if len(env) == 0 {
		return "pass"
	}

	var lines []string
	for k, v := range env {
		lines = append(lines, fmt.Sprintf("os.environ['%s'] = '%s'", k, v))
	}

	return "import os\n" + strings.Join(lines, "\n")
}

// sendResult sends the execution result back to the control plane
func sendResult(client *http.Client, result *ExecutionResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

	log.Printf("Sending execution result for request ID: %s", result.RequestID)

	resp, err := client.Post(
		fmt.Sprintf("%s%s", controlPlaneURL, resultEndpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("Result sent successfully for request ID: %s", result.RequestID)
	return nil
}
