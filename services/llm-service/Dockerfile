# LLM Service - Dedicated container for AI analysis
FROM golang:1.23-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git curl

# Set working directory
WORKDIR /app

# Initialize go module
RUN go mod init llm-service

# Copy service code
COPY main.go .

# Create minimal go.mod for dependencies
RUN echo 'module llm-service

go 1.23

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/rs/zerolog v1.31.0
)' > go.mod

# Download dependencies
RUN go mod download
RUN go mod tidy

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -o llm-service main.go

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl bash

# Create non-root user
RUN adduser -D -s /bin/sh -u 1002 llm

# Create directories
RUN mkdir -p /app && chown -R llm:llm /app

# Copy binary from builder
COPY --from=builder /app/llm-service /usr/local/bin/llm-service
RUN chmod +x /usr/local/bin/llm-service

# Switch to non-root user
USER llm

# Set working directory
WORKDIR /app

# Environment variables
ENV PORT=8082 \
    OLLAMA_URL=http://ollama:11434 \
    OLLAMA_MODEL=phi3:mini

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

# Run the service
CMD ["/usr/local/bin/llm-service"]