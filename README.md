# Distributed Tracing Application

This project demonstrates a distributed tracing setup with multiple microservices. It includes a web service with a button that triggers a chain of API calls through three Go services to fetch data from a PostgreSQL database.

## Architecture

The application follows this flow when the button is clicked:

**Web Service** → **API 1** → **API 2** → **API 3** → **PostgreSQL Database**

1. **Web Service** (Node.js/Express): Serves a simple HTML page with a button
2. **API Service 1** (Go): Receives request from web service, calls API Service 2
3. **API Service 2** (Go): Receives request from API 1, calls API Service 3  
4. **API Service 3** (Go): Receives request from API 2, queries the database
5. **PostgreSQL Database**: Contains sample user data
6. **Tempo**: Distributed tracing backend (ready for future tracing implementation)
7. **Grafana**: Visualization dashboard for traces (ready for future use)

## Quick Start

### Prerequisites

- Docker
- Docker Compose

### Running the Application

1. Clone this repository
2. Navigate to the project directory
3. Start all services:

```bash
docker-compose up --build
```

4. Access the application:
   - **Web Application**: http://localhost:3000
   - **Grafana Dashboard**: http://localhost:3001 (admin/admin)
   - **Individual APIs**:
     - API 1: http://localhost:8081/health
     - API 2: http://localhost:8082/health  
     - API 3: http://localhost:8083/health

### Testing the API Chain

1. Open http://localhost:3000 in your browser
2. Click the "Trigger API Chain" button
3. The response will show user data fetched through the complete chain

## Service Details

### Web Service
- **Port**: 3000
- **Technology**: Node.js, Express
- **Endpoints**:
  - `GET /`: Serves the main HTML page
  - `POST /api/trigger`: Triggers the API chain

### API Services (1, 2, 3)
- **Ports**: 8081, 8082, 8083
- **Technology**: Go
- **Endpoints**:
  - `GET /health`: Health check
  - `GET /api/data`: Main data endpoint (chains to next service)

### Database
- **Port**: 5432
- **Technology**: PostgreSQL
- **Database**: testdb
- **Credentials**: testuser/testpass
- **Sample Data**: 10 users with names and emails

## Development

### Project Structure

```
distributed-tracing-app/
├── web-service/          # Node.js web application
├── api-service-1/        # Go API service 1
├── api-service-2/        # Go API service 2  
├── api-service-3/        # Go API service 3
├── database/             # PostgreSQL setup
├── tempo/                # Tempo tracing configuration
├── grafana/              # Grafana configuration
├── docker-compose.yml    # Docker orchestration
└── README.md
```

### Adding Distributed Tracing

The application is ready for distributed tracing implementation:

- **Tempo** is configured and running on port 3200
- **Grafana** is configured with Tempo as a datasource
- OpenTelemetry instrumentation can be added to each service

### Stopping the Application

```bash
docker-compose down
```

To also remove volumes:

```bash
docker-compose down -v
```

## Troubleshooting

- If services fail to connect, ensure all containers are running: `docker-compose ps`
- Check logs for individual services: `docker-compose logs <service-name>`
- Database connection issues: Verify the database is ready before APIs start

This project is licensed under the MIT License. See the LICENSE file for details.