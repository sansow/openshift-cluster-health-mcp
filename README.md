# OpenShift Cluster Health MCP Server

Model Context Protocol (MCP) server for OpenShift cluster health monitoring and AI Ops integration. Provides real-time cluster health data, incident management, and ML-powered anomaly detection through a standardized MCP interface.

![CI Status](https://github.com/openshift-aiops/openshift-cluster-health-mcp/workflows/CI/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/openshift-aiops/openshift-cluster-health-mcp)](https://goreportcard.com/report/github.com/openshift-aiops/openshift-cluster-health-mcp)

## Features

- **MCP Tools**: 6 tools for cluster operations and AI-powered analysis
  - `get-cluster-health` - Real-time cluster health snapshot
  - `list-pods` - Pod listing with advanced filtering
  - `list-incidents` - Active incident tracking via Coordination Engine
  - `trigger-remediation` - Automated remediation actions
  - `analyze-anomalies` - ML-powered anomaly detection via KServe
  - `get-model-status` - KServe model health monitoring

- **MCP Resources**: 3 resources for passive data access
  - `cluster://health` - Real-time cluster health (10s cache)
  - `cluster://nodes` - Node information and capacity (30s cache)
  - `cluster://incidents` - Active incidents from Coordination Engine (5s cache)

- **Integrations**:
  - ✅ Kubernetes API (required)
  - ✅ Coordination Engine (optional - incident management)
  - ✅ KServe (optional - ML model serving)
  - ✅ Prometheus (optional - enhanced metrics)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  OpenShift Lightspeed / AI Assistant                    │
│  (MCP Client)                                           │
└─────────────────┬───────────────────────────────────────┘
                  │ HTTP/SSE (MCP Protocol)
┌─────────────────▼───────────────────────────────────────┐
│  OpenShift Cluster Health MCP Server                    │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │ MCP Tools   │  │ MCP Resources│  │ Cache (30s TTL)│ │
│  │ (6 total)   │  │ (3 total)    │  │                │ │
│  └─────────────┘  └──────────────┘  └────────────────┘ │
└──────┬──────────────┬──────────────┬──────────────┬─────┘
       │              │              │              │
┌──────▼──────┐ ┌────▼─────┐ ┌──────▼──────┐ ┌────▼────────┐
│ Kubernetes  │ │Coordination│ │   KServe    │ │ Prometheus  │
│ API         │ │ Engine     │ │  (ML Models)│ │   (Metrics) │
└─────────────┘ └──────────┘ └─────────────┘ └─────────────┘
```

## Version Compatibility

| OpenShift Version | Kubernetes Version | Container Image | Status |
|-------------------|-------------------|-----------------|--------|
| **4.18** | 1.31 | `quay.io/takinosh/openshift-cluster-health-mcp:4.18-latest` | ✅ Supported |
| **4.19** | 1.31 | `quay.io/takinosh/openshift-cluster-health-mcp:4.19-latest` | ✅ Supported |
| **4.20** | 1.33 | `quay.io/takinosh/openshift-cluster-health-mcp:4.20-latest` | ✅ Supported |

**📚 [Complete Installation Guide](./docs/INSTALLATION.md)** - Detailed instructions for installing the MCP server for each OpenShift version.

## Quick Start

### Prerequisites

- OpenShift 4.18+ (see version compatibility table above)
- Go 1.24+ (for local development)
- Helm 3.0+
- kubectl/oc CLI

### Local Development

```bash
# Clone repository
git clone https://github.com/openshift-aiops/openshift-cluster-health-mcp.git
cd openshift-cluster-health-mcp

# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Run locally (HTTP transport - default)
MCP_TRANSPORT=http ./bin/mcp-server

# Test with curl
curl http://localhost:8080/health
```

### Deploy to OpenShift

**Option 1: Use Pre-built Container Images (Recommended)**

```bash
# Determine your OpenShift version
oc version

# Install with Helm using the matching image tag
helm install cluster-health-mcp ./charts/openshift-cluster-health-mcp \
  --namespace cluster-health-mcp \
  --create-namespace \
  --set image.repository=quay.io/takinosh/openshift-cluster-health-mcp \
  --set image.tag=4.20-latest  # Use 4.18-latest, 4.19-latest, or 4.20-latest

# Verify deployment
oc get pods -n cluster-health-mcp
oc logs -l app.kubernetes.io/name=openshift-cluster-health-mcp -n cluster-health-mcp
```

**Option 2: Build from Source**

```bash
# Build container image
make docker-build
oc new-build --name cluster-health-mcp --binary --strategy docker -n <namespace>
oc start-build cluster-health-mcp --from-dir=. --follow -n <namespace>

# Deploy with Helm
helm install cluster-health-mcp ./charts/openshift-cluster-health-mcp \
  --namespace <namespace>
```

**📖 For detailed installation instructions, version-specific configurations, and troubleshooting, see the [Installation Guide](./docs/INSTALLATION.md).**

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MCP_TRANSPORT` | Transport mode (http or stdio) | `http` | Yes |
| `MCP_HTTP_PORT` | HTTP server port | `8080` | If HTTP |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` | No |
| `LOG_FORMAT` | Log format (json or text) | `json` | No |
| `ENABLE_COORDINATION_ENGINE` | Enable Coordination Engine integration | `false` | No |
| `COORDINATION_ENGINE_URL` | Coordination Engine endpoint | - | If CE enabled |
| `ENABLE_KSERVE` | Enable KServe integration | `false` | No |
| `KSERVE_NAMESPACE` | Namespace for KServe models | - | If KServe enabled |
| `ENABLE_PROMETHEUS` | Enable Prometheus integration | `false` | No |
| `PROMETHEUS_URL` | Prometheus endpoint | - | If Prom enabled |

### Helm Values

See `charts/openshift-cluster-health-mcp/values.yaml` for full configuration options.

**Key configurations**:

```yaml
# Enable all integrations (dev example)
integrations:
  coordinationEngine:
    enabled: true
    url: http://coordination-engine:8080

  kserve:
    enabled: true
    namespace: self-healing-platform

  prometheus:
    enabled: true
    url: https://prometheus-k8s.openshift-monitoring.svc:9091

# Security context (OpenShift compatible)
podSecurityContext:
  runAsNonRoot: true
  seccompProfile:
    type: RuntimeDefault
```

## Usage

### HTTP Endpoints

```bash
# Health check
curl http://localhost:8080/health

# List available tools
curl http://localhost:8080/mcp/tools

# List available resources
curl http://localhost:8080/mcp/resources

# Get cluster health resource
curl http://localhost:8080/mcp/resources/cluster/health

# Execute tool
curl -X POST http://localhost:8080/mcp/tools/get-cluster-health \
  -H 'Content-Type: application/json' \
  -d '{}'
```

### MCP Client Integration

```typescript
// Example: Connect to MCP server
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { SSEClientTransport } from "@modelcontextprotocol/sdk/client/sse.js";

const transport = new SSEClientTransport(
  new URL("http://cluster-health-mcp:8080/mcp/sse")
);

const client = new Client({
  name: "openshift-lightspeed",
  version: "1.0.0"
}, {
  capabilities: {}
});

await client.connect(transport);

// List tools
const tools = await client.listTools();
console.log("Available tools:", tools);

// Execute tool
const result = await client.callTool({
  name: "get-cluster-health",
  arguments: {}
});
console.log("Cluster health:", result);
```

## Development

### Project Structure

```
openshift-cluster-health-mcp/
├── cmd/
│   └── mcp-server/          # Main entry point
├── internal/
│   ├── server/              # HTTP server and MCP protocol handling
│   ├── tools/               # MCP tool implementations
│   └── resources/           # MCP resource implementations
├── pkg/
│   ├── clients/             # External API clients (K8s, CE, KServe)
│   └── cache/               # Caching layer
├── charts/
│   └── openshift-cluster-health-mcp/  # Helm chart
├── .github/
│   └── workflows/           # GitHub Actions CI/CD
├── docs/
│   └── adrs/                # Architecture Decision Records
└── test/                    # Test files
```

### Running Tests

```bash
# Unit tests
make test

# Unit tests with coverage
make test-coverage

# Lint code
make lint

# Security scan
make security-scan
```

### Build Options

```bash
# Local development build
make build

# Production build (optimized)
make build-prod

# Docker build
make docker-build

# Multi-arch build
make docker-buildx
```

## Production Deployment

### Security Considerations

- **RBAC**: ServiceAccount with minimal ClusterRole permissions (read-only)
- **Security Context**: Runs as nonroot user with read-only filesystem
- **Network Policies**: Optional network isolation
- **Image**: Based on Red Hat UBI 9 Micro (minimal attack surface)

### High Availability

```yaml
# Helm values for HA deployment
replicaCount: 2

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: openshift-cluster-health-mcp
        topologyKey: kubernetes.io/hostname
```

### Monitoring

The server exposes Prometheus metrics at `/metrics`:

- `mcp_tools_total` - Total tool executions
- `mcp_tools_errors_total` - Tool execution errors
- `mcp_resources_requests_total` - Resource access requests
- `mcp_cache_hits_total` - Cache hits
- `mcp_cache_misses_total` - Cache misses

## Troubleshooting

### Common Issues

**Pods not starting (SCC violations)**:
```bash
# Check security context constraints
oc get pods -o yaml | grep -A 10 securityContext

# Ensure podSecurityContext.runAsUser is not set (let OpenShift assign)
```

**Integration failures**:
```bash
# Check logs
oc logs -l app.kubernetes.io/name=openshift-cluster-health-mcp

# Verify service connectivity
oc exec <pod-name> -- curl -I http://coordination-engine:8080/health
```

**Cache issues**:
```bash
# Check cache statistics in logs
oc logs <pod-name> | grep -i cache
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Apache License 2.0

## Support

- Issues: [GitHub Issues](https://github.com/openshift-aiops/openshift-cluster-health-mcp/issues)
- Discussions: [GitHub Discussions](https://github.com/openshift-aiops/openshift-cluster-health-mcp/discussions)
- Documentation: [docs/](./docs/)

## Acknowledgments

- Built on the [Model Context Protocol](https://modelcontextprotocol.io/)
- Integrates with [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift)
- Powered by [KServe](https://kserve.github.io/website/) for ML inference
