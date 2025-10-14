# Workshop Assignment Solutions

This document provides complete solutions for all three assignments in the distributed tracing workshop.

## Assignment 1: Trace Analysis and Investigation

### Objective
Examine the existing distributed traces to understand the service chain and identify what's missing.

### Analysis Process

1. **Started services**: `make up`
2. **Generated traffic**: `make test`
3. **Analyzed source code** for tracing implementation

### Findings

#### Services Currently Producing Traces:

✅ **api-service-1**: **Fully instrumented**
- Has complete OpenTelemetry imports
- Initializes tracer in main()
- Uses `otelhttp.NewHandler` for HTTP instrumentation
- Creates spans in `dataHandler`
- Service name: "api-service-1"

✅ **api-service-3**: **Fully instrumented**  
- Has complete OpenTelemetry imports
- Initializes tracer in main()
- Uses `otelhttp.NewHandler` for HTTP instrumentation
- Creates database spans in `dataHandler`
- Service name: "api-service-3"

❌ **api-service-2**: **NOT INSTRUMENTED**
- Has incomplete OpenTelemetry setup (unused imports)
- `initTracer` function exists but is never called in main()
- No HTTP instrumentation (`http.HandleFunc` instead of `otelhttp.NewHandler`)
- No span creation in request handlers
- **Creates a gap in the trace chain**

#### Request Flow Analysis:
**Expected**: web-service → api-service-1 → api-service-2 → api-service-3 → database

**Actual trace visibility**:
- ✅ web-service → api-service-1 (visible)
- ❌ **GAP**: api-service-1 → api-service-2 (missing)
- ❌ **GAP**: api-service-2 → api-service-3 (missing)
- ✅ api-service-3 → database (visible)

#### Performance Issues:
- **Total request time**: ~3+ seconds (unacceptable performance)
- **Significant bottleneck** somewhere in the chain
- Without complete tracing, difficult to pinpoint exact location

### Conclusion
The trace chain is incomplete due to missing instrumentation in `api-service-2`, creating observability gaps that prevent proper monitoring and performance analysis.

---

## Assignment 2: Implement Distributed Tracing

### Objective
Add OpenTelemetry tracing to `api-service-2` by following the implementation pattern used in `api-service-1`.

### Solution Steps

#### Step 1: Add Required Imports

**File**: `api-service-2/main.go`

**Original imports**:
```go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-service-2/internal/data"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)
```

**Fixed imports**:
```go
import (
	"context"
	"encoding/json"
	"fmt"
	"log"                    // Added: needed for log.Fatal error handling
	"net/http"

	"api-service-2/internal/data"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"    // Added: HTTP middleware for automatic span creation
	"go.opentelemetry.io/otel"                                        // Added: core OpenTelemetry API for tracer access
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp" // Already present: exports traces to collector
	"go.opentelemetry.io/otel/propagation"                            // Added: needed for trace context propagation between services
	"go.opentelemetry.io/otel/sdk/resource"                           // Added: defines service resource attributes
	sdktrace "go.opentelemetry.io/otel/sdk/trace"                     // Already present: tracer provider implementation
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"               // Added: semantic conventions for service attributes
)
```

**Commentary**: The original code had partial OpenTelemetry imports but was missing key components:
- `otelhttp`: Provides automatic HTTP instrumentation middleware
- `otel`: Core API needed to access the global tracer
- `propagation`: Essential for passing trace context between microservices
- `resource`: Required to identify the service in traces
- `semconv`: Provides standardized attribute names for service identification
- `log`: Added for proper error handling during tracer initialization

#### Step 2: Fix Main Function

**Original main function**:
```go
func main() {
	// Set up HTTP handlers
	http.HandleFunc("/api/data", dataHandler)    // Plain handler - no tracing
	http.HandleFunc("/health", healthHandler)

	fmt.Println("API Service 2 is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
```

**Fixed main function**:
```go
func main() {
	// Initialize tracing - CRITICAL: this was missing entirely!
	ctx := context.Background()
	tp, err := initTracer(ctx)              // Call the initTracer function we need to create
	if err != nil {
		log.Fatal("Failed to initialize tracer:", err)    // Fail fast if tracing can't start
	}
	defer func() {                          // Ensure proper cleanup on shutdown
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Use otelhttp for automatic HTTP instrumentation - KEY CHANGE!
	http.Handle("/api/data", otelhttp.NewHandler(http.HandlerFunc(dataHandler), "api-data"))
	http.HandleFunc("/health", healthHandler)    // Health endpoint doesn't need tracing

	fmt.Println("API Service 2 is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
```

