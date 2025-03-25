// create_test_vm.go
// This script creates a test VM using Firecracker SDK for development and testing.
// It's independent of the main application code but uses the same Firecracker SDK.
//
// Usage:
//   sudo go run scripts/create_test_vm.go
//
// Requirements:
//   - Firecracker binary in /usr/local/bin/firecracker
//   - CNI plugins installed and configured
//   - Kernel and rootfs images available
//   - Root privileges (sudo) for network namespace operations

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	kernelPath     = flag.String("kernel", "../assets/vmlinux-6.1.129.bin", "Path to the kernel image")
	rootfsPath     = flag.String("rootfs", "../assets/ubuntu-22.04.ext4", "Path to the rootfs image")
	vcpuCount      = flag.Int("cpu", 1, "Number of vCPUs")
	memSize        = flag.Int("mem", 128, "Memory size in MB")
	firecrackerBin = flag.String("fc", "/usr/local/bin/firecracker", "Path to firecracker binary")
	networkName    = flag.String("net", "fcnet", "CNI network name")
	socketName     = flag.String("socket", "firecracker.sock", "Firecracker socket name")
	debug          = flag.Bool("debug", false, "Enable debug logging")
	storageDir     = flag.String("dir", "/tmp/fc-test-vm", "Directory to store VM files")
	skipNetNs      = flag.Bool("skip-netns", false, "Skip network namespace setup (for debugging)")
)

// checkRoot checks if the script is running with root privileges
func checkRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}

func main() {
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	if *debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Check for root privileges
	if !checkRoot() {
		logger.Fatal("This script requires root privileges. Please run with sudo.")
	}

	// Create VM ID
	vmID := uuid.New().String()
	logger.WithField("vm_id", vmID).Info("Creating test VM")

	// Create storage directory
	vmDir := filepath.Join(*storageDir, vmID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		logger.WithError(err).Fatal("Failed to create VM directory")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		logger.Info("Received termination signal, shutting down VM")
		cancel()
	}()

	// Socket path for Firecracker
	socketPath := filepath.Join(vmDir, *socketName)

	// Create Firecracker configuration
	fcCfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: *kernelPath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off",
		Drives: []models.Drive{
			{
				DriveID:      firecracker.String("1"),
				PathOnHost:   firecracker.String(*rootfsPath),
				IsRootDevice: firecracker.Bool(true),
				IsReadOnly:   firecracker.Bool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(int64(*vcpuCount)),
			MemSizeMib: firecracker.Int64(int64(*memSize)),
		},
		VMID:        vmID,
		LogLevel:    "Info",
		LogFifo:     filepath.Join(vmDir, "firecracker.log"),
		MetricsFifo: filepath.Join(vmDir, "firecracker.metrics"),
	}

	// Add network configuration if not skipped
	if !*skipNetNs {
		fcCfg.NetworkInterfaces = firecracker.NetworkInterfaces{
			firecracker.NetworkInterface{
				// CNI network configuration
				CNIConfiguration: &firecracker.CNIConfiguration{
					NetworkName: *networkName, // matches the name in CNI config file
					IfName:      "veth0",      // changed from tap0 to veth0 for ptp plugin
				},
				AllowMMDS: true,
			},
		}
	} else {
		logger.Warn("Network namespace setup skipped. VM will not have network connectivity.")
	}

	// Create command for Firecracker
	cmd := firecracker.VMCommandBuilder{}.
		WithBin(*firecrackerBin).
		WithSocketPath(socketPath).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)

	// Create machine options
	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(logrus.NewEntry(logger)),
		firecracker.WithProcessRunner(cmd),
	}

	// Create the machine
	machine, err := firecracker.NewMachine(ctx, fcCfg, machineOpts...)
	if err != nil {
		// Provide helpful error message for common permission issues
		if os.IsPermission(err) || (err.Error() != "" &&
			(contains(err.Error(), "permission denied") ||
				contains(err.Error(), "operation not permitted"))) {
			logger.WithError(err).Fatal("Permission denied. Make sure you're running with sudo and check directory permissions.")
		}
		logger.WithError(err).Fatal("Failed to create machine")
	}

	// Start the machine
	startTime := time.Now()
	if err := machine.Start(ctx); err != nil {
		// Check for common network namespace errors
		if contains(err.Error(), "failed to initialize netns") ||
			contains(err.Error(), "permission denied") {
			logger.WithError(err).Fatal("Failed to setup network. This is likely a permissions issue. " +
				"Make sure you're running with sudo. You can also try --skip-netns flag to disable networking.")
		}
		logger.WithError(err).Fatal("Failed to start machine")
	}

	// Get the IP address from the network configuration
	var ipAddress string
	if len(machine.Cfg.NetworkInterfaces) > 0 && machine.Cfg.NetworkInterfaces[0].StaticConfiguration != nil {
		ipConfig := machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration
		if ipConfig != nil && ipConfig.IPAddr.IP != nil {
			ipAddress = ipConfig.IPAddr.IP.String()
		}
	}

	logger.WithFields(logrus.Fields{
		"vm_id":        vmID,
		"ip":           ipAddress,
		"socket_path":  socketPath,
		"start_time":   startTime.Format(time.RFC3339),
		"startup_time": time.Since(startTime).String(),
	}).Info("Machine started successfully")

	// Print helpful message about accessing the VM
	if ipAddress != "" {
		fmt.Printf("\n====================================================\n")
		fmt.Printf("VM started successfully with IP: %s\n", ipAddress)
		fmt.Printf("Socket path: %s\n", socketPath)
		fmt.Printf("====================================================\n\n")
	}

	logger.Info("VM is running. Press Ctrl+C to stop.")

	// Wait for context cancellation (via signal handler)
	<-ctx.Done()

	// Shutdown the VM
	logger.Info("Shutting down VM...")
	if err := machine.StopVMM(); err != nil {
		logger.WithError(err).Error("Failed to stop VM")
	} else {
		logger.Info("VM stopped successfully")
	}

	// Clean up VM directory if needed
	if err := os.RemoveAll(vmDir); err != nil {
		logger.WithError(err).Error("Failed to clean up VM directory")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && (s == substr || len(s) >= len(substr) && s[0:len(substr)] == substr)
}
