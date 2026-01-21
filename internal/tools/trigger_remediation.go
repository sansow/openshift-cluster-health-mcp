package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openshift-aiops/openshift-cluster-health-mcp/pkg/clients"
)

// TriggerRemediationTool provides MCP tool for triggering remediation actions
type TriggerRemediationTool struct {
	ceClient *clients.CoordinationEngineClient
}

// NewTriggerRemediationTool creates a new trigger-remediation tool
func NewTriggerRemediationTool(ceClient *clients.CoordinationEngineClient) *TriggerRemediationTool {
	return &TriggerRemediationTool{
		ceClient: ceClient,
	}
}

// Name returns the tool name
func (t *TriggerRemediationTool) Name() string {
	return "trigger-remediation"
}

// Description returns the tool description
func (t *TriggerRemediationTool) Description() string {
	return "Trigger automated remediation actions for incidents through the Coordination Engine. Requires incident_id, namespace, resource details, and issue information."
}

// InputSchema returns the JSON schema for tool inputs
func (t *TriggerRemediationTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"incident_id": map[string]interface{}{
				"type":        "string",
				"description": "The ID of the incident to remediate",
			},
			"namespace": map[string]interface{}{
				"type":        "string",
				"description": "Kubernetes namespace",
			},
			"resource_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the resource (deployment, pod, etc.)",
			},
			"resource_kind": map[string]interface{}{
				"type":        "string",
				"description": "Kind of resource (Deployment, Pod, StatefulSet)",
				"enum":        []string{"Deployment", "Pod", "StatefulSet", "DaemonSet"},
			},
			"issue_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of issue",
				"enum":        []string{"pod_crash", "oom_kill", "high_cpu", "high_memory", "network_issue"},
			},
			"severity": map[string]interface{}{
				"type":        "string",
				"description": "Issue severity",
				"enum":        []string{"low", "medium", "high", "critical"},
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the issue",
			},
			"dry_run": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, validate without executing",
				"default":     false,
			},
		},
		"required": []string{"incident_id", "namespace", "resource_name", "resource_kind", "issue_type", "severity"},
	}
}

// TriggerRemediationInput represents the input parameters
type TriggerRemediationInput struct {
	IncidentID   string `json:"incident_id"`
	Namespace    string `json:"namespace"`
	ResourceName string `json:"resource_name"`
	ResourceKind string `json:"resource_kind"`
	IssueType    string `json:"issue_type"`
	Severity     string `json:"severity"`
	Description  string `json:"description"`
	DryRun       bool   `json:"dry_run"`
}

// TriggerRemediationOutput represents the tool output
type TriggerRemediationOutput struct {
	Status            string `json:"status"`
	WorkflowID        string `json:"workflow_id"`
	IncidentID        string `json:"incident_id"`
	DeploymentMethod  string `json:"deployment_method"`
	EstimatedDuration string `json:"estimated_duration"`
	Message           string `json:"message"`
	DryRun            bool   `json:"dry_run,omitempty"`
}

// Execute runs the trigger-remediation tool
func (t *TriggerRemediationTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse input arguments
	input := TriggerRemediationInput{
		DryRun: false,
	}

	if argsJSON, err := json.Marshal(args); err == nil {
		_ = json.Unmarshal(argsJSON, &input)
	}

	// Validate required fields
	if input.IncidentID == "" {
		return nil, fmt.Errorf("incident_id is required")
	}
	if input.Namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	if input.ResourceName == "" {
		return nil, fmt.Errorf("resource_name is required")
	}
	if input.ResourceKind == "" {
		return nil, fmt.Errorf("resource_kind is required")
	}
	if input.IssueType == "" {
		return nil, fmt.Errorf("issue_type is required")
	}
	if input.Severity == "" {
		return nil, fmt.Errorf("severity is required")
	}

	// Build remediation request
	req := &clients.TriggerRemediationRequest{
		IncidentID: input.IncidentID,
		Namespace:  input.Namespace,
		DryRun:     input.DryRun,
	}
	req.Resource.Kind = input.ResourceKind
	req.Resource.Name = input.ResourceName
	req.Issue.Type = input.IssueType
	req.Issue.Severity = input.Severity
	req.Issue.Description = input.Description

	// Call Coordination Engine API
	resp, err := t.ceClient.TriggerRemediation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger remediation: %w", err)
	}

	// Build output
	output := TriggerRemediationOutput{
		Status:            resp.Status,
		WorkflowID:        resp.WorkflowID,
		IncidentID:        input.IncidentID,
		DeploymentMethod:  resp.DeploymentMethod,
		EstimatedDuration: resp.EstimatedDuration,
		DryRun:            input.DryRun,
	}

	if input.DryRun {
		output.Message = fmt.Sprintf("DRY RUN: Remediation for %s/%s validated successfully", input.ResourceKind, input.ResourceName)
	} else {
		output.Message = fmt.Sprintf("Remediation triggered successfully (workflow: %s)", resp.WorkflowID)
	}

	return output, nil
}
