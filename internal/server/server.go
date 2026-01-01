package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/internal/resources"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/internal/tools"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/pkg/cache"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/pkg/clients"
)

// MCPServer wraps the official MCP SDK server
type MCPServer struct {
	config     *Config
	mcpServer  *mcp.Server
	httpServer *http.Server
	k8sClient  *clients.K8sClient
	ceClient   *clients.CoordinationEngineClient
	kserve     *clients.KServeClient
	cache      *cache.MemoryCache
	tools      map[string]interface{} // Registry of available tools
	resources  map[string]interface{} // Registry of available resources
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(config *Config) (*MCPServer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize Kubernetes client
	k8sClient, err := clients.NewK8sClient(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Verify cluster connectivity
	ctx := context.Background()
	if err := k8sClient.HealthCheck(ctx); err != nil {
		log.Printf("WARNING: Kubernetes health check failed: %v", err)
		log.Printf("Server will start but cluster health tools may not work")
	} else {
		version, _ := k8sClient.GetServerVersion(ctx)
		log.Printf("Connected to Kubernetes cluster (version: %s)", version)
	}

	// Initialize cache with configured TTL
	memoryCache := cache.NewMemoryCache(config.CacheTTL)
	log.Printf("Initialized cache with TTL: %s", config.CacheTTL)

	// Initialize Coordination Engine client if enabled
	var ceClient *clients.CoordinationEngineClient
	if config.EnableCoordinationEngine {
		ceClient = clients.NewCoordinationEngineClient(config.CoordinationEngineURL)
		log.Printf("Initialized Coordination Engine client: %s", config.CoordinationEngineURL)
	} else {
		log.Printf("Coordination Engine integration disabled (use ENABLE_COORDINATION_ENGINE=true to enable)")
	}

	// Initialize KServe client if enabled
	var kserveClient *clients.KServeClient
	if config.EnableKServe {
		kserveClient = clients.NewKServeClient(clients.KServeConfig{
			Namespace:  config.KServeNamespace,
			Timeout:    config.RequestTimeout,
			Enabled:    true,
			RestConfig: k8sClient.GetConfig(), // Pass Kubernetes config for CRD access
		})
		log.Printf("Initialized KServe client for namespace: %s", config.KServeNamespace)
	} else {
		log.Printf("KServe integration disabled (use ENABLE_KSERVE=true to enable)")
	}

	// Create MCP server with metadata
	impl := &mcp.Implementation{
		Name:    config.Name,
		Version: config.Version,
	}

	mcpServer := mcp.NewServer(impl, nil)

	server := &MCPServer{
		config:    config,
		mcpServer: mcpServer,
		k8sClient: k8sClient,
		ceClient:  ceClient,
		kserve:    kserveClient,
		cache:     memoryCache,
		tools:     make(map[string]interface{}),
		resources: make(map[string]interface{}),
	}

	// Register tools
	if err := server.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Register resources
	if err := server.registerResources(); err != nil {
		return nil, fmt.Errorf("failed to register resources: %w", err)
	}

	log.Printf("MCP Server initialized: %s v%s", config.Name, config.Version)
	log.Printf("Transport: %s", config.Transport)

	return server, nil
}

// registerTools initializes and registers all MCP tools
func (s *MCPServer) registerTools() error {
	// Register cluster health tool (with cache)
	clusterHealthTool := tools.NewClusterHealthTool(s.k8sClient, s.cache)
	s.registerTool(clusterHealthTool)

	// Register list-pods tool (no cache - results change frequently)
	listPodsTool := tools.NewListPodsTool(s.k8sClient)
	s.registerTool(listPodsTool)

	// Register Coordination Engine tools if enabled
	if s.ceClient != nil {
		listIncidentsTool := tools.NewListIncidentsTool(s.ceClient)
		s.registerTool(listIncidentsTool)

		triggerRemediationTool := tools.NewTriggerRemediationTool(s.ceClient)
		s.registerTool(triggerRemediationTool)
	} else {
		log.Printf("Skipping Coordination Engine tools (not enabled)")
	}

	// Register KServe tools if enabled
	if s.kserve != nil {
		analyzeAnomaliesTool := tools.NewAnalyzeAnomaliesTool(s.kserve)
		s.registerTool(analyzeAnomaliesTool)

		getModelStatusTool := tools.NewGetModelStatusTool(s.kserve)
		s.registerTool(getModelStatusTool)

		listModelsTool := tools.NewListModelsTool(s.kserve)
		s.registerTool(listModelsTool)
	} else {
		log.Printf("Skipping KServe tools (not enabled)")
	}

	log.Printf("Total tools registered: %d", len(s.tools))
	return nil
}

// Tool interface that our tools implement
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// registerTool registers a tool with both our internal map and the MCP SDK
func (s *MCPServer) registerTool(tool Tool) {
	// Store in our internal map
	s.tools[tool.Name()] = tool

	// Create MCP tool definition
	mcpTool := &mcp.Tool{
		Name:        tool.Name(),
		Description: tool.Description(),
		InputSchema: tool.InputSchema(),
	}

	// Create handler function that wraps our tool's Execute method
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params map[string]interface{}) (*mcp.CallToolResult, any, error) {
		// Execute the tool (params contains the arguments)
		result, err := tool.Execute(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Convert result to JSON string for MCP response
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		// Return as MCP CallToolResult
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(resultJSON),
				},
			},
		}, nil, nil
	}

	// Register with MCP SDK
	mcp.AddTool(s.mcpServer, mcpTool, handler)

	log.Printf("Registered tool: %s - %s", tool.Name(), tool.Description())
}

