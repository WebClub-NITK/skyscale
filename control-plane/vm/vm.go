
// Package vm provides functionality for managing Firecracker micro-VMs.
//
// The VMManager manages the lifecycle of Firecracker micro-VMs, including:
// - Creating new VMs
// - Returning VMs to the warm pool
// - Terminating VMs


package vm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bluequbit/faas/control-plane/state"
	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VMManager manages the lifecycle of Firecracker micro-VMs
type VMManager struct {
	stateManager *state.StateManager
	logger       *logrus.Logger
	vmDir        string
	warmPoolSize int
	warmPool     chan *state.VM
	mu           sync.Mutex
	vms          map[string]*VMInstance
}

// VMInstance represents a running Firecracker VM instance
type VMInstance struct {
	ID        string
	IP        string
	Machine   *firecracker.Machine
	Status    string
	CreatedAt time.Time
	LastUsed  time.Time
	Memory    int
	CPU       int
	IsWarm    bool
}

// VMConfig represents the configuration for a VM
type VMConfig struct {
	Memory int
	CPU    int
	Kernel string
	RootFS string
}

// NewVMManager creates a new VM manager
func NewVMManager(stateManager *state.StateManager, logger *logrus.Logger) (*VMManager, error) {
	// Create VM directory if it doesn't exist
	vmDir := "vm-storage"
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return nil, err
	}

	manager := &VMManager{
		stateManager: stateManager,
		logger:       logger,
		vmDir:        vmDir,
		warmPoolSize: 5, // Default warm pool size
		warmPool:     make(chan *state.VM, 5),
		vms:          make(map[string]*VMInstance),
	}

	// Start warm pool manager
	go manager.manageWarmPool()

	return manager, nil
}

// manageWarmPool maintains a pool of pre-warmed VMs
func (m *VMManager) manageWarmPool() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			currentSize := len(m.warmPool)
			m.mu.Unlock()

			if currentSize < m.warmPoolSize {
				m.logger.Infof("Warm pool size: %d/%d, creating new warm VM", currentSize, m.warmPoolSize)
				vm, err := m.createVM(true)
				if err != nil {
					m.logger.Errorf("Failed to create warm VM: %v", err)
					continue
				}

				select {
				case m.warmPool <- vm:
					m.logger.Infof("Added VM %s to warm pool", vm.ID)
				default:
					// Pool is full, clean up the VM
					m.logger.Warnf("Warm pool is full, cleaning up VM %s", vm.ID)
					m.terminateVM(vm.ID)
				}
			}
		}
	}
}

// GetVM gets a VM from the warm pool or creates a new one
func (m *VMManager) GetVM() (*state.VM, error) {
	// Try to get a VM from the warm pool
	select {
	case vm := <-m.warmPool:
		m.logger.Infof("Using warm VM %s from pool", vm.ID)

		// Update VM status
		vm.Status = "busy"
		vm.LastUsed = time.Now()
		if err := m.stateManager.SaveVM(vm); err != nil {
			m.logger.Errorf("Failed to update VM status: %v", err)
		}

		return vm, nil
	default:
		// No warm VM available, create a new one
		m.logger.Info("No warm VM available, creating new VM")
		return m.createVM(false)
	}
}

// createVM creates a new Firecracker VM using the Go SDK
func (m *VMManager) createVM(isWarm bool) (*state.VM, error) {
	// Generate VM ID
	id := uuid.New().String()

	// Create VM directory
	vmDir := filepath.Join(m.vmDir, id)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return nil, err
	}

	// Assign IP address
	ip, err := m.assignIP()
	if err != nil {
		return nil, err
	}

	// Create VM configuration
	config := VMConfig{
		Memory: 128, // Default memory in MB
		CPU:    1,   // Default CPU count
		Kernel: "/home/bluequbit/Dev/faas/assets/hello-vmlinux.bin",
		RootFS: "/home/bluequbit/Dev/faas/assets/hello-rootfs.ext4",
	}

	// Create context for VM operations
	ctx := context.Background()

	// Socket path for Firecracker
	socketPath := filepath.Join(vmDir, "firecracker.sock")

	// Create Firecracker machine configuration
	fcCfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: config.Kernel,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off",
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				PathOnHost:   firecracker.String(config.RootFS),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(int64(config.CPU)),
			MemSizeMib: firecracker.Int64(int64(config.Memory)),
		},
		VMID:        id,
		LogLevel:    "Debug",
		LogFifo:     filepath.Join(vmDir, "firecracker.log"),
		MetricsFifo: filepath.Join(vmDir, "firecracker.metrics"),
	}

	// Create command for Firecracker
	cmd := firecracker.VMCommandBuilder{}.
		WithBin("/usr/local/bin/firecracker").
		WithSocketPath(socketPath).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	// Create machine options
	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(logrus.NewEntry(m.logger)),
		firecracker.WithProcessRunner(cmd),
	}

	// Create the machine
	machine, err := firecracker.NewMachine(ctx, fcCfg, machineOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine: %v", err)
	}

	// Start the machine
	if err := machine.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start machine: %v", err)
	}

	// Create VM instance
	vmInstance := &VMInstance{
		ID:      id,
		IP:      ip,
		Machine: machine,
		Status: func() string {
			if isWarm {
				return "ready"
			}
			return "busy"
		}(),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Memory:    config.Memory,
		CPU:       config.CPU,
		IsWarm:    isWarm,
	}

	// Store VM instance
	m.mu.Lock()
	m.vms[id] = vmInstance
	m.mu.Unlock()

	// Create VM in state manager
	vm := &state.VM{
		ID:        id,
		Status:    vmInstance.Status,
		IP:        ip,
		CreatedAt: vmInstance.CreatedAt,
		LastUsed:  vmInstance.LastUsed,
		Memory:    config.Memory,
		CPU:       config.CPU,
		IsWarm:    isWarm,
	}

	if err := m.stateManager.SaveVM(vm); err != nil {
		m.logger.Errorf("Failed to save VM to state manager: %v", err)
	}

	return vm, nil
}

