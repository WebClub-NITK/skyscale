# Skyscale Architecture

## Overview

Skyscale is a serverless platform that supports both Lambda-style function execution and E2B-style long-lived sandboxes, built on Firecracker microVMs.

## Core Principles

1. **Separation by Responsibility**: Modules organized by function, not by language or feature
2. **Single Source of Truth**: VM operations only in `vm/` module
3. **Clean Boundaries**: Clear interfaces between modules
4. **Feature Flexibility**: Support multiple execution models with shared infrastructure

## Module Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Entry Points                         │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  control-    │  │              │  │     CLI      │     │
│  │   plane      │  │  sandboxd    │  │    Tool      │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                    Control Plane Layer                       │
│                                                              │
│  ┌─────────┐  ┌──────────┐  ┌──────┐  ┌───────┐           │
│  │   API   │  │Scheduler │  │ Auth │  │ State │           │
│  └─────────┘  └──────────┘  └──────┘  └───────┘           │
└──────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                    Execution Layer                           │
│                                                              │
│  ┌─────────────────────┐         ┌──────────────────────┐  │
│  │    VM Module        │         │   Sandbox Module     │  │
│  │                     │         │                      │  │
│  │  • Manager          │         │  • API               │  │
│  │  • Pool             │         │  • Lifecycle         │  │
│  │  • Snapshot         │         │  • Exec              │  │
│  │  • Network          │         │  • Filesystem        │  │
│  │  • Jailer           │         │                      │  │
│  └─────────────────────┘         └──────────────────────┘  │
└──────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                    Runtime Layer                             │
│                   (Inside VMs)                              │
│                                                              │
│  ┌─────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Agent  │  │    Python    │  │  Protocol    │          │
│  │  (Go)   │  │   Runtime    │  │              │          │
│  └─────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                      Client Layer                             │
│                                                              │
│                    ┌────────────┐                            │
│                    │ Python SDK │                            │
│                    └────────────┘                            │
└──────────────────────────────────────────────────────────────┘
```

## Data Flow

### Lambda-Style Invocation

```
Client → API → Scheduler → VM Manager → VM Pool → Agent → Python → Result
```

1. Client sends function invocation request
2. API validates and authenticates
3. Scheduler selects VM
4. VM Manager provides VM from pool
5. Agent receives request via protocol
6. Python runtime executes handler
7. Result streams back through chain

### Sandbox-Style Session

```
Client → Sandbox API → Sandbox Lifecycle → VM Manager → VM → Agent → Python
   ↓                                                                    ↑
   └────────────────────── Multiple Exec Requests ───────────────────┘
```

1. Client creates sandbox
2. Sandbox Lifecycle provisions VM
3. VM persists in "ready" state
4. Client sends multiple exec requests
5. Each execution uses same VM
6. Client destroys sandbox when done

## Module Responsibilities

### Entry Points (`cmd/`)
- Parse configuration
- Wire dependencies
- Start services
- **No business logic**

### Control Plane
- **API**: Handle HTTP requests
- **Scheduler**: Route invocations to VMs
- **Auth**: Verify API keys
- **State**: Persist metadata
- **Registry**: Manage functions

### VM Module
- **Manager**: VM lifecycle (create, stop, terminate)
- **Pool**: Pre-warmed VM pool
- **Snapshot**: Snapshot create/restore
- **Network**: Network configuration
- **Jailer**: Security isolation

### Runtime Module
- **Agent**: Host-guest communication (inside VM)
- **Python**: Handler execution (inside VM)
- **Protocol**: Message definitions

### Sandbox Module
- **API**: Sandbox REST endpoints
- **Lifecycle**: Create, suspend, resume, destroy
- **Exec**: Code execution
- **Filesystem**: File operations

### SDK
- **Python**: Client library for both styles

## Communication

### Host ↔ Guest Protocol

Defined in `runtime/protocol/`:

```
Host (Control Plane/Sandbox)
  ↕ vsock or network
Guest (Agent)
  ↕ protocol messages
Runtime (Python)
```

Messages:
- `Exec`: Execute code
- `UploadFile`: Transfer file
- `StreamLogs`: Stream output
- `Shutdown`: Graceful stop

### API Communication

```
Client (SDK)
  ↕ HTTPS
Control Plane (API)
  ↕ Internal calls
VM/Sandbox Modules
```

## Security Model

### Isolation Layers

1. **Firecracker microVM**: Hardware-virtualized isolation
2. **Jailer**: Additional seccomp, namespace, chroot isolation
3. **Network**: Isolated network per VM
4. **Filesystem**: Isolated workspace per sandbox

### Authentication

- API key authentication
- Per-function/sandbox access control
- Rate limiting
- Audit logging

## State Management

### Persistent State (SQLite/Postgres)
- Function metadata
- Sandbox metadata
- VM assignments
- Execution history

### Transient State (Memory)
- VM pool
- Active sandboxes
- In-flight requests

## Scaling

### Horizontal Scaling
- Multiple control plane instances
- Load balancer for API requests
- Shared state via database

### Vertical Scaling
- More VMs per host
- Larger VM pool
- More memory/CPU per VM

## Performance Optimizations

### Cold Start Reduction
- Pre-warmed VM pool
- Snapshot-based initialization
- Concurrent pool filling

### Execution Efficiency
- Keep-alive for sandboxes
- Efficient protocol (vsock)
- Streaming output

### Resource Management
- CPU/memory limits per VM
- Automatic scaling of pool
- Idle timeout and cleanup

## Future Enhancements

1. **Multi-region**: Deploy across regions
2. **Custom runtimes**: Support more languages
3. **GPU support**: Add GPU-enabled VMs
4. **Distributed tracing**: Full request tracing
5. **Auto-scaling**: Dynamic pool sizing

## Related Documents

- [RESTRUCTURING.md](../RESTRUCTURING.md) - Restructuring plan
- [MILESTONES.md](../MILESTONES.md) - Implementation roadmap
- [ISSUES.md](../ISSUES.md) - Detailed issues

---

Last Updated: 2026-01-15