// registerResources initializes and registers all MCP resources
func (s *MCPServer) registerResources() error {
	// Register cluster://health resource (always available)
	clusterHealthResource := resources.NewClusterHealthResource(s.k8sClient, s.ceClient, s.cache)
	s.resources[clusterHealthResource.URI()] = clusterHealthResource
	log.Printf("Registered resource: %s - %s", clusterHealthResource.URI(), clusterHealthResource.Name())

	// Register cluster://nodes resource (always available)
	nodesResource := resources.NewNodesResource(s.k8sClient, s.cache)
	s.resources[nodesResource.URI()] = nodesResource
	log.Printf("Registered resource: %s - %s", nodesResource.URI(), nodesResource.Name())

	// Register cluster://incidents resource (if Coordination Engine enabled)
	if s.ceClient != nil {
		incidentsResource := resources.NewIncidentsResource(s.ceClient, s.cache)
		s.resources[incidentsResource.URI()] = incidentsResource
		log.Printf("Registered resource: %s - %s", incidentsResource.URI(), incidentsResource.Name())
	} else {
		log.Printf("Skipping cluster://incidents resource (Coordination Engine not enabled)")
	}

	log.Printf("Total resources registered: %d", len(s.resources))
	return nil
}

// GetTools returns all registered tools
func (s *MCPServer) GetTools() map[string]interface{} {
	return s.tools
}

// GetResources returns all registered resources
func (s *MCPServer) GetResources() map[string]interface{} {
	return s.resources
}

