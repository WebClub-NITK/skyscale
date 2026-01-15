# Sandbox API

REST API handlers for sandbox operations.

## Responsibilities

- **Request Handling**: Handle HTTP requests for sandbox operations
- **Validation**: Validate request parameters
- **Error Handling**: Return appropriate error responses
- **Authentication**: Verify API keys
- **Documentation**: OpenAPI/Swagger specs

## Endpoints

### POST /sandboxes
Create a new sandbox.

```
Request:
{
  "runtime": "python3.8",
  "memory_mb": 512,
  "cpu_count": 1,
  "timeout_seconds": 3600
}

Response:
{
  "id": "sb_abc123",
  "status": "ready",
  "ip": "172.16.0.5",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### POST /sandboxes/{id}/exec
Execute code in sandbox.

```
Request:
{
  "language": "python",
  "code": "print('hello')",
  "timeout": 30
}

Response:
{
  "status": "success",
  "stdout": "hello\n",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 50
}
```

### POST /sandboxes/{id}/files
Upload file to sandbox.

```
Request:
{
  "path": "/workspace/data.txt",
  "content": "base64...",
  "permissions": "644"
}

Response:
{
  "path": "/workspace/data.txt",
  "size": 1024
}
```

### GET /sandboxes/{id}/files/{path}
Download file from sandbox.

```
Response:
{
  "path": "/workspace/output.txt",
  "content": "base64...",
  "size": 2048
}
```

### GET /sandboxes
List sandboxes.

```
Response:
{
  "sandboxes": [
    {
      "id": "sb_abc123",
      "status": "ready",
      "created_at": "..."
    }
  ]
}
```

### DELETE /sandboxes/{id}
Destroy sandbox.

```
Response:
{
  "status": "deleted"
}
```

## Migration Status

🚧 **New Module** - To be implemented

See [Issue #28](../../ISSUES.md#issue-28-create-sandboxapi-module) for details.
