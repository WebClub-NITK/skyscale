# Sandbox Exec

Code execution in sandboxes supporting shell and Python.

## Responsibilities

- **Shell Execution**: Execute shell commands
- **Python Execution**: Execute Python code
- **Output Streaming**: Stream stdout/stderr
- **Timeout Handling**: Enforce execution timeouts
- **Environment**: Manage execution environment

## Execution Types

### Shell Execution
Execute arbitrary shell commands.

```go
result, err := exec.Shell(ctx, sandboxID, "ls -la /workspace")
```

Features:
- Full shell access
- Environment variables
- Working directory control
- Signal handling

### Python Execution
Execute Python code or scripts.

```go
result, err := exec.Python(ctx, sandboxID, "print('hello')")
```

Features:
- Import installed packages
- Persistent state (globals)
- Exception handling
- Output capture

## Output Streaming

Real-time output streaming for long-running commands:

```go
stream, err := exec.StreamShell(ctx, sandboxID, "python train.py")
for {
    line, err := stream.ReadLine()
    if err == io.EOF {
        break
    }
    fmt.Println(line)
}
```

## Timeout Handling

Executions are subject to timeouts:
- Default timeout: 30 seconds
- Maximum timeout: 1 hour
- Configurable per request

## Environment Management

Control execution environment:
- Working directory
- Environment variables
- PATH configuration
- Resource limits

## Migration Status

🚧 **New Module** - To be implemented

See [Issue #30](../../ISSUES.md#issue-30-create-sandboxexec-module) for details.

## Future API

```go
// Execute shell command
result, err := exec.Shell(ctx, sandboxID, command, options)

// Execute Python code
result, err := exec.Python(ctx, sandboxID, code, options)

// Stream output
stream, err := exec.Stream(ctx, sandboxID, command)
```
