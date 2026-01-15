# Skyscale Python SDK

Official Python SDK for Skyscale.

## Installation

```bash
pip install skyscale
```

## Quick Start

### Lambda-Style Functions

```python
from skyscale import SkyscaleClient

client = SkyscaleClient(api_key="sk_...")

# Invoke function
result = client.functions.invoke("my-function", {"key": "value"})
print(result)
```

### Sandbox-Style Execution

```python
from skyscale import SkyscaleClient

client = SkyscaleClient(api_key="sk_...")

# Create and use sandbox
with client.sandboxes.create() as sandbox:
    # Install dependencies
    sandbox.exec("pip install pandas numpy")
    
    # Upload data
    sandbox.upload_file("data.csv", open("local_data.csv", "rb"))
    
    # Run analysis
    result = sandbox.exec("python analyze.py")
    print(result.stdout)
    
    # Download results
    output = sandbox.download_file("results.json")
```

## API Reference

### Client

```python
class SkyscaleClient:
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://api.skyscale.io",
        timeout: int = 30
    )
```

### Functions API

```python
# Invoke function
result = client.functions.invoke(
    name: str,
    payload: Dict[str, Any],
    sync: bool = True,
    timeout: int = 30
) -> Dict[str, Any]

# List functions
functions = client.functions.list() -> List[Function]

# Get function
function = client.functions.get(name: str) -> Function

# Create function
function = client.functions.create(
    name: str,
    runtime: str,
    handler: str,
    code: bytes
) -> Function
```

### Sandboxes API

```python
# Create sandbox
sandbox = client.sandboxes.create(
    runtime: str = "python3.8",
    memory_mb: int = 512,
    cpu_count: int = 1,
    timeout_seconds: int = 3600
) -> Sandbox

# List sandboxes
sandboxes = client.sandboxes.list() -> List[Sandbox]

# Get sandbox
sandbox = client.sandboxes.get(id: str) -> Sandbox
```

### Sandbox Operations

```python
# Execute code
result = sandbox.exec(
    code: str,
    language: str = "python",
    timeout: int = 30
) -> ExecResult

# Execute shell command
result = sandbox.shell(
    command: str,
    timeout: int = 30
) -> ExecResult

# Upload file
sandbox.upload_file(
    path: str,
    content: Union[str, bytes, BinaryIO],
    permissions: str = "644"
)

# Download file
content = sandbox.download_file(path: str) -> bytes

# List files
files = sandbox.list_files(path: str = "/workspace") -> List[FileInfo]

# Destroy sandbox
sandbox.destroy()
```

## Error Handling

```python
from skyscale import SkyscaleError, TimeoutError, AuthenticationError

try:
    result = client.functions.invoke("my-function", {...})
except TimeoutError:
    print("Function execution timed out")
except AuthenticationError:
    print("Invalid API key")
except SkyscaleError as e:
    print(f"Error: {e}")
```

## Async Support

```python
from skyscale import AsyncSkyscaleClient

async def main():
    client = AsyncSkyscaleClient(api_key="sk_...")
    
    # Async invocation
    result = await client.functions.invoke("my-function", {...})
    
    # Async sandbox
    async with client.sandboxes.create() as sandbox:
        result = await sandbox.exec("print('hello')")
```

## Configuration

```python
# From environment
client = SkyscaleClient()  # uses SKYSCALE_API_KEY

# From config file (~/.skyscale/config.json)
{
    "api_key": "sk_...",
    "base_url": "https://api.skyscale.io"
}

client = SkyscaleClient.from_config()
```

## Migration Status

🚧 **To Be Implemented**

See [Issue #37-#40](../../ISSUES.md) for details.

## Development

```bash
# Install dev dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Type checking
mypy skyscale

# Linting
flake8 skyscale
```
