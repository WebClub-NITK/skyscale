# Runtime Module

The `runtime/` module contains guest-side logic that runs **inside** the VM.

## Principle

**Clean separation between host and guest.**

The runtime module defines what runs inside VMs and how the host communicates with it.

## Submodules

### `runtime/agent/`
The Go agent that runs inside each VM:
- Receives requests from the control plane
- Manages the Python runtime
- Handles code execution
- Streams logs and results

### `runtime/python/`
Python runtime bootstrap and management:
- Python environment initialization
- Handler loading and execution
- Dependency management
- Error handling

### `runtime/protocol/`
Host ↔ guest communication protocol:
- Message definitions
- Encoding/decoding
- Version negotiation
- Request/response contracts

## Protocol is Key

The protocol defines:
- `Exec` - Execute code
- `UploadFile` - Transfer files
- `StreamLogs` - Stream execution logs
- `Shutdown` - Graceful shutdown

Once frozen, everything becomes composable.

## Usage

The runtime module is used by:
- VM module to bootstrap guest environment
- Control plane to execute functions
- Sandbox module for interactive sessions

## Status

🚧 **Under Development** - Agent exists in `cmd/daemon/`, being formalized as a module.

See [Issue #13-#19](../ISSUES.md) for tracking.