// Start begins serving MCP requests using the configured transport
// As of 2025-12-17, only HTTP/SSE transport is supported (stdio DEPRECATED)
func (s *MCPServer) Start(ctx context.Context) error {
	switch s.config.Transport {
	case TransportHTTP:
		return s.startHTTPTransport(ctx)
	case TransportStdio:
		// stdio transport DEPRECATED - return error
		return s.startStdioTransport(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s (only 'http' is supported)", s.config.Transport)
	}
}

// startHTTPTransport starts the server with HTTP/SSE transport
func (s *MCPServer) startHTTPTransport(ctx context.Context) error {
	addr := s.config.GetHTTPAddr()
	log.Printf("Starting HTTP transport on %s", addr)

	// Create the MCP SSE handler (handles SSE transport for OpenShift Lightspeed compatibility)
	// OpenShift Lightspeed expects SSE transport at the root endpoint
	// Reference: https://github.com/openshift/lightspeed-service/
	mcpHandler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	// Create custom handler that routes to either MCP or health endpoints
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route specific endpoints to their handlers
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
			return
		case "/ready":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "READY")
			return
		case "/metrics":
			// TODO: Implement Prometheus metrics in Phase 3
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "# Metrics endpoint (Phase 3)\n")
			return
		case "/cache/stats":
			s.handleCacheStats(w, r)
			return
		case "/mcp/info":
			s.handleMCPInfo(w, r)
			return
		case "/mcp/tools":
			s.handleListTools(w, r)
			return
		case "/mcp/resources":
			s.handleListResources(w, r)
			return
		default:
			// All other paths go to MCP handler (including root "/")
			// This supports GET (SSE) and POST (messages) for MCP protocol
			log.Printf("Routing %s %s to MCP handler", r.Method, r.URL.Path)
			mcpHandler.ServeHTTP(w, r)
		}
	})

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mainHandler,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Printf("MCP Server listening on %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		log.Println("Shutting down HTTP server...")
		return s.httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// startStdioTransport is DEPRECATED as of 2025-12-17
// stdio transport is no longer supported - use HTTP/SSE transport instead
func (s *MCPServer) startStdioTransport(ctx context.Context) error {
	log.Println("ERROR: stdio transport is DEPRECATED as of 2025-12-17")
	log.Println("Please use HTTP/SSE transport instead:")
	log.Println("  MCP_TRANSPORT=http ./mcp-server")
	log.Println("")
	log.Println("Rationale for deprecation:")
	log.Println("  - Primary use case is OpenShift Lightspeed integration via HTTP/SSE")
	log.Println("  - Local development works fine with HTTP transport")
	log.Println("  - stdio was never fully implemented (stub only)")
	log.Println("  - Reduces codebase complexity and maintenance burden")

	return fmt.Errorf("stdio transport is DEPRECATED - use HTTP transport (MCP_TRANSPORT=http)")
}

// handleMCPInfo returns server info
func (s *MCPServer) handleMCPInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"name":"%s","version":"%s","transport":"http/sse","tools_count":%d,"resources_count":%d}`,
		s.config.Name, s.config.Version, len(s.tools), len(s.resources))
}

// handleListTools returns all available tools
func (s *MCPServer) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build tools list response
	type ToolInfo struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		InputSchema map[string]interface{} `json:"input_schema"`
	}

	toolsList := []ToolInfo{}
	for _, tool := range s.tools {
		switch t := tool.(type) {
		case *tools.ClusterHealthTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		case *tools.ListPodsTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		case *tools.ListIncidentsTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		case *tools.TriggerRemediationTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		case *tools.AnalyzeAnomaliesTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		case *tools.GetModelStatusTool:
			toolsList = append(toolsList, ToolInfo{
				Name:        t.Name(),
				Description: t.Description(),
				InputSchema: t.InputSchema(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"tools": toolsList,
		"count": len(toolsList),
	}

	if err := writeJSON(w, response); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// handleClusterHealthTool executes the cluster health tool
func (s *MCPServer) handleClusterHealthTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed - use POST", http.StatusMethodNotAllowed)
		return
	}

	// Get the tool
	tool, ok := s.tools["get-cluster-health"].(*tools.ClusterHealthTool)
	if !ok {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	// Parse request body for arguments
	var args map[string]interface{}
	if r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&args); err != nil {
			// If no body or invalid JSON, use empty args
			args = make(map[string]interface{})
		}
		defer r.Body.Close()
	} else {
		args = make(map[string]interface{})
	}

	// Execute the tool
	ctx := r.Context()
	result, err := tool.Execute(ctx, args)
	if err != nil {
		http.Error(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"result":  result,
	}

	if err := writeJSON(w, response); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// handleListPodsTool executes the list-pods tool
func (s *MCPServer) handleListPodsTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed - use POST", http.StatusMethodNotAllowed)
		return
	}

	// Get the tool
	tool, ok := s.tools["list-pods"].(*tools.ListPodsTool)
	if !ok {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	// Parse request body for arguments
	var args map[string]interface{}
	if r.Body != nil {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&args); err != nil {
			// If no body or invalid JSON, use empty args
			args = make(map[string]interface{})
		}
		defer r.Body.Close()
	} else {
		args = make(map[string]interface{})
	}

	// Execute the tool
	ctx := r.Context()
	result, err := tool.Execute(ctx, args)
	if err != nil {
		http.Error(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"result":  result,
	}

	if err := writeJSON(w, response); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// handleCacheStats returns cache statistics
func (s *MCPServer) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed - use GET", http.StatusMethodNotAllowed)
		return
	}

	stats := s.cache.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := writeJSON(w, stats); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// handleListResources returns all available resources
func (s *MCPServer) handleListResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build resources list response
	type ResourceInfo struct {
		URI         string `json:"uri"`
		Name        string `json:"name"`
		Description string `json:"description"`
		MimeType    string `json:"mime_type"`
	}

	resourcesList := []ResourceInfo{}
	for _, resource := range s.resources {
		switch r := resource.(type) {
		case *resources.ClusterHealthResource:
			resourcesList = append(resourcesList, ResourceInfo{
				URI:         r.URI(),
				Name:        r.Name(),
				Description: r.Description(),
				MimeType:    r.MimeType(),
			})
		case *resources.NodesResource:
			resourcesList = append(resourcesList, ResourceInfo{
				URI:         r.URI(),
				Name:        r.Name(),
				Description: r.Description(),
				MimeType:    r.MimeType(),
			})
		case *resources.IncidentsResource:
			resourcesList = append(resourcesList, ResourceInfo{
				URI:         r.URI(),
				Name:        r.Name(),
				Description: r.Description(),
				MimeType:    r.MimeType(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"resources": resourcesList,
		"count":     len(resourcesList),
	}

	if err := writeJSON(w, response); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

// writeJSON is a helper to write JSON responses
func writeJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Stop gracefully shuts down the server
func (s *MCPServer) Stop() error {
	if s.httpServer != nil {
		log.Println("Stopping HTTP server...")
		return s.httpServer.Shutdown(context.Background())
	}
	return nil
}
