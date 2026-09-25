# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Install certificates and timezone data needed by the build and runtime.
RUN apk add --no-cache ca-certificates tzdata

# Set GOPROXY for better reliability
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Download dependencies before copying source to keep this layer cacheable.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" \
    -o /build/task-server ./main.go

# Stage 2: Runtime
FROM alpine:3.20.3

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

RUN addgroup -S app && adduser -S -G app app

# Copy binary from builder
COPY --from=builder /build/task-server .

# Create data directory (if needed)
RUN mkdir -p /data && chown -R app:app /app /data

USER app

# Expose HTTP port
EXPOSE 9002

# Run the binary
CMD ["./task-server"]