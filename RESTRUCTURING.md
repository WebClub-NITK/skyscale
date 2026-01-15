# Skyscale Repository Restructuring

## Overview

This document outlines the comprehensive restructuring plan for Skyscale to support both Lambda-style execution and E2B-style long-lived sandboxes while maintaining clean architectural boundaries.

## Core Design Principle

**Separate by responsibility, not by language or feature.**

The restructuring addresses the current interleaving of concerns:
- Function execution
- VM lifecycle management
- Control-plane orchestration logic
- Guest runtime

These will become **first-class modules** with clear ownership boundaries.

---

## Proposed Top-Level Structure

```
skyscale/
│
├── cmd/                     # Entry points only (thin)
│   ├── control-plane/       # Control plane daemon
│   ├── sandboxd/            # NEW: Sandbox-oriented daemon
│   └── cli/                 # CLI tool
│
├── control-plane/           # API + orchestration logic ONLY
│   ├── api/                 # REST/gRPC API handlers
│   ├── scheduler/           # Request routing and scheduling
│   ├── auth/                # Authentication and authorization
│   ├── state/               # State management
│   └── registry/            # Function registry
│
├── vm/                      # Firecracker & VM lifecycle (CRITICAL)
│   ├── manager/             # VM start/stop/lifecycle
│   ├── pool/                # Pre-warmed VM pool management
│   ├── snapshot/            # Snapshot create/restore
│   ├── network/             # TAP, vsock, bridges
│   └── jailer/              # Seccomp, UID isolation
│
├── runtime/                 # Guest-side logic (inside VM)
│   ├── agent/               # Go agent (runs inside VM)
│   ├── python/              # Python runtime bootstrap
│   └── protocol/            # Host ↔ guest communication contracts
│
├── sandbox/                 # NEW: E2B-style sandbox abstraction
│   ├── api/                 # Sandbox REST API handlers
│   ├── lifecycle/           # Create, suspend, resume, destroy
│   ├── exec/                # Shell + Python execution
│   └── fs/                  # Workspace, mounts, limits
│
├── sdk/                     # Client-facing SDKs
│   ├── python/              # Python SDK
│   └── types/               # Shared type definitions
│
├── assets/                  # VM assets
│   ├── kernels/             # Kernel images
│   ├── rootfs/              # Root filesystem images
│   └── snapshots/           # VM snapshots
│
├── internal/                # Shared internal utilities
│   ├── logging/             # Logging utilities
│   ├── config/              # Configuration management
│   └── errors/              # Error handling
│
├── perf/                    # Performance testing
├── scripts/                 # Build and utility scripts
├── examples/                # Example functions
└── docs/                    # Documentation
```

---

## Why This Structure Works

### 1. `vm/` becomes sacred ground

Everything Firecracker-related lives here. This is the single source of truth for:
- Starting/stopping microVMs
- Managing VM pools
- Creating/restoring snapshots
- Network configuration
- Security isolation (jailer)

**Critical Rule:** Nothing outside `vm/` talks to Firecracker directly.

This enables:
- Lambda execution
- Long-lived sandboxes
- REPL environments
- Testing
...to all reuse the same VM machinery.

### 2. `runtime/` cleanly separates host vs guest

The Go agent and Python runtime are formalized as first-class components.

#### Protocol is Key

Define clear contracts for:
- `Exec` - Execute code
- `UploadFile` - Transfer files
- `StreamLogs` - Stream execution logs
- `Shutdown` - Graceful shutdown

Once frozen, everything becomes composable.

### 3. `sandbox/` is a new first-class product

This enables E2B-style sandboxes **without touching Lambda logic**.

A sandbox is simply:
> A long-lived VM + relaxed execution semantics

Same VM manager, different policy.

### 4. `control-plane/` stops doing too much

It becomes orchestration-only:
- Request routing
- Authentication
- State management
- Function registry

**No Firecracker code. No runtime logic.**

### 5. `cmd/` stays thin (non-negotiable)

Each entry point:
- Loads config
- Wires dependencies
- Starts services

**No business logic in main.go**

### 6. SDKs live outside infrastructure

The Python SDK will support both:

```python
# Lambda-style
invoke("fn", payload)

# Sandbox-style
sb = Sandbox()
sb.exec("pip install torch")
sb.exec("python train.py")
```

Same backend. Different UX.

---

## Migration Strategy (Low Risk)

### Step 1: Extract VM Logic
- Move Firecracker code → `vm/`
- No behavior change
- Update imports

### Step 2: Formalize Guest Protocol
- Lock host ↔ agent contracts
- Document protocol in `runtime/protocol/`

### Step 3: Introduce `sandbox/` Behind Feature Flag
- Reuse VM pool, snapshots, agent
- Implement sandbox API
- Keep disabled by default

### Step 4: SDK-First Development
- Design Python SDK before API changes
- Validate developer experience

### Step 5: Reorganize Control Plane
- Move entry points to `cmd/`
- Remove VM dependencies
- Use `vm/` module exclusively

### Step 6: Add Internal Utilities
- Extract common utilities to `internal/`
- Improve code reuse

---

## What This Unlocks

### Immediately
- REPL-like sandboxes
- Stateful execution
- Better cold-start control
- Cleaner code boundaries

### Soon
- Notebook-style workflows
- CI sandboxes
- ML training jobs
- Dev environments

### Long-term
**Skyscale ≠ Lambda clone**
**Skyscale = programmable compute substrate**

---

## Implementation Phases

See [MILESTONES.md](MILESTONES.md) for detailed milestone breakdown and tracking.

---

## Notes

- All changes should be incremental and testable
- Each phase should maintain backward compatibility
- Feature flags will protect experimental features
- Comprehensive testing at each phase

---

## Related Documents

- [MILESTONES.md](MILESTONES.md) - Detailed milestone tracking
- [ISSUES.md](ISSUES.md) - Individual issue tracking
- [README.md](README.md) - Project overview
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
