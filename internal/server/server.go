package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/internal/tools"
	"github.com/openshift-aiops/openshift-cluster-health-mcp/pkg/clients"
)

// MCPServer wraps the official MCP SDK server
type MCPServer struct {
	config     *Config
	mcpServer  *mcp.Server
	httpServer *http.Server
	k8sClient  *clients.K8sClient
	tools      map[string]interface{} // Registry of available tools
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
		tools:     make(map[string]interface{}),
	}

	// Register tools
	if err := server.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	log.Printf("MCP Server initialized: %s v%s", config.Name, config.Version)
	log.Printf("Transport: %s", config.Transport)

	return server, nil
}

// registerTools initializes and registers all MCP tools
func (s *MCPServer) registerTools() error {
	// Register cluster health tool
	clusterHealthTool := tools.NewClusterHealthTool(s.k8sClient)
	s.tools[clusterHealthTool.Name()] = clusterHealthTool
	log.Printf("Registered tool: %s - %s", clusterHealthTool.Name(), clusterHealthTool.Description())

	// Register list-pods tool
	listPodsTool := tools.NewListPodsTool(s.k8sClient)
	s.tools[listPodsTool.Name()] = listPodsTool
	log.Printf("Registered tool: %s - %s", listPodsTool.Name(), listPodsTool.Description())

	// Future tools will be registered here:
	// - get-pod-logs
	// - get-events
	// etc.

	log.Printf("Total tools registered: %d", len(s.tools))
	return nil
}

// GetTools returns all registered tools
func (s *MCPServer) GetTools() map[string]interface{} {
	return s.tools
}

// Start begins serving MCP requests using the configured transport
func (s *MCPServer) Start(ctx context.Context) error {
	switch s.config.Transport {
	case TransportHTTP:
		return s.startHTTPTransport(ctx)
	case TransportStdio:
		return s.startStdioTransport(ctx)
	default:
		return fmt.Errorf("unsupported transport: %s", s.config.Transport)
	}
}

// startHTTPTransport starts the server with HTTP/SSE transport
func (s *MCPServer) startHTTPTransport(ctx context.Context) error {
	addr := s.config.GetHTTPAddr()
	log.Printf("Starting HTTP transport on %s", addr)

	// Create HTTP server with MCP handler
	mux := http.NewServeMux()

	// MCP endpoint (using SSE transport)
	mux.HandleFunc("/mcp", s.handleMCPInfo)

	// Tools endpoints
	mux.HandleFunc("/mcp/tools", s.handleListTools)
	mux.HandleFunc("/mcp/tools/get-cluster-health/call", s.handleClusterHealthTool)
	mux.HandleFunc("/mcp/tools/list-pods/call", s.handleListPodsTool)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Readiness check
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "READY")
	})

	// Metrics endpoint (Prometheus)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement Prometheus metrics in Phase 3
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "# Metrics endpoint (Phase 3)\n")
	})

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
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

// startStdioTransport starts the server with stdio transport (for local dev)
func (s *MCPServer) startStdioTransport(ctx context.Context) error {
	log.Println("Starting stdio transport (for local development)")

	// The official SDK uses stdio by default
	// Running the server with nil transport uses stdio
	// TODO: Verify exact stdio API in Phase 1.3
	log.Println("TODO: Implement stdio transport using official SDK API")
	log.Println("For now, use HTTP transport for testing")

	// Block until context is cancelled
	<-ctx.Done()
	return nil
}

// handleMCPInfo returns server info
func (s *MCPServer) handleMCPInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"name":"%s","version":"%s","transport":"http/sse","tools_count":%d}`,
		s.config.Name, s.config.Version, len(s.tools))
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
