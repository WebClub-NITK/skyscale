# Sandbox Filesystem

Filesystem operations for sandboxes including file upload/download.

## Responsibilities

- **File Upload**: Upload files to sandbox
- **File Download**: Download files from sandbox
- **Directory Operations**: Create, list, delete directories
- **Quota Management**: Enforce storage quotas
- **Permissions**: Manage file permissions

## Operations

### Upload File
Upload file to sandbox workspace.

```go
err := fs.UploadFile(ctx, sandboxID, path, content, permissions)
```

Features:
- Base64 encoded content
- Automatic directory creation
- Permission control
- Size validation

### Download File
Download file from sandbox.

```go
content, err := fs.DownloadFile(ctx, sandboxID, path)
```

Features:
- Base64 encoded content
- Streaming for large files
- Path validation
- Access control

### List Directory
List files in directory.

```go
entries, err := fs.ListDirectory(ctx, sandboxID, path)
```

Returns:
- File names
- File sizes
- Permissions
- Modification times

### Create Directory
Create directory in sandbox.

```go
err := fs.CreateDirectory(ctx, sandboxID, path, permissions)
```

### Delete Path
Delete file or directory.

```go
err := fs.Delete(ctx, sandboxID, path, recursive)
```

## Workspace

Each sandbox has a workspace directory:
- Default: `/workspace`
- Persistent across executions
- Subject to quota limits
- Isolated from other sandboxes

## Quota Management

Storage quotas per sandbox:
- Default: 1 GB
- Maximum: 10 GB
- Configurable per sandbox
- Enforced on upload

## Security

- Path validation to prevent escaping workspace
- Permission enforcement
- Size limits
- Type restrictions

## Migration Status

🚧 **New Module** - To be implemented

See [Issue #31](../../ISSUES.md#issue-31-create-sandboxfs-module) for details.

## Future API

```go
// Upload file
err := fs.UploadFile(ctx, sandboxID, path, content, options)

// Download file
content, err := fs.DownloadFile(ctx, sandboxID, path)

// List directory
entries, err := fs.ListDirectory(ctx, sandboxID, path)

// Create directory
err := fs.CreateDirectory(ctx, sandboxID, path, permissions)

// Delete
err := fs.Delete(ctx, sandboxID, path, recursive)
```
