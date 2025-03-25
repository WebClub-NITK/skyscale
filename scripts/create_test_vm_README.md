# Firecracker Test VM Creator

This script creates a test VM using the Firecracker SDK for development and testing. It's a standalone tool that is independent of the main application code but uses the same Firecracker SDK.

## Requirements

- Go 1.19 or higher
- Firecracker binary installed at `/usr/local/bin/firecracker` (can be changed via flag)
- CNI plugins installed and configured
- Kernel and rootfs images (default locations are in the `../assets/` directory)
- Root privileges (sudo) for network namespace operations

## Installation

1. Install the required Go packages:

```bash
cd scripts
go mod init test-vm
go get github.com/firecracker-microvm/firecracker-go-sdk
go get github.com/google/uuid
go get github.com/sirupsen/logrus
```

## Usage

### Using the Helper Script (Recommended)

The easiest way to run the VM is to use the helper script which handles sudo permissions automatically:

```bash
./run_test_vm.sh
```

With custom options:

```bash
./run_test_vm.sh --kernel=/path/to/kernel --rootfs=/path/to/rootfs --cpu=2 --mem=256 --debug
```

The helper script automatically:
- Detects the Go executable path
- Preserves the PATH environment when switching to sudo
- Provides helpful error messages if Go is not found

### Running Directly

Run the script with default options (must use sudo):

```bash
sudo go run create_test_vm.go
```

Or customize the VM settings with command line flags:

```bash
sudo go run create_test_vm.go --kernel /path/to/kernel --rootfs /path/to/rootfs --cpu 2 --mem 256 --debug
```

**Note:** When running with `sudo` directly, you may need to specify the full path to Go or use `sudo -E` to preserve your environment variables if you encounter "go: command not found" errors.

## Available Options

- `--kernel`: Path to the kernel image (default: "../assets/vmlinux-6.1.129.bin")
- `--rootfs`: Path to the rootfs image (default: "../assets/ubuntu-22.04.ext4")
- `--cpu`: Number of vCPUs (default: 1)
- `--mem`: Memory size in MB (default: 128)
- `--fc`: Path to firecracker binary (default: "/usr/local/bin/firecracker")
- `--net`: CNI network name (default: "fcnet")
- `--socket`: Firecracker socket name (default: "firecracker.sock")
- `--debug`: Enable debug logging (default: false)
- `--dir`: Directory to store VM files (default: "/tmp/fc-test-vm")
- `--skip-netns`: Skip network namespace setup for debugging (default: false)

## Stopping the VM

The VM will run until you press Ctrl+C. It will then gracefully shut down and clean up the temporary files.

## Permissions Requirements

This script requires root privileges (sudo) because:

1. Creating and managing network namespaces requires root access
2. CNI plugins need to modify network interfaces
3. Some Firecracker operations require access to `/dev/kvm`

If you run without sudo, you will see permissions errors like:

```
Failed handler "fcinit.SetupNetwork": failed to initialize netns: failed to open new netns path at "/var/run/netns/...": permission denied
```

## Troubleshooting

### Permissions Issues

If you see permission errors:

1. Make sure you're running with `sudo`
2. Check that the firecracker binary is executable:
   ```bash
   sudo chmod +x /usr/local/bin/firecracker
   ```
3. Verify that your user has access to `/dev/kvm`:
   ```bash
   ls -la /dev/kvm
   # Add yourself to the appropriate group if needed
   sudo usermod -a -G kvm $(whoami)
   ```
4. If you just want to test without networking, use the `--skip-netns` flag:
   ```bash
   sudo go run create_test_vm.go --skip-netns
   ```

### CNI Issues

1. Make sure CNI plugins are installed in `/opt/cni/bin/`
2. Verify CNI configuration exists in `/etc/cni/conf.d/`
3. Check the contents of `/etc/cni/conf.d/fcnet.conflist` (or your custom network name)

### Firecracker Issues

1. Check if firecracker is installed correctly:
   ```bash
   which firecracker
   firecracker --version
   ```
2. Make sure kernel and rootfs paths are correct and files exist
3. Run with `--debug` flag to see more detailed logs 