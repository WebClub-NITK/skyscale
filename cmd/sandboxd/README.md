# Sandboxd

Sandbox-oriented daemon for managing long-lived execution environments.

## Purpose

A specialized daemon for sandbox operations, separate from the control plane.

## Responsibilities

- Manage sandbox lifecycle
- Handle sandbox API requests
- Monitor sandbox health
- Enforce resource limits
- Clean up idle sandboxes

## Comparison to Control Plane

| Feature | Control Plane | Sandboxd |
|---------|---------------|----------|
| Focus | Function invocation | Long-lived sandboxes |
| Lifetime | Seconds | Minutes to hours |
| State | Stateless | Stateful |
| API | Function API | Sandbox API |

## Migration Status

🚧 **Future Enhancement** - Optional separate daemon

May be implemented later to separate concerns or can be integrated into control plane.

## Configuration

```yaml
sandboxd:
  port: 8081
  max_sandboxes: 100
  idle_timeout: 3600
  cleanup_interval: 300
```

## Running

```bash
./sandboxd --config sandboxd.yaml
```
