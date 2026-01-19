# ---------------------------
# Stage 1: Build Go binary
# ---------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git if needed for private packages
RUN apk add --no-cache git

# Copy go.mod and download dependencies
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o dashboard main.go

# ---------------------------
# Stage 2: Runtime container
# ---------------------------
FROM alpine:latest

WORKDIR /app

# Copy binary and static/template files
COPY --from=builder /app/dashboard .
COPY --from=builder /app/template ./template
COPY --from=builder /app/static ./static

# Optional: Set timezone
RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

EXPOSE 2097

# Start the dashboard
CMD ["./dashboard"]
