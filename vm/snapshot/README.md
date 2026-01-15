# VM Snapshot

The VM Snapshot module handles creating and restoring VM snapshots for fast initialization.

## Responsibilities

- **Snapshot Creation**: Create memory and disk snapshots
- **Snapshot Restoration**: Restore VMs from snapshots
- **Metadata Management**: Track snapshot metadata
- **Storage**: Manage snapshot storage locations
- **Layering**: Support snapshot layering for efficiency

## Key Concepts

### Full Snapshots
Complete VM state including memory and disk.

### Incremental Snapshots
Only changes since last snapshot (future).

### Snapshot Layering
Base snapshot + deltas for different configurations.

## Use Cases

1. **Fast Cold Starts**: Pre-initialized Python runtime
2. **Different Configurations**: Base + function-specific layers
3. **Version Management**: Snapshot different runtime versions

## Migration Status

🚧 **New Module** - To be implemented

See [Issue #7](../../ISSUES.md#issue-7-create-vmsnapshot-module) for details.

## Future API

```go
// Create a snapshot from running VM
snapshot, err := snapshotMgr.Create(ctx, vmID, metadata)

// Restore VM from snapshot
vm, err := snapshotMgr.Restore(ctx, snapshotID)

// List available snapshots
snapshots, err := snapshotMgr.List(ctx)
```
