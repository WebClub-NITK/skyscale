how to build a serverless platform

The core of the serverless platform is the Lambda function.

A Lambda function is a function that is executed in the cloud, and can be triggered by an event.

rn we will only support python

- there will be go agent running inside the firecracker VM.
- The agent will handle upcoming requests and pass it to the correct runtime - python in this case
- The agent will be written in Go and only the binary will be copied to the VM.


Here's a step-by-step plan to build **Skyscale**:

---

### **1. Define Project Scope and Requirements**
- **Core Features**:
  - Support Python functions with optimized invocation.
  - CLI to scaffold and manage functions.
  - Deployment and execution in isolated micro-VMs using Firecracker.
  - Logging and monitoring for executed functions.
- **Additional Goals**:
  - Minimize cold start times.
  - Provide pre-warmed micro-VMs for faster execution.

---

https://qxf2.com/blog/build-a-microvm-using-firecracker/

  - **CLI Tool**:
    - Scaffolds new functions.
    - Deploys and manages function configurations.
    - Manages local testing.
  - **Function Runtime**:
    - Isolated micro-VMs with Firecracker.
    - Python server (e.g., FastAPI) inside the micro-VM to execute functions.
  - **Control Plane**:
    - Schedules function invocations.
    - Manages VM lifecycles.
    - Allocates resources dynamically.
  - **Data Plane**:
    - Handles communication between the host and micro-VMs.
    - Transfers inputs/outputs efficiently.
  - **Pre-Warm Pool**:
    - Maintains a pool of pre-warmed micro-VMs to reduce cold starts.

---

### **3. Development Plan**
#### **Step 1: Setup Base Micro-VM Environment**
- **Optimize Firecracker Image**:
  - Use a minimal custom kernel and filesystem to reduce image size.
  - Include only the required Python runtime and dependencies.
- **Automate Image Creation**:
  - Use tools like Packer or custom scripts to build VM images.
- **Add Networking**:
  - Implement virtual network interfaces for micro-VMs (fcnet).

#### **Step 2: Implement CLI for Scaffolding Functions**
- **Features**:
  - Scaffold Python functions with a predefined template.
  - Support function metadata (e.g., runtime, memory, timeout).
  - Deploy functions using the CLI (packaging and uploading code to a server).
- **Technologies**:
  - Use Python's `argparse` or `click` for CLI development.

#### **Step 3: Build the Function Runtime**
- Create a base VM image with:
  - A lightweight Python server (e.g., FastAPI or Flask) to handle HTTP requests.
  - Dependency installation support (`requirements.txt`).
- Implement a script for bootstrapping user-defined functions.

#### **Step 4: Develop the Control Plane**
- **VM Orchestrator**:
  - Automate the lifecycle of Firecracker VMs (start, stop, terminate).
- **Scheduler**:
  - Distribute incoming function requests to appropriate VMs.
  - Use a lightweight queue (e.g., RabbitMQ or Redis).

#### **Step 5: Optimize Invocation Flow**
- **Reduce Cold Starts**:
  - Maintain pre-warmed VMs ready to execute functions.
  - Implement lazy loading for rarely-used functions.
- **Efficient Data Transfer**:
  - Use shared memory or sockets for communication between the host and micro-VMs.
- **Request Routing**:
  - Proxy requests through a lightweight service (e.g., Nginx or HAProxy).

#### **Step 6: Build Logging and Monitoring**
- Add logging for:
  - Function execution times.
  - Resource usage (CPU, memory).
  - Errors and exceptions.
- Implement a lightweight monitoring dashboard.

#### **Step 7: Add Deployment Capabilities**
- Package and deploy functions using the CLI.
- Store function metadata in a lightweight database (e.g., SQLite, Postgres).

#### **Step 8: Optimize Performance**
- Benchmark cold start times and function invocation latency.
- Experiment with:
  - Snapshotting VMs after the Python server initializes.
  - Kernel and filesystem optimization for Firecracker.
  - Multi-threading in the control plane for parallel invocations.

#### **Step 9: Implement a Local Testing Environment**
- Allow developers to test their functions locally before deployment.
- Use Docker to mimic the runtime environment.

---

### **4. Milestones**
1. **Month 1**: 
   - Setup Firecracker-based VM with Python runtime.
   - Implement the CLI tool for scaffolding functions.
2. **Month 2**:
   - Build and test the control plane (VM lifecycle and function scheduling).
   - Develop logging and monitoring capabilities.
3. **Month 3**:
   - Optimize invocation flow with pre-warmed VMs.
   - Add local testing capabilities.

---

### **5. Deployment Strategy**
- Use Docker or Kubernetes to manage Skyscale services.
- Automate deployments with CI/CD pipelines (GitHub Actions, GitLab CI).
- Provide a self-hosted option for early users.

---

### **6. Developer Experience (DevEx)**
- **CLI Commands**:
  - `skyscale init` - Scaffold a new function.
  - `skyscale deploy` - Deploy a function to the platform.
  - `skyscale logs` - View function logs.
  - `skyscale test` - Test the function locally.
