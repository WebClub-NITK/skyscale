# VM Network

The VM Network module handles network configuration for Firecracker VMs.

## Responsibilities

- **TAP Devices**: Create and configure TAP devices
- **Bridge Setup**: Configure network bridges
- **IP Allocation**: Assign and manage IP addresses
- **vsock**: Configure vsock for host-guest communication
- **Network Isolation**: Ensure network isolation between VMs

## Key Concepts

### TAP Devices
Virtual network interfaces attached to VMs.

### CNI Integration
Integration with Container Network Interface for advanced networking.

### vsock
Virtual socket for efficient host-guest communication.

## Migration Status

🚧 **To Be Extracted** from `vm/manager/`

See [Issue #8](../../ISSUES.md#issue-8-create-vmnetwork-module) for details.

## Future API

```go
// Configure network for VM
netConfig, err := network.Configure(ctx, vmID, options)

// Allocate IP address
ip, err := network.AllocateIP(ctx)

// Setup vsock
vsockConfig, err := network.SetupVsock(ctx, vmID)
```
