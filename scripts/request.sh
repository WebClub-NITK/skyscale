#!/bin/bash

curl -X POST http://localhost:8081/execute \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "test-func-123",
    "name": "test-function",
    "code": "import os\ndef handler(event, context):\n   print(os.environ)\n   return \"Hello from FaaS!\"",
    "requirements": "",
    "config": "name: test-function\nversion: 1.0.0",
    "runtime": "python3",
    "entry_point": "handler.handler",
    "environment": {},
    "request_id": "req-123-456",
    "timeout": 30,
    "memory": 128,
    "version": "1.0",
    "context": {
      "function_name": "test-function",
      "function_version": "1.0",
      "memory_limit_mb": 128,
      "request_id": "req-123-456",
      "remaining_time_ms": 30000
    }
  }'