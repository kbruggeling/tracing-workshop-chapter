# Distributed Tracing Workshop

This project demonstrates a distributed tracing setup with multiple microservices. It includes a web service with a button that triggers a chain of API calls through three Go services to fetch data from a PostgreSQL database.

## Architecture

The application consists of the following components arranged in a chain architecture:

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐
│ Web Service │───▶│ API Service │───▶│ API Service │───▶│ API Service │───▶│   PostgreSQL    │
│  (Node.js)  │    │      1      │    │      2      │    │      3      │    │    Database     │
│   Port 3000 │    │  Port 8081  │    │  Port 8082  │    │  Port 8083  │    │    Port 5432    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────────┘
```

### Service Details

1. **Web Service** (Node.js/Express)
   - Serves a simple HTML page with a button
   - Initiates the API chain when button is clicked
   - Port: 3000

2. **API Service 1** (Go)
   - Receives requests from the web service
   - Forwards requests to API Service 2
   - Port: 8081

3. **API Service 2** (Go)
   - Receives requests from API Service 1
   - Forwards requests to API Service 3
   - Port: 8082

4. **API Service 3** (Go)
   - Receives requests from API Service 2
   - Queries the PostgreSQL database
   - Returns user data up the chain
   - Port: 8083

5. **PostgreSQL Database**
   - Contains sample user data (10 users with names and emails)
   - Database: testdb
   - Credentials: testuser/testpass
   - Port: 5432

### Observability Stack

6. **OpenTelemetry Collector**
   - Collects and processes telemetry data
   - Ready for distributed tracing implementation
   - Port: 8888 (metrics), 13133 (health)

7. **Tempo**
   - Distributed tracing backend
   - Stores and queries trace data
   - Port: 3200

8. **Grafana**
   - Visualization dashboard for traces and metrics
   - Pre-configured with Tempo as datasource
   - Port: 3002 (admin/admin)

## Setup Guide

### Prerequisites

- Docker
- Docker Compose
- Make (optional, but recommended)

### Quick Setup with Make

This project includes a Makefile with convenient commands:

```bash
# Show all available commands
make help

# Build all Docker images
make build

# Start all services in the background
make up

# Check health of all services
make health

# Show all service URLs
make urls

# Test the complete API chain
make test
```

### Step-by-Step Setup

1. **Clone and navigate to the project**:
   ```bash
   git clone <repository-url>
   cd tracing-workshop-chapter
   ```

2. **Build and start all services**:
   ```bash
   make build
   make up
   ```

3. **Verify all services are running**:
   ```bash
   make health
   ```

4. **Access the applications**:
   ```bash
   make urls
   ```
   - Web Application: http://localhost:3000
   - Grafana Dashboard: http://localhost:3002 (admin/admin)
   - Individual API health checks: ports 8081, 8082, 8083

5. **Test the API chain**:
   ```bash
   make test
   ```
   Or manually visit http://localhost:3000 and click "Trigger API Chain"

### Useful Commands

- **View logs**: `make logs`
- **Restart services**: `make restart`
- **Stop services**: `make down`
- **Clean up (stop and remove volumes)**: `make clean`

## Assignments

### Assignment 1: Trace Analysis and Investigation

**Objective**: Examine the existing distributed traces to understand the service chain and identify what's missing.

**Background**: The application is already partially instrumented with distributed tracing. Your task is to analyze the current tracing data to understand how the services interact and spot any gaps.

**Steps**:
1. Start the services: `make up`
2. Generate some traffic: `make test`
3. Open Grafana: http://localhost:3000 (admin/admin)
4. Navigate to **Explore** → **Tempo** data source
5. Search for traces and examine them

**Questions to answer**:
- How many services are currently producing traces?
- Can you see the complete request flow from web-service to api-service-3?
- Which service appears to be missing from the trace chain?
- What's the total duration of requests? Is there any performance issue visible?

### Assignment 2: Implement Distributed Tracing

**Objective**: Add OpenTelemetry tracing to `api-service-2` by following the implementation pattern used in `api-service-1`.

**Background**: 
Currently, `api-service-1` has distributed tracing implemented, but `api-service-2` is missing this instrumentation. This creates a gap in our trace visibility when requests flow through the service chain.

**Task**:
1. Examine the tracing implementation in `api-service-1/main.go`
2. Identify the OpenTelemetry imports, initialization, and span creation patterns
3. Apply the same tracing pattern to `api-service-2/main.go`
**Use `make rebuild` after making changes to ensure they are implemented in your stack*
4. Ensure spans are properly created for incoming requests and outgoing calls to `api-service-3`

**Expected Outcome**:
After implementation, traces should flow continuously from `api-service-1` → `api-service-2` → `api-service-3`, with no missing spans in the chain.

**Verification**:
- Use `make test` to trigger the API chain
- Check Grafana (http://localhost:3002) to verify traces appear for `api-service-2`
- Ensure trace context is properly propagated to `api-service-3`

### Assignment 3: Find and Fix Performance Bottleneck

**Objective**: Use distributed tracing to identify and resolve a performance bottleneck that causes API requests to take more than 3 seconds.

**Background**: 
Users have reported that the API chain is slow, with requests taking significantly longer than expected. Now that you have complete tracing visibility across all services (from Assignment 1), you can use this observability to pinpoint exactly where the delay occurs.

**Task**:
1. Trigger multiple API requests using `make test` or the web interface
2. Use Grafana to analyze the trace data and identify which service/operation is causing the delay
3. Examine the code in the problematic service to understand the root cause
4. Implement a fix to reduce the response time to under 1 second
5. Verify the fix using traces to confirm the bottleneck is resolved

**Investigation Steps**:
- Look at trace duration and span timings in Grafana
- Identify which service has the longest span duration
- Check for any artificial delays, inefficient database queries, or blocking operations
- Compare "before" and "after" trace timings

**Expected Outcome**:
The complete API chain should respond in under 1 second, with traces clearly showing the performance improvement across all services.

**Verification**:
- Use `make test` to measure response times
- Compare trace durations in Grafana before and after the fix
- Ensure all services maintain proper tracing after the performance fix