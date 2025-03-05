package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bluequbit/faas/control-plane/vm"
	"github.com/sirupsen/logrus"
)

// TestMode indicates whether the control plane is running in test mode
var TestMode bool

// TestHostVMID is the ID of the test host VM
const TestHostVMID = "host-vm-test"

func init() {
	// Add a command-line flag for test mode
	flag.BoolVar(&TestMode, "test", false, "Run in test mode with a simulated host VM")
}

// SetupTestEnvironment sets up the test environment if running in test mode
func SetupTestEnvironment(vmManager *vm.VMManager, logger *logrus.Logger) error {
	if !TestMode {
		return nil
	}

	logger.Info("Setting up test environment")

	// Create test host VM
	hostVM, err := vmManager.GetOrCreateTestHostVM()
	if err != nil {
		return fmt.Errorf("failed to create test host VM: %v", err)
	}

	logger.Infof("Created test host VM with ID %s and IP %s", hostVM.ID, hostVM.IP)

	// Check if the daemon is already running
	if !isDaemonRunning() {
		// Start the daemon in the background
		if err := startDaemon(logger); err != nil {
			return fmt.Errorf("failed to start daemon: %v", err)
		}
	}

	logger.Info("Test environment setup complete")
	return nil
}

// isDaemonRunning checks if the daemon is already running
func isDaemonRunning() bool {
	// Simple check: try to connect to the daemon's health endpoint
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://localhost:8081/health")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If we get a 200 OK, the daemon is running
	return string(output) == "200"
}

// startDaemon starts the daemon process
func startDaemon(logger *logrus.Logger) error {
	logger.Info("Starting daemon process")

	// Get the path to the daemon binary
	daemonPath := filepath.Join(os.Getenv("HOME"), "Dev", "faas", "cmd", "deamon", "daemon")
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		return fmt.Errorf("daemon binary not found at %s", daemonPath)
	}

	// Set environment variables for the daemon
	env := []string{
		fmt.Sprintf("VM_ID=%s", TestHostVMID),
		"VM_IP=127.0.0.1",
		"PATH=" + os.Getenv("PATH"),
	}

	// Start the daemon process
	cmd := exec.Command(daemonPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process in the background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon process: %v", err)
	}

	logger.Infof("Started daemon process with PID %d", cmd.Process.Pid)

	// Don't wait for the process to complete
	go func() {
		if err := cmd.Wait(); err != nil {
			logger.Errorf("Daemon process exited with error: %v", err)
		}
	}()

	// Wait a moment for the daemon to start
	logger.Info("Waiting for daemon to start...")
	waitCmd := exec.Command("sleep", "2")
	waitCmd.Run()

	// Verify the daemon is running
	if !isDaemonRunning() {
		return fmt.Errorf("daemon failed to start")
	}

	logger.Info("Daemon started successfully")
	return nil
}
