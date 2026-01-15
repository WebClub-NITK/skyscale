# SDK Module

Client-facing SDKs for interacting with Skyscale.

## Principle

**SDK-first development: Design the developer experience before the API.**

The SDK provides both Lambda-style and Sandbox-style interfaces using the same backend.

## Submodules

### `sdk/python/`
Python SDK for Skyscale:
- Lambda-style function invocation
- Sandbox-style interactive execution
- File operations
- Authentication
- Error handling

### `sdk/types/`
Shared type definitions:
- Request/response types
- Configuration types
- Error types
- Common utilities

## Dual API Design

### Lambda-Style

For stateless function execution:

```python
from skyscale import SkyscaleClient

client = SkyscaleClient(api_key="...")

# Invoke function
result = client.functions.invoke(
    name="my-function",
    payload={"key": "value"}
)
```

### Sandbox-Style

For interactive, stateful execution:

```python
from skyscale import SkyscaleClient

client = SkyscaleClient(api_key="...")

# Create sandbox
sandbox = client.sandboxes.create(
    runtime="python3.8",
    memory_mb=512
)

# Execute code
result = sandbox.exec("pip install requests")
result = sandbox.exec("python train.py")

# Upload files
sandbox.upload_file("data.csv", content)

# Download results
output = sandbox.download_file("output.txt")

# Clean up
sandbox.destroy()
```

## Key Features

- **Simple API**: Pythonic, intuitive interface
- **Type Hints**: Full type annotations
- **Async Support**: Asyncio-compatible
- **Error Handling**: Clear, actionable errors
- **Documentation**: Comprehensive docs and examples

## Status

🚧 **Under Development** - To be implemented

See [Issue #36-#43](../ISSUES.md) for tracking.

## Installation

```bash
pip install skyscale
```

## Authentication

```python
# From environment variable
client = SkyscaleClient()  # uses SKYSCALE_API_KEY

# Explicit API key
client = SkyscaleClient(api_key="sk_...")

# From config file
client = SkyscaleClient.from_config("~/.skyscale/config")
```
