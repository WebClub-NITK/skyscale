# VM Module

The `vm/` module is the **sacred ground** for all Firecracker-related operations. This module provides the single source of truth for VM lifecycle management.

## Principle

**Nothing outside `vm/` talks to Firecracker directly.**

This isolation enables Lambda execution, sandboxes, REPLs, and tests to all reuse the same VM machinery with different policies.

## Submodules

### `vm/manager/`
Core VM lifecycle management:
- Starting and stopping microVMs
- VM instance tracking
- Resource allocation
- VM configuration

### `vm/pool/`
Pre-warmed VM pool management:
- Maintaining pool of ready VMs
- Pool size management
- VM reuse strategies
- Cold start optimization

### `vm/snapshot/`
VM snapshot operations:
- Creating VM snapshots
- Restoring from snapshots
- Snapshot metadata management
- Fast VM initialization

### `vm/network/`
Network configuration and management:
- TAP device setup
- Bridge configuration
- IP allocation
- vsock communication

### `vm/jailer/`
Security isolation and containment:
- Seccomp configuration
- UID isolation
- Chroot setup
- Resource limits

## Usage

The VM module is used by:
- Control plane for Lambda-style execution
- Sandbox module for long-lived sessions
- Testing framework for isolated tests

## Status

🚧 **Under Development** - This module is being extracted from `control-plane/vm/` as part of the restructuring effort.

See [Issue #5-#12](../ISSUES.md) for tracking.
