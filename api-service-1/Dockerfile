FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o api-service-1 main.go

FROM gcr.io/distroless/base

COPY --from=builder /app/api-service-1 /api-service-1

CMD ["/api-service-1"]