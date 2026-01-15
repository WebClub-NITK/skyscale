# Runtime Agent

The Go agent runs inside each Firecracker VM and handles communication with the host.

## Responsibilities

- **Request Handling**: Receive and process host requests
- **Runtime Management**: Start and manage Python runtime
- **Code Execution**: Execute function code
- **Log Streaming**: Stream execution logs to host
- **File Operations**: Handle file uploads/downloads
- **Health Reporting**: Report VM health status

## Key Features

- Lightweight and fast
- Single binary deployment
- Minimal dependencies
- Efficient communication via vsock or network

## Current State

The agent currently exists in `cmd/daemon/` and will be formalized as a reusable module.

## Migration Status

🚧 **To Be Formalized** from `cmd/daemon/`

See [Issue #13](../../ISSUES.md#issue-13-create-runtimeagent-module) for details.

## Future API

```go
// Start the agent
agent := runtime.NewAgent(config)
err := agent.Start(ctx)

// Handle incoming request
response, err := agent.HandleRequest(ctx, request)

// Execute code
result, err := agent.Execute(ctx, code, runtime)
```

## Deployment

The agent binary is copied into the VM rootfs and started on boot.
