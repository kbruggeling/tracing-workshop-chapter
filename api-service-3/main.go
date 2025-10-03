package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Created string `json:"created"`
}

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

	fmt.Println("API Service 3 is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "api-service-3"})
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
			semconv.ServiceNameKey.String("api-service-3"),
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
	tracer := otel.Tracer("api-service-3")

	// Create a span for database operations using the request context
	ctx, dbSpan := tracer.Start(r.Context(), "database-query-users")
	defer dbSpan.End()

	// Connect to database
	db, err := sql.Open("postgres", "host=database port=5432 user=testuser password=testpass dbname=testdb sslmode=disable")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database within the span context
	rows, err := db.QueryContext(ctx, "SELECT id, name, email, created_at FROM users ORDER BY id LIMIT 10")
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var createdAt time.Time
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &createdAt)
		if err != nil {
			http.Error(w, "Failed to scan database row", http.StatusInternalServerError)
			return
		}
		user.Created = createdAt.Format("2006-01-02 15:04:05")
		users = append(users, user)
	}

	// Return the data
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"service":   "api-service-3",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"users":     users,
	}
	json.NewEncoder(w).Encode(response)
}