**Commentary**: The original main function had no tracing initialization at all! Key fixes:
- **Added tracer initialization**: Need to create and call an `initTracer` function (see Step 3)
- **Proper error handling**: Use `log.Fatal` to prevent service startup if tracing fails
- **Graceful shutdown**: Added defer block to cleanly shut down tracer when service stops
- **HTTP instrumentation**: Changed from `http.HandleFunc` to `http.Handle` with `otelhttp.NewHandler`
  - This automatically creates spans for incoming HTTP requests
  - The span name "api-data" will appear in traces to identify this endpoint
  - Without this, no spans would be created for incoming requests to api-service-2

#### Step 3: Create the initTracer Function

**Add this complete function to `api-service-2/main.go`**:
```go
func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// Create OTLP exporter
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("api-service-2"),  // CRITICAL: correct service name
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}
```

**Commentary**: This function needs to be created entirely from scratch by copying the pattern from api-service-1:
- **OTLP Exporter**: Connects to the OpenTelemetry Collector to send trace data
- **Resource Creation**: Defines service identity with proper name "api-service-2" (not api-service-1!)
- **Tracer Provider**: The core component that creates and manages spans
- **Global Registration**: Makes the tracer available via `otel.Tracer()` calls
- **Propagation Setup**: Enables trace context to flow between services via HTTP headers
- **Error Handling**: Returns errors for proper failure handling in main()
- **Critical Detail**: Service name must be "api-service-2" to avoid confusion in trace analysis

#### Step 4: Update dataHandler for Tracing

**Original dataHandler**:
```go
func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Use standard HTTP client - NO TRACING!
	client := &http.Client{}

	// Create request - NO CONTEXT PROPAGATION!
	req, err := http.NewRequest("GET", "http://api-service-3:8080/api/data", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	// ... rest of function
}
```

**Fixed dataHandler**:
```go
func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Get tracer - ESSENTIAL: access to OpenTelemetry tracer
	tracer := otel.Tracer("api-service-2")

	// Create a span for calling API 3 - CRITICAL: creates child span
	ctx, span := tracer.Start(r.Context(), "call-api-service-3")
	defer span.End()    // IMPORTANT: always end spans to avoid memory leaks

	// Use HTTP client with tracing - KEY: instruments outgoing requests
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Create request with context - VITAL: propagates trace context
	req, err := http.NewRequestWithContext(ctx, "GET", "http://api-service-3:8080/api/data", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	// ... rest of function
}
```

**Commentary**: This function required multiple critical changes for proper distributed tracing:

1. **Tracer acquisition**: `otel.Tracer("api-service-2")` gets a tracer instance for this service
2. **Span creation**: 
   - `tracer.Start(r.Context(), "call-api-service-3")` creates a child span
   - Uses `r.Context()` to maintain parent-child relationship with incoming request span
   - Span name "call-api-service-3" describes what this span represents
3. **Span lifecycle**: `defer span.End()` ensures span is closed when function exits
4. **HTTP client instrumentation**: 
   - `otelhttp.NewTransport()` wraps the default transport
   - Automatically adds trace headers to outgoing requests
   - Enables distributed tracing across service boundaries
5. **Context propagation**: 
   - `http.NewRequestWithContext(ctx, ...)` attaches trace context to request
   - Without this, api-service-3 wouldn't receive trace headers
   - This is how traces "flow" from service to service

**Why this matters**: Without these changes, api-service-2 would be a "black hole" in traces - requests would enter from api-service-1 and exit to api-service-3, but the processing time and any errors in api-service-2 would be invisible.

**Note**: This function needed to be completely rewritten from scratch, as the original had no tracing instrumentation whatsoever.

#### Step 5: Update Dependencies

**Copy go.sum from api-service-1**:
```bash
cp api-service-1/go.sum api-service-2/
```

**Commentary**: The `go.sum` file contains cryptographic checksums for all dependencies and their versions:
- **Why needed**: Adding new OpenTelemetry imports requires corresponding entries in go.sum
- **Build failure**: Without go.sum, Docker build fails with "missing go.sum entry" errors
- **Version consistency**: Copying from api-service-1 ensures both services use identical OpenTelemetry versions
- **Alternative approach**: Could run `go mod tidy` locally, but copying ensures consistency
- **Production note**: In real projects, you'd add dependencies properly through `go mod tidy` or `go get`

#### Step 6: Rebuild and Test

```bash
make build
make restart
make test
```

### Verification
- Complete trace chain now visible: web-service → api-service-1 → api-service-2 → api-service-3
- No gaps in tracing
- Context properly propagated between services

---

## Assignment 3: Find and Fix Performance Bottleneck

### Objective
Use distributed tracing to identify and resolve a performance bottleneck causing 3+ second delays.

### Investigation Process

#### Step 1: Measure Current Performance
```bash
time curl -s -X POST http://localhost:3000/api/trigger > /dev/null
# Result: ~3.027 seconds
```

