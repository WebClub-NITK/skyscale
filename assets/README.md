# Assets

VM assets including kernels, root filesystems, and snapshots.

## Structure

### `assets/kernels/`
Linux kernel images for Firecracker VMs:
- vmlinux-5.10 (default)
- vmlinux-6.1 (newer)
- Custom kernels

### `assets/rootfs/`
Root filesystem images:
- base-python3.8.ext4
- base-python3.9.ext4
- base-python3.10.ext4

### `assets/snapshots/`
VM snapshots for fast initialization:
- Organized by runtime
- Versioned
- Metadata included

## Building Assets

### Kernel

See [scripts/build-kernel.sh](../scripts/build-kernel.sh)

### Root Filesystem

See [scripts/build-rootfs.sh](../scripts/build-rootfs.sh)

### Snapshots

Snapshots are created automatically by the VM manager.

## Configuration

Asset paths are configured via environment variables:

```bash
FAAS_VM_KERNEL_PATH=/path/to/vmlinux
FAAS_VM_ROOTFS_PATH=/path/to/rootfs.ext4
FAAS_SNAPSHOT_DIR=/path/to/snapshots
```

## Migration Status

🚧 **To Be Organized** - Assets currently in various locations

See [Issue #47-#48](../ISSUES.md) for details.

## Best Practices

- Keep assets immutable
- Version assets clearly
- Document asset requirements
- Automate asset builds
