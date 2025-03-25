#!/bin/bash
# Helper script to run the Firecracker test VM creator with the correct permissions

# Find the Go executable path before we switch to root
GO_BIN=$(which go 2>/dev/null)
if [ -z "$GO_BIN" ]; then
  if [ -x "/usr/local/go/bin/go" ]; then
    GO_BIN="/usr/local/go/bin/go"
  elif [ -x "/usr/bin/go" ]; then
    GO_BIN="/usr/bin/go"
  else
    echo "Error: Go executable not found. Please make sure Go is installed and in your PATH."
    exit 1
  fi
fi

# Remember the original PATH
ORIGINAL_PATH="$PATH"

# Check if we're running as root
if [ "$EUID" -ne 0 ]; then
  echo "This script requires root privileges. Running with sudo..."
  # Pass the environment variables to sudo
  exec sudo PATH="$ORIGINAL_PATH" "$0" "$@"
  exit $?
fi

# Directory of this script
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

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

# Verify the Go executable is available
if ! command -v "$GO_BIN" &> /dev/null; then
  echo "Error: Go executable not found at $GO_BIN"
  echo "Paths searched in PATH: $PATH"
  echo "You may need to install Go or update your PATH environment variable."
  exit 1
fi

echo "Using Go executable: $GO_BIN"

# Build command
CMD="cd $SCRIPT_DIR && $GO_BIN run create_test_vm.go"
CMD="$CMD --kernel=$KERNEL_PATH --rootfs=$ROOTFS_PATH --cpu=$CPU_COUNT --mem=$MEM_SIZE"

if [ "$DEBUG" = true ]; then
  CMD="$CMD --debug"
fi

if [ "$SKIP_NETNS" = true ]; then
  CMD="$CMD --skip-netns"
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

# Execute the command
eval $CMD