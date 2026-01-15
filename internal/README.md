# Internal Module

Shared internal utilities used across Skyscale modules.

## Purpose

Common utilities to reduce code duplication and improve maintainability.

## Submodules

### `internal/logging/`
Centralized logging utilities:
- Structured logging
- Log levels
- Context-aware logging
- Log formatting

### `internal/config/`
Configuration management:
- Config file loading
- Environment variable parsing
- Validation
- Defaults

### `internal/errors/`
Error handling utilities:
- Error types
- Error codes
- Error wrapping
- Error formatting

## Important

**Internal packages are not part of the public API.**

These utilities are for use within Skyscale only and may change without notice.

## Status

🚧 **To Be Implemented**

See [Issue #44-#50](../ISSUES.md) for tracking.

## Usage

```go
import (
    "github.com/bluequbit/faas/internal/logging"
    "github.com/bluequbit/faas/internal/config"
    "github.com/bluequbit/faas/internal/errors"
)

// Logging
logger := logging.New("component")
logger.Info("message", "key", "value")

// Config
cfg, err := config.Load("config.yaml")

// Errors
err := errors.New(errors.CodeNotFound, "resource not found")
```