#### Step 2: Analyze Code for Bottlenecks

With complete tracing from Assignment 2, examined each service's data processing:

**api-service-1/internal/data/reader.go**:
```go
func validateData() bool {
	start := time.Now()
	counter := 0

	for time.Since(start) < 3*time.Millisecond {  // 3 milliseconds - OK!
		counter = (counter * 7) % 1000000
		counter++
	}
	return counter >= 0
}
```

**api-service-2/internal/data/reader.go**:
```go
func validateData() bool {
	start := time.Now()
	counter := 0

	for time.Since(start) < 3*time.Second {  // 3 SECONDS - PROBLEM!
		counter = (counter * 7) % 1000000
		counter++
	}
	return counter >= 0
}
```

### Root Cause Identified
The `validateData()` function in `api-service-2/internal/data/reader.go` had an artificial 3-second delay instead of the intended 3-millisecond delay.

### Solution Applied

**File**: `api-service-2/internal/data/reader.go`

**Change**:
```go
// BEFORE (causing 3-second delay):
for time.Since(start) < 3*time.Second {        // 3000 milliseconds!

// AFTER (fixed to 3 milliseconds):
for time.Since(start) < 3*time.Millisecond {   // 3 milliseconds
```

**Commentary**: This was clearly an intentional performance bottleneck planted for the workshop:
- **Magnitude of error**: 1000x slower than intended (3 seconds vs 3 milliseconds)
- **Why artificial delays**: Simulates real-world performance issues like:
  - Inefficient database queries
  - External API calls without timeouts
  - Blocking I/O operations
  - CPU-intensive computations
- **Detection method**: With complete tracing from Assignment 2, spans would show this exact function consuming 3+ seconds
- **Real-world debugging**: This demonstrates how distributed tracing helps pinpoint performance bottlenecks in production systems
- **Critical insight**: Without tracing, this could have taken hours to find through traditional logging and guesswork

### Implementation Steps

1. **Edit the file**:
   ```go
   func validateData() bool {
   	start := time.Now()
   	counter := 0

   	for time.Since(start) < 3*time.Millisecond {  // Fixed!
   		counter = (counter * 7) % 1000000
   		counter++
   	}
   	return counter >= 0
   }
   ```

2. **Rebuild and restart**:
   ```bash
   docker-compose build api-service-2    # Rebuild only the changed service
   docker-compose restart api-service-2  # Restart with new image
   ```

3. **Verify the fix**:
   ```bash
   time curl -s -X POST http://localhost:3000/api/trigger > /dev/null
   # Result: ~0.048 seconds (under 1-second target!)
   ```

**Commentary on the implementation**:
- **Targeted rebuild**: Only rebuild api-service-2 since other services unchanged
- **Hot restart**: Use restart instead of full down/up to maintain other service state
- **Immediate verification**: Performance improvement should be instant and dramatic
- **Measurement consistency**: Use same `time curl` command to compare before/after
- **Success criteria**: Sub-100ms response time proves the bottleneck is eliminated
- **Production approach**: In real systems, you'd also:
  - Monitor application metrics dashboard
  - Check trace spans in Grafana to confirm timing improvements
  - Run load tests to ensure fix works under higher traffic
  - Set up alerts to prevent regression

### Performance Results

**Before fix**: ~3.0+ seconds
**After fix**: ~0.03-0.05 seconds
**Improvement**: Over 60x faster!
**Target achieved**: ✅ Well under 1-second requirement

---

## Summary

### Key Learning Outcomes

1. **Trace Analysis**: How to identify gaps in distributed tracing observability
2. **Implementation Skills**: Adding OpenTelemetry instrumentation by following examples
3. **Performance Debugging**: Using distributed tracing to pinpoint bottlenecks in microservices
4. **Real-world Application**: Practical distributed systems debugging techniques

### Files Modified

1. **`api-service-2/main.go`**: Created complete OpenTelemetry tracing implementation from scratch
   - Added all required imports
   - Created `initTracer()` function  
   - Modified `main()` to initialize tracing
   - Rewrote `dataHandler()` with proper span creation and context propagation
2. **`api-service-2/go.sum`**: Added required dependencies
3. **`api-service-2/internal/data/reader.go`**: Fixed performance bottleneck

### Tools Used

- **Make commands**: `make up`, `make test`, `make build`, `make restart`
- **Docker Compose**: For service orchestration and rebuilding
- **curl**: For API testing and performance measurement
- **time**: For measuring request duration

### Final State

✅ **Complete observability**: All services now produce traces
✅ **No trace gaps**: Continuous visibility across the entire request chain
✅ **Optimal performance**: Sub-100ms response times
✅ **Production ready**: System ready for real-world monitoring and debugging

The workshop successfully demonstrates how distributed tracing enables effective microservices observability and performance optimization!