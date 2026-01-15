# Runtime Python

The Python runtime module handles Python execution inside VMs.

## Responsibilities

- **Environment Setup**: Initialize Python environment
- **Handler Loading**: Load and validate handler functions
- **Dependency Management**: Install and manage dependencies
- **Execution**: Execute handler with event and context
- **Error Handling**: Capture and format errors
- **Output Capture**: Capture stdout/stderr

## Key Features

- AWS Lambda-compatible handler format
- Support for requirements.txt
- Isolated execution environment
- Comprehensive error reporting

## Handler Format

```python
def handler(event, context):
    # Your code here
    return response
```

## Migration Status

🚧 **To Be Extracted** - Python runtime logic currently embedded in various places

See [Issue #14](../../ISSUES.md#issue-14-create-runtimepython-module) for details.

## Future Structure

```
runtime/python/
├── bootstrap.py    # VM boot initialization
├── handler.py      # Handler loading and execution
├── context.py      # Context object
└── requirements.py # Dependency management
```

## Execution Flow

1. Bootstrap: Initialize Python on VM boot
2. Load: Load handler function from uploaded code
3. Execute: Call handler(event, context)
4. Return: Stream result back to host
