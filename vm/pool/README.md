# VM Pool

The VM Pool maintains a collection of pre-warmed VMs to minimize cold start times.

## Responsibilities

- **Pool Maintenance**: Keep pool at target size
- **VM Acquisition**: Provide VMs on demand from pool
- **VM Return**: Accept VMs back into pool for reuse
- **Health Checking**: Remove unhealthy VMs from pool
- **Metrics**: Track pool usage and efficiency

## Key Concepts

### Pre-warmed VMs
VMs that are already booted and initialized, ready to accept workloads immediately.

### Pool Sizing
Dynamic or fixed pool sizes based on workload patterns.

### Eviction Policy
Strategy for removing VMs from pool (FIFO, LRU, health-based).

## Migration Status

🚧 **To Be Extracted** from `vm/manager/`

See [Issue #6](../../ISSUES.md#issue-6-create-vmpool-module) for details.

## Future API

```go
// Get a VM from the pool
vm, err := pool.Acquire(ctx)

// Return a VM to the pool
err := pool.Release(ctx, vm)

// Get pool statistics
stats := pool.Stats()
```