// ReturnVM returns a VM to the warm pool
func (m *VMManager) ReturnVM(id string) error {
	// Get VM from state manager
	vm, err := m.stateManager.GetVM(id)
	if err != nil {
		return err
	}

	// Update VM status
	vm.Status = "ready"
	vm.LastUsed = time.Now()
	vm.IsWarm = true
	if err := m.stateManager.SaveVM(vm); err != nil {
		return err
	}

	// Add VM to warm pool
	select {
	case m.warmPool <- vm:
		m.logger.Infof("Returned VM %s to warm pool", id)
	default:
		// Pool is full, terminate the VM
		m.logger.Warnf("Warm pool is full, terminating VM %s", id)
		return m.terminateVM(id)
	}

	return nil
}

// terminateVM terminates a VM
func (m *VMManager) terminateVM(id string) error {
	m.mu.Lock()
	vmInstance, exists := m.vms[id]
	m.mu.Unlock()

	if !exists {
		return errors.New("VM not found")
	}

	// Stop the VM
	if err := vmInstance.Machine.StopVMM(); err != nil {
		m.logger.Errorf("Failed to stop VM: %v", err)
	}

	// Remove VM directory
	vmDir := filepath.Join(m.vmDir, id)
	if err := os.RemoveAll(vmDir); err != nil {
		m.logger.Errorf("Failed to remove VM directory: %v", err)
	}

	// Remove VM from state manager
	if err := m.stateManager.DeleteVM(id); err != nil {
		m.logger.Errorf("Failed to delete VM from state manager: %v", err)
	}

	// Remove VM from map
	m.mu.Lock()
	delete(m.vms, id)
	m.mu.Unlock()

	m.logger.Infof("Terminated VM %s", id)
	return nil
}

// assignIP assigns an IP address to a VM
func (m *VMManager) assignIP() (string, error) {
	// For simplicity, we'll use a hardcoded IP range
	// In a production environment, this would be more sophisticated
	return "172.16.0.2", nil
}

// Cleanup cleans up all VMs
func (m *VMManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, vmInstance := range m.vms {
		if err := vmInstance.Machine.StopVMM(); err != nil {
			m.logger.Errorf("Failed to stop VM: %v", err)
		}
		m.logger.Infof("Terminated VM %s during cleanup", id)
	}

	m.vms = make(map[string]*VMInstance)
}

// GetVMStatus gets the status of a VM
func (m *VMManager) GetVMStatus(id string) (string, error) {
	vm, err := m.stateManager.GetVM(id)
	if err != nil {
		return "", err
	}
	return vm.Status, nil
}

// ListVMs lists all VMs
func (m *VMManager) ListVMs() ([]state.VM, error) {
	return m.stateManager.ListVMs()
}

// GetVMByID gets a VM by ID
func (m *VMManager) GetVMByID(id string) (*state.VM, error) {
	return m.stateManager.GetVM(id)
}

// CreateTestHostVM creates a test VM that represents the host machine for testing
func (m *VMManager) CreateTestHostVM() (*state.VM, error) {
	m.logger.Info("Creating test host VM for testing")

	// Generate VM ID
	id := "host-vm-test"

	// Use the host machine's IP (localhost)
	ip := "127.0.0.1"

	// Create VM in state manager
	vm := &state.VM{
		ID:        id,
		Status:    "ready",
		IP:        ip,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		Memory:    1024, // 1GB
		CPU:       2,    // 2 cores
		IsWarm:    true,
	}

	if err := m.stateManager.SaveVM(vm); err != nil {
		return nil, fmt.Errorf("failed to save test VM to state manager: %v", err)
	}

	// Create VM instance (without actual Firecracker machine)
	vmInstance := &VMInstance{
		ID:        id,
		IP:        ip,
		Machine:   nil, // No actual Firecracker machine
		Status:    "ready",
		CreatedAt: vm.CreatedAt,
		LastUsed:  vm.LastUsed,
		Memory:    vm.Memory,
		CPU:       vm.CPU,
		IsWarm:    true,
	}

	// Store VM instance
	m.mu.Lock()
	m.vms[id] = vmInstance
	m.mu.Unlock()

	m.logger.Infof("Created test host VM with ID %s and IP %s", id, ip)
	return vm, nil
}

// GetOrCreateTestHostVM gets the test host VM if it exists, or creates it if it doesn't
func (m *VMManager) GetOrCreateTestHostVM() (*state.VM, error) {
	// Check if test host VM already exists
	vm, err := m.stateManager.GetVM("host-vm-test")
	if err == nil {
		// VM exists, return it
		return vm, nil
	}

	// VM doesn't exist, create it
	return m.CreateTestHostVM()
}
