# Dockerfile - Multi-Stage Build for OpenShift Cluster Health MCP Server
# Implements ADR-008: Distroless Container Images
# Target: Container image <50MB

# Stage 1: Build stage (full Go environment)
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy Go modules files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary with size optimizations
# CGO_ENABLED=0: Fully static linking (no C dependencies)
# -ldflags="-s -w": Strip debug info and symbol table (reduces binary size)
# -trimpath: Remove file system paths from executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /build/mcp-server \
    ./cmd/mcp-server

# Verify binary is static
RUN file /build/mcp-server
RUN ls -lh /build/mcp-server

# Stage 2: Runtime stage (distroless)
FROM gcr.io/distroless/static:nonroot

# Copy binary from builder
COPY --from=builder /build/mcp-server /usr/local/bin/mcp-server

# Distroless uses nonroot user (UID 65532) by default
USER nonroot:nonroot

# MCP HTTP transport port (if using HTTP mode)
EXPOSE 8080

# Health check endpoint (if implemented)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD ["/usr/local/bin/mcp-server", "healthcheck"] || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/mcp-server"]
