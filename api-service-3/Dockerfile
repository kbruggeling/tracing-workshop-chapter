FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o api-service-3 main.go

EXPOSE 8080

CMD ["./api-service-3"]