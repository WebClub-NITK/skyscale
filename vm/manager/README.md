# VM Manager

The VM Manager handles the core lifecycle of Firecracker microVMs.

## Responsibilities

- **VM Creation**: Initialize and start new Firecracker VMs
- **VM Termination**: Stop and clean up VMs
- **Resource Management**: Track CPU, memory, and other resources
- **Configuration**: Apply VM configurations (kernel, rootfs, network, etc.)
- **State Tracking**: Monitor VM status and health

## Key Types

- `VMManager`: Main manager coordinating VM operations
- `VMInstance`: Represents a running VM
- `VMConfig`: Configuration for VM creation

## Migration Status

🚧 **To Be Migrated** from `control-plane/vm/`

See [Issue #5](../../ISSUES.md#issue-5-create-vmmanager-module) for details.

## Future API

```go
// Create a new VM with configuration
vm, err := manager.Create(ctx, config)

// Stop and clean up a VM
err := manager.Terminate(ctx, vmID)

// Get VM status
status, err := manager.GetStatus(ctx, vmID)
```
