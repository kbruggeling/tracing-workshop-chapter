FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o api-service-2 main.go

FROM gcr.io/distroless/base

COPY --from=builder /app/api-service-2 /api-service-2

CMD ["/api-service-2"]