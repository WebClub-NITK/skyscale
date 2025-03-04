package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	controlPlaneURL = "http://172.16.0.1:8080" // Control plane URL (host machine)
	pollInterval    = 5 * time.Second
	codeDir         = "/tmp/faas/code"
	logDir          = "/var/log/faas"

	// Endpoints
	functionEndpoint = "/api/v1/functions"
	resultEndpoint   = "/api/v1/results"
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
}

func main() {
	log.Printf("Starting FaaS daemon on %s (ID: %s)", vmInfo.MachineName, vmInfo.VMID)

	// Configure HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For development only
			},
		},
	}

	// Main polling loop
	for {
		payload, err := pollForFunction(client)
		if err != nil {
			log.Printf("Error polling for function: %v", err)
			time.Sleep(pollInterval)
			continue
		}

		if payload != nil {
			log.Printf("Received function execution request: %s (ID: %s)", payload.Name, payload.RequestID)
			vmInfo.Status = "busy"

			// Execute the function
			result := executeFunction(payload)

			// Send the result back to the control plane
			err := sendResult(client, result)
			if err != nil {
				log.Printf("Error sending result: %v", err)
			}

			// Mark VM as ready again
			vmInfo.Status = "ready"
		}

		time.Sleep(pollInterval)
	}
}

// pollForFunction checks the control plane for new functions to execute
func pollForFunction(client *http.Client) (*FunctionPayload, error) {
	data, err := json.Marshal(vmInfo)
	if err != nil {
		return nil, fmt.Errorf("error marshaling VM info: %v", err)
	}

	resp, err := client.Post(
		fmt.Sprintf("%s%s/poll", controlPlaneURL, functionEndpoint),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No new function to execute
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload FunctionPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("error decoding payload: %v", err)
	}

	return &payload, nil
}

// executeFunction prepares and executes the function code
func executeFunction(payload *FunctionPayload) *ExecutionResult {
	startTime := time.Now()
	result := &ExecutionResult{
		RequestID:  payload.RequestID,
		FunctionID: payload.FunctionID,
		StatusCode: 500, // Default to error
	}

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
	} else {
		result.StatusCode = 200
		result.Output = output
	}

	return result
}

// prepareFunction writes the function code and requirements to disk
func prepareFunction(payload *FunctionPayload, execDir string) error {
	// Write handler.py
	if err := ioutil.WriteFile(filepath.Join(execDir, "handler.py"), []byte(payload.Code), 0644); err != nil {
		return fmt.Errorf("failed to write handler.py: %v", err)
	}

	// Write requirements.txt
	if err := ioutil.WriteFile(filepath.Join(execDir, "requirements.txt"), []byte(payload.Requirements), 0644); err != nil {
		return fmt.Errorf("failed to write requirements.txt: %v", err)
	}

	// Write config file
	if err := ioutil.WriteFile(filepath.Join(execDir, "faas.yaml"), []byte(payload.Config), 0644); err != nil {
		return fmt.Errorf("failed to write faas.yaml: %v", err)
	}

	// Install requirements if any
	if payload.Requirements != "" {
		cmd := exec.Command("pip", "install", "-r", filepath.Join(execDir, "requirements.txt"))
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

		// Create a Python script to import and call the function
		runnerScript := fmt.Sprintf(`
import sys
import json
import os

sys.path.append("%s")
from %s import %s

try:
    # Set environment variables
    %s
    
    # Call the function
    result = %s()
    print(json.dumps({"result": result}))
    sys.exit(0)
except Exception as e:
    print(json.dumps({"error": str(e)}))
    sys.exit(1)
`,
			execDir,
			parts[0],
			parts[1],
			generateEnvSetup(payload.Environment),
			parts[1],
		)

		// Write the runner script
		runnerPath := filepath.Join(execDir, "_runner.py")
		if err := ioutil.WriteFile(runnerPath, []byte(runnerScript), 0755); err != nil {
			return "", fmt.Errorf("failed to write runner script: %v", err)
		}

		// Execute the runner script
		cmd = exec.CommandContext(ctx, "python3", runnerPath)

	case "node", "nodejs", "node14", "node16":
		// Similar implementation for Node.js
		return "", fmt.Errorf("nodejs runtime not implemented yet")

	default:
		return "", fmt.Errorf("unsupported runtime: %s", payload.Runtime)
	}

	// Set up command environment
	cmd.Dir = execDir
	cmd.Env = os.Environ()
	for k, v := range payload.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return stdout.String() + "\n" + stderr.String(), fmt.Errorf("function execution timed out after %d seconds", payload.Timeout)
	}

	if err != nil {
		return stdout.String() + "\n" + stderr.String(), fmt.Errorf("execution failed: %v", err)
	}

	return stdout.String(), nil
}

// generateEnvSetup creates Python code to set environment variables
func generateEnvSetup(env map[string]string) string {
	if len(env) == 0 {
		return "pass  # No environment variables to set"
	}

	var lines []string
	for k, v := range env {
		lines = append(lines, fmt.Sprintf("os.environ['%s'] = '%s'", k, v))
	}
	return strings.Join(lines, "\n")
}

// sendResult sends the execution result back to the control plane
func sendResult(client *http.Client, result *ExecutionResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	return nil
}
