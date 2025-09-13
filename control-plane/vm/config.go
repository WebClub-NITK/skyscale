package vm

import (
	"os"
	"path/filepath"
	"strconv"
)

// Environment variable names
const (
	EnvVMKernelPath = "FAAS_VM_KERNEL_PATH"
	EnvVMRootFSPath = "FAAS_VM_ROOTFS_PATH"
	EnvVMMemoryMB   = "FAAS_VM_MEMORY_MB"
	EnvVMCPUCount   = "FAAS_VM_CPU_COUNT"
)

// getDefaultKernelPath returns the default kernel path
func getDefaultKernelPath() string {
	// Check environment variable first
	if path := os.Getenv(EnvVMKernelPath); path != "" {
		return path
	}
	// Default to the hardcoded path
	return filepath.Join("/home", "bluequbit", "Dev", "faas", "assets", "vmlinux-5.10.225")
}

// getDefaultRootFSPath returns the default rootfs path
func getDefaultRootFSPath() string {
	// Check environment variable first
	if path := os.Getenv(EnvVMRootFSPath); path != "" {
		return path
	}
	// Default to the hardcoded path
	return filepath.Join("/home", "bluequbit", "Dev", "faas", "scripts", "rootfs.ext4")
}

// getDefaultMemoryMB returns the default memory in MB
func getDefaultMemoryMB() int {
	// Check environment variable first
	if mem := os.Getenv(EnvVMMemoryMB); mem != "" {
		if val, err := strconv.Atoi(mem); err == nil && val > 0 {
			return val
		}
	}
	// Default to 128MB
	return 128
}

// getDefaultCPUCount returns the default CPU count
func getDefaultCPUCount() int {
	// Check environment variable first
	if cpu := os.Getenv(EnvVMCPUCount); cpu != "" {
		if val, err := strconv.Atoi(cpu); err == nil && val > 0 {
			return val
		}
	}
	// Default to 1 CPU
	return 1
}
