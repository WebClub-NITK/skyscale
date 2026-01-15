# Sandbox Lifecycle

Sandbox lifecycle management for create, suspend, resume, and destroy operations.

## Responsibilities

- **Creation**: Provision new sandboxes from VMs
- **Suspension**: Pause sandbox and save state
- **Resumption**: Restore suspended sandbox
- **Destruction**: Clean up sandbox resources
- **State Tracking**: Track sandbox lifecycle state

## Lifecycle States

```
        create
   ┌──────────────┐
   │              ▼
   │         ┌─────────┐
   │         │ READY   │◄─┐
   │         └─────────┘  │
   │              │       │
   │         exec │       │ resume
   │              ▼       │
   │         ┌─────────┐  │
   │         │ BUSY    │──┘
   │         └─────────┘
   │              │
   │      suspend │
   │              ▼
   │         ┌──────────┐
   │         │SUSPENDED │
   │         └──────────┘
   │              │
   └──────────────┼──────────┐
              destroy        │
                  ▼          │
              ┌─────────┐    │
              │DESTROYED│◄───┘
              └─────────┘
```

## Operations

### Create
1. Acquire VM from pool or create new
2. Initialize sandbox workspace
3. Set up filesystem
4. Register in state manager
5. Return sandbox ID

### Suspend
1. Save VM state to snapshot
2. Stop VM
3. Update state to SUSPENDED
4. Free resources

### Resume
1. Restore VM from snapshot
2. Restore workspace
3. Update state to READY
4. Return connection info

### Destroy
1. Stop VM if running
2. Clean up workspace
3. Delete snapshots
4. Remove from state manager
5. Free resources

## Migration Status

🚧 **New Module** - To be implemented

See [Issue #29](../../ISSUES.md#issue-29-create-sandboxlifecycle-module) for details.

## Future API

```go
// Create new sandbox
sandbox, err := lifecycle.Create(ctx, config)

// Suspend sandbox
err := lifecycle.Suspend(ctx, sandboxID)

// Resume sandbox
sandbox, err := lifecycle.Resume(ctx, sandboxID)

// Destroy sandbox
err := lifecycle.Destroy(ctx, sandboxID)
```
