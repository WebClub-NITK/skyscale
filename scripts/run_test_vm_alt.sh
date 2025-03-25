#!/bin/bash
# Alternative helper script to run the Firecracker test VM creator
# This version builds the Go program first, then runs the binary with sudo

# Directory of this script
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$SCRIPT_DIR"

# Default options
KERNEL_PATH="../assets/vmlinux-6.1.129.bin"
ROOTFS_PATH="../assets/ubuntu-22.04.ext4"
CPU_COUNT=1
MEM_SIZE=128
DEBUG=false
SKIP_NETNS=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --kernel=*)
      KERNEL_PATH="${1#*=}"
      shift
      ;;
    --rootfs=*)
      ROOTFS_PATH="${1#*=}"
      shift
      ;;
    --cpu=*)
      CPU_COUNT="${1#*=}"
      shift
      ;;
    --mem=*)
      MEM_SIZE="${1#*=}"
      shift
      ;;
    --debug)
      DEBUG=true
      shift
      ;;
    --skip-netns)
      SKIP_NETNS=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--kernel=PATH] [--rootfs=PATH] [--cpu=N] [--mem=N] [--debug] [--skip-netns]"
      exit 1
      ;;
  esac
done

# Build the Go program first
echo "Building test VM creator..."
go build -o test_vm_creator create_test_vm.go

if [ $? -ne 0 ]; then
  echo "Error: Failed to build the Go program. Make sure Go is installed correctly."
  exit 1
fi

# Build command arguments
ARGS="--kernel=$KERNEL_PATH --rootfs=$ROOTFS_PATH --cpu=$CPU_COUNT --mem=$MEM_SIZE"

if [ "$DEBUG" = true ]; then
  ARGS="$ARGS --debug"
fi

if [ "$SKIP_NETNS" = true ]; then
  ARGS="$ARGS --skip-netns"
fi

echo "Starting Firecracker test VM with the following configuration:"
echo "  Kernel: $KERNEL_PATH"
echo "  Rootfs: $ROOTFS_PATH"
echo "  CPU: $CPU_COUNT"
echo "  Memory: $MEM_SIZE MB"
echo "  Debug: $DEBUG"
echo "  Skip Network Namespace: $SKIP_NETNS"
echo ""
echo "Press Ctrl+C to stop the VM"
echo "================================"

# Execute the program with sudo
sudo ./test_vm_creator $ARGS 