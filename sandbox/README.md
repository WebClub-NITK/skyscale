# Sandbox Module

The `sandbox/` module provides E2B-style long-lived sandbox functionality.

## Principle

**A sandbox is just a long-lived VM + relaxed execution semantics.**

Same VM manager, different policy. No need to touch Lambda logic.

## What is a Sandbox?

A sandbox is a persistent, interactive execution environment that:
- Lives longer than a single function invocation
- Supports multiple executions in sequence
- Maintains state between executions
- Allows file uploads and downloads
- Can be suspended and resumed

## Submodules

### `sandbox/api/`
REST API handlers for sandbox operations:
- Create sandbox
- Execute code
- Upload/download files
- List sandboxes
- Delete sandbox

### `sandbox/lifecycle/`
Sandbox lifecycle management:
- Create: Provision a new sandbox
- Suspend: Pause sandbox and save state
- Resume: Restore suspended sandbox
- Destroy: Clean up sandbox resources

### `sandbox/exec/`
Code execution in sandboxes:
- Shell command execution
- Python code execution
- Output streaming
- Timeout handling

### `sandbox/fs/`
Filesystem operations:
- File upload
- File download
- Directory operations
- Quota management

## Use Cases

1. **Interactive Development**: REPL-like environments
2. **Notebook Workflows**: Jupyter-style execution
3. **CI/CD**: Isolated build environments
4. **ML Training**: Long-running training jobs
5. **Dev Environments**: Cloud-based development

## Comparison: Lambda vs Sandbox

| Feature | Lambda | Sandbox |
|---------|--------|---------|
| Lifetime | Single invocation | Persistent |
| State | Stateless | Stateful |
| Execution | One function call | Multiple executions |
| Files | Read-only code | Read/write workspace |
| Use Case | Event processing | Interactive work |

## Status

🚧 **New Feature** - To be implemented behind feature flag

See [Issue #27-#35](../ISSUES.md) for tracking.

## Example Usage

```python
# Create sandbox
sandbox = client.sandboxes.create()

# Execute code
result = sandbox.exec("pip install requests")
result = sandbox.exec("python train.py")

# Upload files
sandbox.upload_file("data.csv", content)

# Download results
results = sandbox.download_file("output.txt")

# Clean up
sandbox.destroy()
```
