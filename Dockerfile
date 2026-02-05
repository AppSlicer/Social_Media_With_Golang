# Stage 1: Build
FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server/main.go

# Stage 2: Production
FROM alpine:latest
WORKDIR /app
COPY --from=builder /server /app/server
COPY .env .
COPY firebase_credentials.json .
EXPOSE 8080
CMD ["/app/server"]