- **Templates**:
  - Provide common function templates for tasks like API calls or file processing.

---

### **7. Testing and Iteration**
- Write unit and integration tests for each component.
- Set up a benchmark suite to test performance under load.
- Gather feedback from early adopters to refine the platform.



Let's break down the lifecycle and internal working of **Skyscale** using the example function `print("hello fire")`. I'll explain each step in detail, from function creation to execution and cleanup.

---

### **Lifecycle of the Function**

#### **1. User Scaffolds the Function**
- The user runs the CLI command:
  ```bash
  skyscale init --name hello_fire
  ```
  - **What Happens Internally**:
    - The CLI creates a directory structure like:
      ```
      hello_fire/
        ├── handler.py  # User's function
        ├── requirements.txt  # Dependencies
        └── skyscale_config.json  # Metadata for the function
      ```
    - `handler.py` contains a scaffolded template for a Python function:
      ```python
      def handler():
          print("hello fire")
      ```

---

#### **2. Deploy the Function**
- The user deploys the function using:
  ```bash
  skyscale deploy --name hello_fire
  ```
  - **What Happens Internally**:
    - **Step 1: Package Function**:
      - The function code, dependencies, and metadata are bundled into a `.zip` or `.tar.gz` file.
    - **Step 2: Upload to Skyscale Control Plane**:
      - The CLI sends the package to a server hosting the control plane.
      - The control plane stores the package in a central repository (e.g., an S3 bucket or a file system).
    - **Step 3: VM Image Preparation**:
      - A base Firecracker VM image is cloned.
      - The function package is injected into the VM image at a predefined path (e.g., `/home/hello_fire`).
      - A lightweight Python server is pre-installed in the VM to execute the function.

---

#### **3. Function Invocation**
- The user invokes the function:
  ```bash
  skyscale invoke --name hello_fire
  ```
  - **What Happens Internally**:
    - **Step 1: Request Received**:
      - The control plane receives the request and checks:
        - If a pre-warmed VM is available for the function.
        - If not, it starts a new Firecracker micro-VM using the prepared VM image.
    - **Step 2: VM Initialization**:
      - The VM boots up with a minimal Linux kernel and a root filesystem.
      - The Python server inside the VM starts listening for requests.
    - **Step 3: Function Execution**:
      - The request payload (if any) is sent to the Python server inside the VM via an HTTP POST request.
      - The server dynamically loads and executes `handler.py` using Python's `exec()` function.
      - `print("hello fire")` writes output to the VM's standard output (stdout).
    - **Step 4: Capture Output**:
      - The server captures the stdout of the function execution.
      - The output is sent back to the control plane, which relays it to the CLI.
    - **Step 5: VM Management**:
      - If the VM is pre-warmed, it is returned to the pool for reuse.
      - If the VM was created for a single invocation, it is terminated.

---

#### **4. Logging and Monitoring**
- Logs are generated at each step:
  - **Invocation Logs**:
    - Function output: `"hello fire"`.
    - Execution time, resource usage, and errors (if any).
  - **Control Plane Logs**:
    - Function scheduling details.
    - VM lifecycle events.
  - These logs are stored in a central logging system and accessible via:
    ```bash
    skyscale logs --name hello_fire
    ```

---

#### **5. Cleanup**
- After execution, the VM is either:
  - Returned to the **Pre-Warm Pool** for future invocations.
  - Terminated if it's no longer needed (e.g., after idle timeout).
- Temporary files and data inside the VM are cleared to maintain security and optimize resource usage.

---

### **Internal Working Details**

#### **Firecracker Micro-VMs**
- **Bootstrapping**:
  - A Firecracker VM boots in milliseconds with a minimal kernel and root filesystem.
- **Execution Environment**:
  - The VM contains only the required Python runtime and dependencies.
  - The function is executed inside an isolated Python server running in the VM.

#### **Optimized Invocation**
- **Cold Start Optimization**:
  - Pre-warmed VMs are kept ready with Python initialized and the server running.
  - This avoids the overhead of booting a VM from scratch for each invocation.
- **Efficient Communication**:
  - Input/output data is passed between the host and VM using Unix sockets or shared memory for low latency.

#### **Control Plane**
- Handles scheduling and lifecycle management of VMs.
- Allocates resources dynamically based on function demand.
- Manages the pre-warm pool to minimize cold starts.

#### **CLI Tool**
- Provides a seamless interface for scaffolding, deploying, and managing functions.
- Relays user commands to the control plane via REST or gRPC APIs.

---

### **Example Timeline**
Let's assume the user invokes the function:
```bash
skyscale invoke --name hello_fire
```
- **T=0 ms**: Request is received by the control plane.
- **T=50 ms**: A pre-warmed Firecracker VM is assigned to the request.
- **T=100 ms**: The function package is loaded, and `handler.py` is executed.
- **T=200 ms**: Output (`hello fire`) is captured and returned to the user.
- **T=250 ms**: The VM is returned to the pre-warm pool.

