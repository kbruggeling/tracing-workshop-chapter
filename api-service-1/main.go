package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"api-service-1/internal/data"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// Initialize tracing
	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Use otelhttp for automatic HTTP instrumentation
	http.Handle("/api/data", otelhttp.NewHandler(http.HandlerFunc(dataHandler), "api-data"))
	http.HandleFunc("/health", healthHandler)

	fmt.Println("API Service 1 is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "api-service-1"})
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	// Create OTLP exporter
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("api-service-1"),
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

func dataHandler(w http.ResponseWriter, r *http.Request) {
	// Get tracer
	tracer := otel.Tracer("api-service-1")

	// Create a span for calling API 2
	ctx, span := tracer.Start(r.Context(), "call-api-service-2")
	defer span.End()

	// Use HTTP client with tracing
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", "http://api-service-2:8080/api/data", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Call API 2
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to call API 2", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response from API 2 using internal package
	body, err := data.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from API 2", http.StatusInternalServerError)
		return
	}

	// Return the data from API 2
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
