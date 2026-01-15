# SDK Types

Shared type definitions for SDKs.

## Purpose

Common types used across SDK implementations to ensure consistency.

## Type Categories

### Request Types
- Function invocation requests
- Sandbox creation requests
- File operation requests

### Response Types
- Function invocation results
- Sandbox information
- Execution results
- File metadata

### Configuration Types
- Client configuration
- Timeout settings
- Resource limits

### Error Types
- Error codes
- Error messages
- Error categories

## Migration Status

🚧 **To Be Implemented**

See [Issue #37](../../ISSUES.md#issue-37-create-sdkpython-module) for details.

## Future Structure

```
sdk/types/
├── requests.go      # Request type definitions
├── responses.go     # Response type definitions
├── errors.go        # Error types
└── config.go        # Configuration types
```

## Example Types

```go
// FunctionInvokeRequest
type FunctionInvokeRequest struct {
    Name    string                 `json:"name"`
    Payload map[string]interface{} `json:"payload"`
    Sync    bool                   `json:"sync"`
    Timeout int                    `json:"timeout"`
}

// SandboxCreateRequest
type SandboxCreateRequest struct {
    Runtime        string `json:"runtime"`
    MemoryMB       int    `json:"memory_mb"`
    CPUCount       int    `json:"cpu_count"`
    TimeoutSeconds int    `json:"timeout_seconds"`
}

// ExecResult
type ExecResult struct {
    Status     string `json:"status"`
    Stdout     string `json:"stdout"`
    Stderr     string `json:"stderr"`
    ExitCode   int    `json:"exit_code"`
    DurationMs int    `json:"duration_ms"`
}
```
