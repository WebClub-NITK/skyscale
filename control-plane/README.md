# Skyscale Control Plane

The Skyscale Control Plane is the central component of the Skyscale serverless platform. It manages the lifecycle of functions, VMs, and executions.

## Architecture

The control plane consists of several components:

1. **State Manager**: Manages the state of the system, including functions, VMs, and executions.
2. **Function Registry**: Manages the registration, update, and deletion of functions.
3. **VM Manager**: Manages the lifecycle of Firecracker micro-VMs.
4. **Scheduler**: Schedules function executions on VMs.
5. **Auth Manager**: Handles authentication and authorization.
6. **API Handler**: Exposes the control plane functionality through a REST API.

## API Endpoints

### Authentication

- `POST /api/auth/api-key`: Generate a new API key

### Functions

- `GET /api/functions`: List all functions
- `POST /api/functions`: Register a new function
- `GET /api/functions/{id}`: Get a function by ID
- `PUT /api/functions/{id}`: Update a function
- `DELETE /api/functions/{id}`: Delete a function
- `POST /api/functions/{id}/invoke`: Invoke a function
- `GET /api/functions/name/{name}`: Get a function by name
- `POST /api/functions/name/{name}/invoke`: Invoke a function by name

### Executions

- `GET /api/executions/{id}`: Get an execution by ID
- `GET /api/executions/function/{id}`: List all executions for a function

### VMs

- `GET /api/vms`: List all VMs
- `GET /api/vms/{id}`: Get a VM by ID

## Getting Started

### Prerequisites

- Go 1.21 or later
- SQLite
- Redis (optional, for caching)
- Firecracker

### Installation

1. Clone the repository:

```bash
git clone https://github.com/bluequbit/faas.git
cd faas/control-plane
```

2. Install dependencies:

```bash
go mod download
```

3. Build the control plane:

```bash
go build -o skyscale-control-plane
```

4. Run the control plane:

```bash
./skyscale-control-plane
```

## Configuration

The control plane can be configured using environment variables:

- `PORT`: The port to listen on (default: 8080)
- `DB_PATH`: The path to the SQLite database (default: skyscale.db)
- `REDIS_ADDR`: The address of the Redis server (default: localhost:6379)
- `REDIS_PASSWORD`: The password for the Redis server (default: none)
- `REDIS_DB`: The Redis database to use (default: 0)
- `LOG_LEVEL`: The log level (default: info)
- `WARM_POOL_SIZE`: The size of the warm VM pool (default: 5)

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o skyscale-control-plane -ldflags="-s -w" .
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 