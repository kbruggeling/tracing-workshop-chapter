# Distributed Tracing Application Makefile

.PHONY: build up down logs restart clean test health

# Build all services
build:
	docker-compose build

# Start all services in background
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Stop all services and remove volumes
clean:
	docker-compose down -v

# View logs of all services
logs:
	docker-compose logs -f

# Restart all services
restart: down up

# Test the API chain
test:
	@echo "Testing API health endpoints..."
	@curl -s http://localhost:8081/health | jq .
	@curl -s http://localhost:8082/health | jq .
	@curl -s http://localhost:8083/health | jq .
	@echo "\nTesting full API chain..."
	@curl -s -X POST http://localhost:3000/api/trigger | jq .

# Check health of all services
health:
	@echo "=== Service Health Check ==="
	@echo "API Service 1:"
	@curl -s http://localhost:8081/health | jq . 2>/dev/null || echo "❌ Service 1 not responding"
	@echo "\nAPI Service 2:"
	@curl -s http://localhost:8082/health | jq . 2>/dev/null || echo "❌ Service 2 not responding"
	@echo "\nAPI Service 3:"
	@curl -s http://localhost:8083/health | jq . 2>/dev/null || echo "❌ Service 3 not responding"
	@echo "\nWeb Service:"
	@curl -s http://localhost:3000/ > /dev/null && echo "✅ Web service is running" || echo "❌ Web service not responding"
	@echo "\nGrafana:"
	@curl -s http://localhost:3002/ > /dev/null && echo "✅ Grafana is running" || echo "❌ Grafana not responding"
	@echo "\nOTel Collector:"
	@curl -s http://localhost:13133/ > /dev/null && echo "✅ OpenTelemetry Collector is running" || echo "❌ OTel Collector not responding"
	@echo "\nTempo:"
	@curl -s http://localhost:3200/ > /dev/null && echo "✅ Tempo is running" || echo "❌ Tempo not responding"

# Show service URLs
urls:
	@echo "=== Service URLs ==="
	@echo "Web Application:     http://localhost:3000"
	@echo "API Service 1:       http://localhost:8081/health"
	@echo "API Service 2:       http://localhost:8082/health"
	@echo "API Service 3:       http://localhost:8083/health"
	@echo "Grafana Dashboard:   http://localhost:3002 (admin/admin)"
	@echo "OTel Collector:      http://localhost:8888/metrics"
	@echo "Tempo:               http://localhost:3200"

# Show help
help:
	@echo "=== Distributed Tracing Application ==="
	@echo "Available commands:"
	@echo "  make build    - Build all Docker images"
	@echo "  make up       - Start all services"
	@echo "  make down     - Stop all services"
	@echo "  make clean    - Stop services and remove volumes"
	@echo "  make restart  - Restart all services"
	@echo "  make logs     - Show logs from all services"
	@echo "  make health   - Check health of all services"
	@echo "  make test     - Test the API chain"
	@echo "  make urls     - Show service URLs"
	@echo "  make help     - Show this help message"