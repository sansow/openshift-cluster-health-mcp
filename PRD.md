# Product Requirements Document: OpenShift Cluster Health MCP Server

**Project Name**: OpenShift Cluster Health MCP Server  
**Repository**: https://github.com/[your-org]/openshift-cluster-health-mcp  
**Parent Platform**: [OpenShift AI Ops Platform](https://github.com/[your-org]/openshift-aiops-platform)  
**License**: Apache 2.0  
**Language**: Go 1.21+  
**Status**: Planning / Phase 0  
**Version**: 0.1.0-alpha  
**Last Updated**: 2025-12-09

---

## Executive Summary

### Vision
A **lightweight, standards-compliant Model Context Protocol (MCP) server** that exposes OpenShift cluster health, monitoring, and operational data through natural language interfaces (OpenShift Lightspeed, Claude Desktop, and other MCP clients).

### Problem Statement
Current MCP server implementations for Kubernetes/OpenShift are either:
- **Too generic**: Don't integrate with OpenShift AI Ops workflows (Coordination Engine, KServe ML models)
- **Too complex**: Embedded with unnecessary dependencies (databases, workflow orchestration)
- **Wrong language**: TypeScript/Node.js doesn't align with Kubernetes-native Go ecosystem

### Solution
Build a **standalone Go-based MCP server** that:
1. Uses official **modelcontextprotocol/go-sdk** (Anthropic SDK)
2. Leverages proven patterns from **containers/kubernetes-mcp-server** (856 stars, Red Hat/containers team)
3. Integrates seamlessly with existing OpenShift AI Ops platform components
4. Provides natural language access to cluster health, ML-powered anomaly detection, and remediation workflows

### Target Users
| User Persona | Use Case | Priority |
|--------------|----------|----------|
| **OpenShift Platform Engineers** | Query cluster health via OpenShift Lightspeed | 🔴 Critical |
| **SREs and DevOps** | Trigger remediation workflows via AI assistants | 🔴 Critical |
| **Data Scientists** | Analyze anomalies using KServe ML models | 🟡 High |
| **Developers** | Integrate MCP tools in VS Code, Claude Desktop | 🟢 Medium |

---

## Project Context and References

### Parent Platform
This MCP server is designed to integrate with the **OpenShift AI Ops Self-Healing Platform**:
- **Repository**: https://github.com/[your-org]/openshift-aiops-platform
- **Documentation**: See `openshift-aiops-platform/AGENTS.md`
- **Architecture Decision**: [ADR-036: Go-Based Standalone MCP Server](https://github.com/[your-org]/openshift-aiops-platform/blob/main/docs/adrs/036-go-based-standalone-mcp-server.md)

### Integration Components
| Component | Language | Repository | Purpose |
|-----------|----------|------------|---------|
| **Coordination Engine** | Python/Flask | openshift-aiops-platform/src/coordination-engine | Remediation workflows, incident management |
| **KServe Models** | Python/Notebooks | openshift-aiops-platform/notebooks | ML-powered anomaly detection |
| **Prometheus** | - | OpenShift Monitoring | Cluster metrics and alerting |
| **Kubernetes API** | - | OpenShift 4.18+ | Cluster state and operations |

### Reference Implementations
| Project | Language | Stars | License | Usage |
|---------|----------|-------|---------|-------|
| [containers/kubernetes-mcp-server](https://github.com/containers/kubernetes-mcp-server) | Go | 856 | Apache 2.0 | Architecture reference, project structure patterns |
| [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) | Go | - | Apache 2.0 | Official MCP protocol implementation |
| [modelcontextprotocol/typescript-sdk](https://github.com/modelcontextprotocol/typescript-sdk) | TypeScript | - | MIT | Protocol specification reference |

---

## Goals and Success Metrics

### Primary Goals
1. **Standards Compliance**: 100% MCP protocol compliance using official Go SDK
2. **Lightweight**: <500 lines core server code, <10 direct dependencies
3. **Performance**: <100ms p95 tool response time, <50MB memory at rest
4. **Reusability**: Deployable on any OpenShift 4.14+ cluster
5. **Integration**: Seamless connection to Coordination Engine and KServe models

### Success Metrics
| Metric | Target | Measurement |
|--------|--------|-------------|
| **Build Time** | <30 seconds | OpenShift BuildConfig duration |
| **Binary Size** | <20MB | Compiled Go binary |
| **Container Image** | <50MB | Distroless image with binary |
| **Test Coverage** | >85% | Go test coverage report |
| **Tool Latency** | <100ms (p95) | Prometheus metrics |
| **Memory Usage** | <50MB at rest | Pod resource metrics |
| **Uptime** | >99.9% | OpenShift deployment status |

### Key Performance Indicators (KPIs)
- **Week 4**: MVP deployed to dev cluster with 4 working tools
- **Week 8**: Production deployment with OpenShift Lightspeed integration
- **Month 3**: Adopted by 3+ OpenShift clusters
- **Month 6**: External contributions from open source community

---

## User Stories and Acceptance Criteria

### US-1: OpenShift Lightspeed Integration (Critical Priority)
**As an** OpenShift platform engineer  
**I want to** ask OpenShift Lightspeed "What's my cluster health?"  
**So that** I get real-time status without manual `oc` commands

**Acceptance Criteria:**
- ✅ Lightspeed detects MCP server via OLSConfig
- ✅ Natural language queries return cluster health JSON
- ✅ Response includes: node status, pod health, resource utilization
- ✅ Follow-up questions maintain context
- ✅ Response time <2 seconds (p95)

**Technical Implementation:**
- MCP Tool: `get-cluster-health`
- Backend: Kubernetes API + Prometheus
- Transport: HTTP (StreamableHTTP for Lightspeed)

---

### US-2: ML-Powered Anomaly Detection (Critical Priority)
**As a** data scientist or SRE  
**I want to** ask "Are there anomalies in my Prometheus metrics?"  
**So that** I can quickly identify issues without writing PromQL queries

**Acceptance Criteria:**
- ✅ MCP tool calls KServe predictive-analytics-predictor model
- ✅ Returns anomaly scores with confidence levels (0-1)
- ✅ Explains findings in natural language
- ✅ Supports custom time ranges (1h, 6h, 24h, 7d)
- ✅ Handles missing KServe gracefully (fallback to rule-based)

**Technical Implementation:**
- MCP Tool: `analyze-anomalies`
- Backend: KServe InferenceService HTTP endpoint
- Fallback: Prometheus threshold-based rules
- Model: `http://predictive-analytics-predictor:8080/v1/models/predictive-analytics:predict`

---

### US-3: Remediation Workflow Triggering (Critical Priority)
**As an** SRE  
**I want to** ask "Trigger remediation for incident INC-12345"  
**So that** automated healing workflows execute without manual intervention

**Acceptance Criteria:**
- ✅ MCP tool delegates to Coordination Engine REST API
- ✅ Returns workflow execution status and ID
- ✅ Supports workflow types: restart_pod, scale_deployment, drain_node, etc.
- ✅ Validates incident ID exists before triggering
- ✅ Provides workflow progress tracking

**Technical Implementation:**
- MCP Tool: `trigger-remediation`
- Backend: Coordination Engine `POST /api/v1/remediation/trigger`
- Auth: ServiceAccount token for RBAC
- Response: Workflow ID, status, estimated completion time

---

### US-4: Cluster Resource Listing (High Priority)
**As a** developer or SRE  
**I want to** ask "List pods in namespace production"  
**So that** I can check application deployment status

**Acceptance Criteria:**
- ✅ MCP tool queries Kubernetes API directly
- ✅ Supports filtering by namespace, labels, field selectors
- ✅ Returns pod name, status, node, restarts, age
- ✅ Response time <500ms for <100 pods
- ✅ Graceful handling of RBAC permission errors

**Technical Implementation:**
- MCP Tool: `list-pods`
- Backend: Kubernetes API via client-go
- RBAC: ClusterRole with `pods.list` permission
- Caching: 30-second TTL for repeated queries

---

### US-5: Cluster Health Resource (High Priority)
**As an** AI assistant (OpenShift Lightspeed)  
**I want to** access cluster health as a resource URI `cluster://health`  
**So that** I can provide context-aware answers about cluster state

**Acceptance Criteria:**
- ✅ MCP Resource returns real-time cluster health snapshot
- ✅ Includes: node count, pod count, failed pods, resource utilization
- ✅ Delegates to Coordination Engine for aggregated health
- ✅ Fallback to Kubernetes API if Coordination Engine unavailable
- ✅ Response format: JSON with timestamp

**Technical Implementation:**
- MCP Resource: `cluster://health`
- Backend: Coordination Engine `GET /api/v1/cluster/status`
- Fallback: Direct Kubernetes API aggregation
- Cache: 10-second TTL

---

### US-6: Reusability Across Clusters (High Priority)
**As a** platform team  
**I want to** deploy this MCP server on multiple OpenShift clusters  
**So that** all clusters have consistent AI assistant capabilities

**Acceptance Criteria:**
- ✅ Standalone Helm chart with configurable values
- ✅ No hard-coded cluster-specific configuration
- ✅ RBAC auto-created via Helm chart
- ✅ Works with OpenShift 4.14+ (4.18.21 tested)
- ✅ Optional integrations (Coordination Engine, KServe)

**Technical Implementation:**
- Helm Chart: `charts/openshift-cluster-health-mcp/`
- Configuration: `values.yaml` with sensible defaults
- Deployment: `kubectl apply -k` or `helm install`
- Documentation: DEPLOYMENT.md with step-by-step guide

---

## Technical Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│  MCP Clients (OpenShift Lightspeed, Claude Desktop, VS Code)    │
└─────────────────────────┬────────────────────────────────────────┘
                          │ MCP Protocol (HTTP/stdio)
                          ▼
┌──────────────────────────────────────────────────────────────────┐
│  OpenShift Cluster Health MCP Server (Go)                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  MCP Server (modelcontextprotocol/go-sdk)                  │  │
│  │  ├─ Transport: StreamableHTTP (Lightspeed) / stdio (local)│  │
│  │  ├─ Tools: 5 tools (cluster-health, anomalies, etc.)      │  │
│  │  └─ Resources: 3 resources (cluster://, model://)         │  │
│  └────────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  HTTP Clients (Go net/http)                                │  │
│  │  ├─ CoordinationEngineClient                               │  │
│  │  ├─ KServeClient                                           │  │
│  │  ├─ PrometheusClient                                       │  │
│  │  └─ KubernetesClient (client-go)                          │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────┬─────────────┬─────────────┬─────────────┬──────────┘
              │ HTTP REST   │ HTTP REST   │ HTTP REST   │ K8s API
              ▼             ▼             ▼             ▼
┌───────────────────┐ ┌────────────┐ ┌────────────┐ ┌─────────────┐
│ Coordination Eng. │ │ KServe     │ │ Prometheus │ │ Kubernetes  │
│ (Python/Flask)    │ │ Predictor  │ │ (Metrics)  │ │ API Server  │
│ - Incidents       │ │ - Anomaly  │ │ - PromQL   │ │ - Pods      │
│ - Remediation     │ │   Detection│ │ - Alerts   │ │ - Nodes     │
│ - Workflows       │ │ - ML Model │ │            │ │ - Events    │
└───────────────────┘ └────────────┘ └────────────┘ └─────────────┘
```

### MCP Tools (5 Core Tools)

| Tool Name | Input Parameters | Output | Backend |
|-----------|-----------------|--------|---------|
| **get-cluster-health** | `namespace?: string` | Health metrics JSON | Prometheus + K8s API |
| **analyze-anomalies** | `metric: string`, `timeRange: string` | Anomaly list with confidence | KServe predictor |
| **trigger-remediation** | `incident: string`, `action: enum` | Workflow status | Coordination Engine |
| **list-pods** | `namespace: string`, `labels?: string` | Pod list JSON | Kubernetes API |
| **get-model-status** | `modelName: string` | Model metadata | KServe API |

### MCP Resources (3 Resources)

| Resource URI | Description | Data Source | Cache TTL |
|--------------|-------------|-------------|-----------|
| **cluster://health** | Real-time cluster health snapshot | Coordination Engine `/api/v1/cluster/status` | 10s |
| **cluster://nodes** | Node information with status | Kubernetes API `/api/v1/nodes` | 30s |
| **cluster://incidents** | Active incidents list | Coordination Engine `/api/v1/incidents` | 5s |

### Technology Stack

| Component | Technology | Version | Justification |
|-----------|-----------|---------|---------------|
| **Language** | Go | 1.21+ | Kubernetes-native, performance, single binary |
| **MCP SDK** | modelcontextprotocol/go-sdk | Latest | Official Anthropic SDK, protocol compliance |
| **HTTP Server** | net/http (stdlib) | stdlib | No dependencies, production-ready |
| **K8s Client** | k8s.io/client-go | v0.29+ | Official Kubernetes client library |
| **Logging** | slog (stdlib) | stdlib | Structured logging, no dependencies |
| **Metrics** | prometheus/client_golang | v1.18+ | Standard Prometheus metrics export |
| **Testing** | testing (stdlib) + testify | stdlib + v1.8+ | Unit and integration tests |
| **Build** | Go modules | - | Standard Go dependency management |
| **Container** | Distroless base | gcr.io/distroless/static | Minimal attack surface (~10MB) |

**Dependencies Intentionally Excluded:**
- ❌ Database (PostgreSQL, SQLite) - Stateless design
- ❌ ORM (GORM, sqlc) - No persistence needed
- ❌ Web framework (Gin, Echo) - stdlib net/http sufficient
- ❌ DI container - Simple manual DI

---

## Functional Requirements

### FR-1: MCP Protocol Support
- **FR-1.1**: Implement MCP protocol version 2025-03-26 (latest)
- **FR-1.2**: Support StreamableHTTP transport for OpenShift Lightspeed
- **FR-1.3**: Support stdio transport for Claude Desktop/local clients
- **FR-1.4**: Session management with `mcp-session-id` header
- **FR-1.5**: Root discovery endpoint returning server capabilities (JSON)

### FR-2: Kubernetes Integration
- **FR-2.1**: List cluster nodes with health status
- **FR-2.2**: List pods filtered by namespace/labels
- **FR-2.3**: Get cluster events (warnings, errors, normal)
- **FR-2.4**: Support both in-cluster (ServiceAccount) and kubeconfig auth
- **FR-2.5**: RBAC-compliant access control

### FR-3: Prometheus Integration
- **FR-3.1**: Query Prometheus metrics with PromQL
- **FR-3.2**: Get active alerts from Alertmanager
- **FR-3.3**: Calculate cluster resource utilization percentages
- **FR-3.4**: Support bearer token authentication for OpenShift monitoring

### FR-4: KServe Model Integration
- **FR-4.1**: Check InferenceService readiness status
- **FR-4.2**: Call model prediction endpoints with JSON payloads
- **FR-4.3**: Get model metadata (version, runtime, replicas)
- **FR-4.4**: Return prediction results with confidence scores
- **FR-4.5**: Graceful degradation if KServe unavailable

### FR-5: Coordination Engine Integration
- **FR-5.1**: Delegate remediation actions via REST API
- **FR-5.2**: Query incident history and status
- **FR-5.3**: Get workflow queue status
- **FR-5.4**: Trigger anomaly analysis workflows
- **FR-5.5**: Graceful degradation if Coordination Engine unavailable

---

## Non-Functional Requirements

### NFR-1: Performance
- **NFR-1.1**: Tool execution <100ms (p95), <200ms (p99)
- **NFR-1.2**: Resource queries <50ms (p95), <100ms (p99)
- **NFR-1.3**: Memory footprint <50MB at rest, <100MB under load
- **NFR-1.4**: CPU usage <0.2 cores at rest, <1 core under load
- **NFR-1.5**: Binary size <20MB (compiled)
- **NFR-1.6**: Container image <50MB (distroless)

### NFR-2: Scalability
- **NFR-2.1**: Support 20+ concurrent MCP sessions
- **NFR-2.2**: Handle 200+ requests/minute per replica
- **NFR-2.3**: Horizontal scaling via Kubernetes deployment replicas
- **NFR-2.4**: Stateless design (no session affinity required)

### NFR-3: Security
- **NFR-3.1**: Use ServiceAccount tokens for K8s API access
- **NFR-3.2**: No persistent storage of credentials or secrets
- **NFR-3.3**: RBAC-based access control for all operations
- **NFR-3.4**: No sensitive data in logs (mask secrets, tokens)
- **NFR-3.5**: Pass OpenShift security context constraints (SCC)
- **NFR-3.6**: Run as non-root user (UID 1000)
- **NFR-3.7**: Read-only root filesystem

### NFR-4: Reliability
- **NFR-4.1**: Graceful degradation if Prometheus unavailable
- **NFR-4.2**: Graceful degradation if Coordination Engine unavailable
- **NFR-4.3**: Health check endpoint `/health` for liveness/readiness probes
- **NFR-4.4**: 99.9% uptime SLA target
- **NFR-4.5**: Automatic pod restart on failure (Kubernetes)

### NFR-5: Observability
- **NFR-5.1**: Prometheus metrics at `/metrics` endpoint
- **NFR-5.2**: Structured JSON logging with levels (debug, info, warn, error)
- **NFR-5.3**: Request tracing with unique request IDs
- **NFR-5.4**: Tool execution duration metrics
- **NFR-5.5**: Error rate metrics per tool

### NFR-6: Testability
- **NFR-6.1**: 85%+ unit test coverage
- **NFR-6.2**: Integration tests with mock backends
- **NFR-6.3**: E2E tests with real OpenShift cluster
- **NFR-6.4**: Automated CI/CD with GitHub Actions

---

## Project Structure

```
openshift-cluster-health-mcp/
├── cmd/
│   └── mcp-server/
│       └── main.go                       # Server entry point
├── internal/
│   ├── server/
│   │   ├── server.go                     # MCP server setup (Go SDK)
│   │   ├── transport.go                  # HTTP/stdio transport logic
│   │   └── handlers.go                   # Request routing
│   ├── tools/                            # MCP tool implementations
│   │   ├── cluster_health.go
│   │   ├── analyze_anomalies.go
│   │   ├── trigger_remediation.go
│   │   ├── list_pods.go
│   │   └── model_status.go
│   ├── resources/                        # MCP resource handlers
│   │   ├── cluster_health.go
│   │   ├── nodes.go
│   │   └── incidents.go
│   └── config/
│       └── config.go                     # Configuration management
├── pkg/
│   ├── clients/                          # HTTP clients for integrations
│   │   ├── coordination_engine.go        # Coordination Engine client
│   │   ├── kserve.go                     # KServe InferenceService client
│   │   ├── prometheus.go                 # Prometheus PromQL client
│   │   └── kubernetes.go                 # Kubernetes API client wrapper
│   └── models/
│       └── types.go                      # Common data structures
├── charts/
│   └── openshift-cluster-health-mcp/     # Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       ├── templates/
│       │   ├── deployment.yaml
│       │   ├── service.yaml
│       │   ├── serviceaccount.yaml
│       │   ├── rbac.yaml
│       │   └── configmap.yaml
│       └── README.md
├── test/
│   ├── unit/                             # Unit tests
│   ├── integration/                      # Integration tests
│   └── e2e/                              # End-to-end tests
├── docs/
│   ├── ARCHITECTURE.md                   # Technical architecture
│   ├── DEPLOYMENT.md                     # Deployment guide
│   ├── API.md                            # MCP tools/resources reference
│   └── TROUBLESHOOTING.md                # Common issues
├── .github/
│   └── workflows/
│       ├── ci.yml                        # CI/CD pipeline
│       └── release.yml                   # Release automation
├── Dockerfile                            # Container image build
├── Makefile                              # Build automation
├── go.mod                                # Go module definition
├── go.sum                                # Go module checksums
├── PRD.md                                # This document
└── README.md                             # Project overview
```

---

## Integration Requirements

### Required External Services

| Service | Endpoint | Protocol | Required | Fallback |
|---------|----------|----------|----------|----------|
| **Kubernetes API** | `https://kubernetes.default.svc` | K8s client-go | ✅ Yes | None (critical) |
| **Prometheus** | `https://prometheus-k8s.openshift-monitoring.svc:9091` | HTTP REST | ❌ No | Threshold-based rules |
| **Coordination Engine** | `http://coordination-engine:8080/api/v1/` | HTTP REST | ❌ No | Direct K8s API |
| **KServe Predictors** | `http://{model}-predictor:8080/v1/models/` | HTTP REST | ❌ No | Disabled anomaly detection |

### RBAC Requirements

```yaml
# Minimum ClusterRole permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: openshift-cluster-health-mcp
rules:
  # Core Kubernetes resources (read-only)
  - apiGroups: [""]
    resources: ["nodes", "pods", "events", "namespaces", "services"]
    verbs: ["get", "list", "watch"]
  
  # Deployments and workloads (read-only)
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list", "watch"]
  
  # KServe models (read-only)
  - apiGroups: ["serving.kserve.io"]
    resources: ["inferenceservices", "inferenceservices/status"]
    verbs: ["get", "list", "watch"]
  
  # Prometheus API access
  - apiGroups: [""]
    resources: ["services/prometheus-k8s"]
    verbs: ["get"]
    resourceNames: ["prometheus-k8s"]
```

---

## Out of Scope (Phase 1)

The following features are **intentionally excluded** from Phase 1:

### ❌ Not Included
- **Database persistence**: Stateless design, no PostgreSQL/SQLite
- **Workflow orchestration**: Delegated to Coordination Engine
- **REST API endpoints**: MCP protocol only (no custom HTTP API)
- **Custom authentication**: Uses Kubernetes ServiceAccount RBAC
- **Model training**: Only inference via KServe (training in notebooks)
- **Log aggregation**: Uses OpenShift logging infrastructure
- **GitOps workflows**: Managed by ArgoCD separately
- **Multi-cluster management**: Single cluster scope (Phase 2)
- **Kiali integration**: Not required for MVP (containers/kubernetes-mcp-server feature)
- **KubeVirt VMs**: Not required for MVP (containers/kubernetes-mcp-server feature)

---

## Implementation Phases

### Phase 0: Project Setup (Week 1)
**Duration**: 3-5 days  
**Goal**: Repository structure, build system, basic CI/CD

**Deliverables**:
- ✅ GitHub repository created with Apache 2.0 license
- ✅ Go modules initialized (`go.mod`, `go.sum`)
- ✅ Project structure (cmd/, internal/, pkg/, charts/)
- ✅ Makefile with build/test/lint targets
- ✅ Dockerfile with distroless base image
- ✅ GitHub Actions CI workflow
- ✅ README.md with quick start guide

---

### Phase 1: MVP Implementation (Week 1-2)
**Duration**: 7-10 days  
**Goal**: Working MCP server with 4 core tools

**Tasks**:
1. **MCP Server Setup** (2 days)
   - Implement `cmd/mcp-server/main.go` with Go SDK
   - Add stdio and HTTP transport support
   - Root discovery endpoint (`/`)
   - Health check endpoint (`/health`)

2. **HTTP Clients** (2 days)
   - `KubernetesClient` (client-go wrapper)
   - `PrometheusClient` (PromQL queries)
   - `CoordinationEngineClient` (REST API)
   - `KServeClient` (InferenceService API)

3. **MCP Tools** (3 days)
   - `get-cluster-health` (Prometheus + K8s)
   - `list-pods` (Kubernetes API)
   - `analyze-anomalies` (KServe predictor)
   - `trigger-remediation` (Coordination Engine)

4. **MCP Resources** (2 days)
   - `cluster://health` (Coordination Engine)
   - `cluster://nodes` (Kubernetes API)
   - `cluster://incidents` (Coordination Engine)

5. **Testing** (2 days)
   - Unit tests for all tools and clients (>80% coverage)
   - Mock backends for integration tests
   - Test with `mcp-inspector` tool

**Deliverables**:
- ✅ Working MCP server binary
- ✅ 4 tools, 3 resources implemented
- ✅ Unit tests with >80% coverage
- ✅ Dockerfile builds successfully

---

### Phase 2: Integration Testing (Week 3)
**Duration**: 5-7 days  
**Goal**: Deploy to dev cluster, test with real backends

**Tasks**:
1. **Helm Chart** (2 days)
   - Chart structure and templates
   - values.yaml with sensible defaults
   - RBAC manifests (ServiceAccount, ClusterRole)
   - ConfigMap for configuration

2. **OpenShift Deployment** (2 days)
   - Deploy to dev cluster
   - Configure RBAC permissions
   - Test with real Prometheus
   - Test with real Coordination Engine

3. **OpenShift Lightspeed Integration** (2 days)
   - OLSConfig creation
   - Test natural language queries
   - Verify tool execution
   - Test resource access

4. **E2E Tests** (1 day)
   - Test suite against real cluster
   - Verify all integration points
   - Performance benchmarks

**Deliverables**:
- ✅ Helm chart for deployment
- ✅ Deployed to dev cluster
- ✅ OpenShift Lightspeed integration working
- ✅ E2E test suite passing

---

### Phase 3: Production Hardening (Week 4)
**Duration**: 5-7 days  
**Goal**: Production-ready deployment

**Tasks**:
1. **Security** (2 days)
   - Security scanning (gosec, trivy)
   - Non-root user enforcement
   - Read-only root filesystem
   - Secret management validation

2. **Observability** (2 days)
   - Prometheus metrics export
   - Structured logging with levels
   - Request tracing (request IDs)
   - Error rate monitoring

3. **Reliability** (2 days)
   - Graceful degradation testing
   - Health probe tuning
   - Resource limit recommendations
   - Failure injection testing

4. **Performance** (1 day)
   - Load testing (200 req/min)
   - Memory profiling
   - CPU profiling
   - Latency optimization

**Deliverables**:
- ✅ Security scan passing (zero critical vulnerabilities)
- ✅ Prometheus metrics dashboard
- ✅ Performance benchmarks documented
- ✅ Production deployment guide

---

### Phase 4: Documentation and Release (Week 5)
**Duration**: 3-5 days  
**Goal**: Public release and documentation

**Tasks**:
1. **Documentation** (2 days)
   - ARCHITECTURE.md (technical design)
   - DEPLOYMENT.md (step-by-step guide)
   - API.md (tool/resource reference)
   - TROUBLESHOOTING.md (common issues)

2. **Release Preparation** (2 days)
   - GitHub release with binaries
   - Container image push to Quay.io
   - Helm chart publication
   - CHANGELOG.md

3. **Community** (1 day)
   - CONTRIBUTING.md
   - Code of Conduct
   - GitHub issue templates
   - PR template

**Deliverables**:
- ✅ v0.1.0 release on GitHub
- ✅ Container image on Quay.io
- ✅ Comprehensive documentation
- ✅ Public announcement

---

### Phase 5: TypeScript Deprecation (Month 2-6)
**Duration**: 4 months (parallel with other work)  
**Goal**: Migrate from TypeScript MCP server

**Timeline**:
- **Month 2**: Run both servers in parallel, verify feature parity
- **Month 3**: Migrate Lightspeed integration to Go server
- **Month 4**: Internal testing, gather feedback
- **Month 5**: Deprecate TypeScript server (announce EOL)
- **Month 6**: Archive TypeScript implementation

**Deliverables**:
- ✅ Migration guide for existing users
- ✅ Feature parity validation
- ✅ TypeScript server deprecated
- ✅ Updated platform documentation

---

## Risk Assessment and Mitigation

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| **Go SDK immaturity** | Low | Medium | Use containers/kubernetes-mcp-server as proof of viability |
| **OpenShift Lightspeed compatibility** | Medium | High | Test early with StreamableHTTP transport, follow official docs |
| **Coordination Engine downtime** | Medium | Medium | Implement graceful degradation, direct K8s API fallback |
| **Performance degradation** | Low | Medium | Benchmark early, profile regularly, optimize hot paths |
| **RBAC permission issues** | High | High | Document minimal permissions, test with restricted ServiceAccount |
| **Team Go expertise gap** | Medium | Low | Leverage containers/kubernetes-mcp-server as reference, 1-week learning curve |
| **Timeline overruns** | Medium | Low | Phased approach, MVP-first, feature flags for optional integrations |

---

## Testing Strategy

### Unit Tests (>85% coverage target)
- **Tools**: Test each MCP tool with mocked clients
- **Resources**: Test each MCP resource with mocked backends
- **Clients**: Test HTTP clients with mock servers
- **Edge Cases**: Error handling, timeouts, invalid inputs

### Integration Tests
- **Mock Backends**: httptest servers for Coordination Engine, KServe
- **Real K8s API**: Use kind/minikube for local testing
- **MCP Protocol**: Use `mcp-inspector` for protocol validation

### E2E Tests
- **Real Cluster**: Deploy to dev OpenShift cluster
- **Lightspeed**: Test with actual OpenShift Lightspeed
- **Performance**: Load testing with k6 or hey
- **Chaos**: Failure injection (kill pods, network delays)

### CI/CD Pipeline
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: make lint test coverage
      - run: make build
      - run: make docker-build
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make security-scan
```

---

## Deployment Models

### Mode 1: Standalone (Minimal)
**Use Case**: Read-only cluster monitoring, no AI Ops integration

**Components**:
- MCP server + Kubernetes API only
- No Coordination Engine, no KServe
- Tools: `get-cluster-health`, `list-pods`
- Resources: `cluster://nodes`

**Deployment**:
```bash
helm install cluster-health-mcp ./charts/openshift-cluster-health-mcp \
  --set coordinationEngine.enabled=false \
  --set kserve.enabled=false
```

---

### Mode 2: With Coordination Engine
**Use Case**: Self-healing platform with remediation workflows

**Components**:
- MCP server + Coordination Engine integration
- Tools: `trigger-remediation`, `get-cluster-health`
- Resources: `cluster://health`, `cluster://incidents`

**Deployment**:
```bash
helm install cluster-health-mcp ./charts/openshift-cluster-health-mcp \
  --set coordinationEngine.enabled=true \
  --set coordinationEngine.url=http://coordination-engine:8080
```

---

### Mode 3: Full Stack (OpenShift AI Ops)
**Use Case**: Complete AI Ops platform with ML-powered anomaly detection

**Components**:
- MCP server + Coordination Engine + KServe models
- All tools and resources enabled
- ML-powered anomaly detection

**Deployment**:
```bash
helm install cluster-health-mcp ./charts/openshift-cluster-health-mcp \
  --set coordinationEngine.enabled=true \
  --set kserve.enabled=true \
  --set kserve.namespace=self-healing-platform
```

---

## Success Criteria Summary

### Phase 1 Success (Week 2)
- ✅ MCP server compiles and runs
- ✅ 4 tools work with mocked backends
- ✅ Unit tests >80% coverage
- ✅ Docker image builds (<50MB)

### Phase 2 Success (Week 3)
- ✅ Deployed to dev cluster
- ✅ OpenShift Lightspeed detects server
- ✅ Natural language queries work
- ✅ Integration with real backends

### Phase 3 Success (Week 4)
- ✅ Security scan passing
- ✅ Prometheus metrics exported
- ✅ Performance benchmarks met
- ✅ Production RBAC configured

### Phase 4 Success (Week 5)
- ✅ v0.1.0 released on GitHub
- ✅ Documentation complete
- ✅ Public announcement
- ✅ Community engagement started

---

## Appendix

### A. Coordination Engine API Reference
See: https://github.com/[your-org]/openshift-aiops-platform/blob/main/src/coordination-engine/README.md

**Key Endpoints**:
- `GET /api/v1/cluster/status` - Cluster health snapshot
- `GET /api/v1/incidents` - Active incidents list
- `POST /api/v1/remediation/trigger` - Trigger remediation workflow

### B. KServe Model Inference Reference
**Predictive Analytics Model**:
- Endpoint: `http://predictive-analytics-predictor:8080/v1/models/predictive-analytics:predict`
- Input: `{"instances": [{"metric": "cpu_usage", "values": [0.8, 0.9, 0.95]}]}`
- Output: `{"predictions": [{"anomaly_score": 0.87, "confidence": 0.92}]}`

### C. containers/kubernetes-mcp-server Comparison

| Feature | containers/kubernetes-mcp-server | openshift-cluster-health-mcp |
|---------|----------------------------------|------------------------------|
| **Language** | Go | Go ✅ |
| **Kubernetes Tools** | ✅ Comprehensive | ✅ Basic (pods, nodes) |
| **Kiali Integration** | ✅ Yes | ❌ Out of scope |
| **KubeVirt VMs** | ✅ Yes | ❌ Out of scope |
| **Prometheus** | ❌ No | ✅ Yes (PromQL) |
| **Custom Integration** | ❌ No | ✅ Coordination Engine + KServe |
| **ML Models** | ❌ No | ✅ KServe inference |
| **Use Case** | General K8s tooling | OpenShift AI Ops platform |

### D. MCP Protocol Resources
- **Specification**: https://spec.modelcontextprotocol.io/
- **Go SDK**: https://github.com/modelcontextprotocol/go-sdk
- **TypeScript SDK** (reference): https://github.com/modelcontextprotocol/typescript-sdk
- **MCP Inspector**: https://github.com/modelcontextprotocol/inspector

---

**Document Version**: 1.0  
**Last Updated**: 2025-12-09  
**Next Review**: 2025-12-16 (after Phase 0 completion)  
**Owner**: Platform Architecture Team  
**Approvers**: [Pending]