---

This lifecycle ensures:
1. **Fast Invocations**: Pre-warmed VMs reduce cold starts.
2. **Isolation**: Each function runs in a secure, isolated micro-VM.
3. **Scalability**: The system can dynamically scale by spinning up new Firecracker VMs based on demand.

Does this address your question, or would you like a deep dive into a specific component?

ref  
https://qxf2.com/blog/build-a-microvm-using-firecracker/

https://stanislas.blog/2021/08/firecracker/#agent

https://news.ycombinator.com/item?id=34964197

Regarding copying the code to the VM, we can use the following command:

I need to keep the worker nodes isolated from each other. This means that the code that needs to be copied to the VM will happen at runtime.

This means that it will be the responsiblity of the agent/deamon running insdie each VM to copy the code to the VM.

Now till the invocation happens. The code is stored somewhere secured and handled by the control plane.

The agent, when live, will poll the control plane for new code to be executed.

## Daemon Implementation

The daemon will run inside the micro VM and will be responsible for:

1. Polling the control plane for new code.
2. Downloading the code securely.
3. Executing the downloaded code.

### Example Daemon Workflow

1. **Initialization**: Start the daemon and establish a secure connection to the control plane.
2. **Polling**: Periodically check for new code updates.
3. **Download**: If new code is available, download it securely.
4. **Execution**: Execute the downloaded code in a controlled environment.
5. **Logging**: Log the execution results and any errors for monitoring.

### Security Considerations

- Ensure that the communication with the control plane is encrypted.
- Validate the integrity of the downloaded code before execution.

### **Detailed Control Plane Requirements**

#### **1. API Layer**
- **Function Management API**:
  - Register new functions with metadata
  - Update existing functions
  - Delete functions
  - List available functions
  - Get function status and configuration
- **Invocation API**:
  - Synchronous function invocation
  - Asynchronous function invocation
  - Get invocation results
  - Cancel running invocations
- **System Management API**:
  - Health checks
  - Resource utilization metrics
  - System configuration updates
  - VM pool management

#### **2. Authentication & Authorization**
- **Security Requirements**:
  - API key management
  - Role-based access control (RBAC)
  - Token-based authentication
  - Request signing for secure communication
- **Access Control**:
  - Function-level permissions
  - Resource usage quotas
  - Rate limiting

#### **3. Function Registry**
- **Storage Requirements**:
  - Function code storage (compressed packages)
  - Function metadata storage
  - Version control and rollback capability
- **Database Schema**:
  ```sql
  functions:
    - id: uuid
    - name: string
    - runtime: string
    - memory: integer
    - timeout: integer
    - created_at: timestamp
    - updated_at: timestamp
    - status: enum
    - version: string
  ```

#### **4. VM Management**
- **Lifecycle Operations**:
  - VM creation and initialization
  - VM health monitoring
  - VM termination and cleanup
  - Pre-warm pool management
- **Resource Management**:
  - CPU allocation
  - Memory monitoring
  - Network bandwidth control
  - Disk space management

#### **5. Scheduler**
- **Scheduling Capabilities**:
  - Function placement optimization
  - Load balancing across VMs
  - Resource-aware scheduling
  - Priority-based execution
- **Queue Management**:
  - Request queuing
  - Backpressure handling
  - Dead letter queues
  - Retry mechanisms

#### **6. Monitoring & Logging**
- **Metrics Collection**:
  - Function execution metrics
  - VM performance metrics
  - System health metrics
  - Resource utilization
- **Logging Requirements**:
  - Function execution logs
  - System operation logs
  - Error tracking
  - Audit trails

#### **7. Communication Protocol**
- **VM Communication**:
  - Secure WebSocket connections
  - gRPC for efficient communication
  - Health check protocol
  - Binary protocol for data transfer
- **Agent Communication**:
  ```go
  type AgentMessage struct {
    Type    string          // command, response, error
    ID      string          // unique message identifier
    Payload json.RawMessage // message content
    Time    time.Time       // timestamp
  }
  ```

#### **8. State Management**
- **System State**:
  - Active function executions
  - VM pool status
  - Resource allocation state
  - System configuration
- **Persistence Layer**:
  - Database for long-term storage
  - In-memory cache for hot data
  - State synchronization
  - Backup and recovery

#### **9. Scaling & High Availability**
- **Scalability Features**:
  - Horizontal scaling of control plane
  - Load distribution
  - Data partitioning
  - Cache distribution
- **High Availability**:
  - Leader election
  - Failover mechanisms
  - Data replication
  - Service discovery

#### **10. Integration Points**
- **External Systems**:
  - Metrics exporters (Prometheus)
  - Log aggregators (ELK stack)
  - Alert managers
  - External storage systems
- **API Integrations**:
  - Webhook support
  - Event streaming
  - External triggers
  - Service mesh integration