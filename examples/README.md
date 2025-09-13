# FaaS Lambda-style Handler Functions

This document provides examples and guidelines for writing Lambda-style handler functions for the FaaS platform.

## Handler Function Signature

All handler functions must have the following signature:

```python
def handler(event, context):
    # Function code
    return some_value
```

Where:
- `event`: A dictionary containing the input data for the function
- `context`: An object providing runtime information about the function execution
- `some_value`: The return value (can be a dict, list, string, or other JSON-serializable value)

## Event Object

The `event` parameter contains the input data for the function. It will be a Python dictionary with whatever data was passed to the function when it was invoked. For example:

```python
{
    "name": "John",
    "age": 30,
    "query_parameters": {
        "limit": 10,
        "offset": 0
    }
}
```

## Context Object

The `context` object provides information about the function's execution environment. It has the following attributes and methods:

### Attributes:
- `function_name`: The name of the Lambda function
- `function_version`: The version of the function
- `memory_limit_mb`: The memory limit in MB
- `request_id`: The unique ID of this execution request
- `remaining_time_ms`: Initial value of function timeout in milliseconds

### Methods:
- `get_remaining_time_in_millis()`: Returns the number of milliseconds left before the execution times out

## Return Values

Your function can return:
- A Python dict or list: Will be automatically converted to JSON
- A string: Will be returned as-is
- Numbers, booleans, or None: Will be converted to their JSON equivalents

## Examples

### Basic handler

```python
def handler(event, context):
    return {
        "message": "Hello, " + event.get("name", "World"),
        "request_id": context.request_id
    }
```

### Web API handler

```python
def handler(event, context):
    # Extract information from event
    http_method = event.get("httpMethod")
    path = event.get("path")
    query_params = event.get("queryStringParameters", {})
    body = event.get("body")
    
    # Process based on method and path
    if http_method == "GET" and path == "/users":
        return {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "body": {
                "users": ["user1", "user2", "user3"],
                "limit": query_params.get("limit", 10)
            }
        }
    
    # Handle unsupported routes
    return {
        "statusCode": 404,
        "body": {"error": "Not found"}
    }
```

### Using the context object

```python
def handler(event, context):
    # Use context info to make decisions
    if context.get_remaining_time_in_millis() < 1000:
        # Less than 1 second remaining, don't start complex operations
        return {"error": "Insufficient time to process request"}
    
    return {
        "function_info": {
            "name": context.function_name,
            "version": context.function_version,
            "memory": context.memory_limit_mb,
            "request_id": context.request_id
        }
    }
``` 