# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o taskflow ./cmd/taskflow-sabbavarapu

# Runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/taskflow /app/taskflow
EXPOSE 8080

CMD ["/app/taskflow"]
