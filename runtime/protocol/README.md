# Runtime Protocol

The protocol module defines the communication contract between host and guest.

## Responsibilities

- **Message Definitions**: Define all message types
- **Encoding/Decoding**: Handle message serialization
- **Version Negotiation**: Support protocol versioning
- **Error Handling**: Standard error responses
- **Documentation**: Protocol specification

## Protocol Messages

### Core Operations

#### Exec
Execute code in the VM.

```
Request:
{
  "type": "exec",
  "runtime": "python3.8",
  "code": "...",
  "event": {...},
  "timeout": 30
}

Response:
{
  "status": "success",
  "result": {...},
  "logs": "...",
  "duration_ms": 150
}
```

#### UploadFile
Transfer file to VM.

```
Request:
{
  "type": "upload_file",
  "path": "/tmp/data.txt",
  "content": "base64...",
  "permissions": "644"
}

Response:
{
  "status": "success",
  "path": "/tmp/data.txt"
}
```

#### StreamLogs
Stream execution logs.

```
Request:
{
  "type": "stream_logs",
  "follow": true
}

Response (streaming):
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "..."
}
```

#### Shutdown
Graceful shutdown.

```
Request:
{
  "type": "shutdown",
  "timeout": 10
}

Response:
{
  "status": "success"
}
```

## Version Negotiation

Protocol versions follow semantic versioning.

```
Handshake:
{
  "protocol_version": "1.0.0",
  "capabilities": ["exec", "upload_file", "stream_logs"]
}
```

## Migration Status

🚧 **To Be Defined** - Protocol currently implicit, needs formalization

See [Issue #15-#17](../../ISSUES.md) for details.

## Future Structure

```
runtime/protocol/
├── messages.go      # Message type definitions
├── codec.go         # Encoding/decoding
├── version.go       # Version negotiation
└── PROTOCOL.md      # Protocol specification
```
