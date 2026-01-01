package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openshift-aiops/openshift-cluster-health-mcp/pkg/clients"
)

// AnalyzeAnomaliesTool provides MCP tool for ML-powered anomaly detection
type AnalyzeAnomaliesTool struct {
	kserveClient *clients.KServeClient
}

// NewAnalyzeAnomaliesTool creates a new analyze-anomalies tool
func NewAnalyzeAnomaliesTool(kserveClient *clients.KServeClient) *AnalyzeAnomaliesTool {
	return &AnalyzeAnomaliesTool{
		kserveClient: kserveClient,
	}
}

// Name returns the tool name
func (t *AnalyzeAnomaliesTool) Name() string {
	return "analyze-anomalies"
}

// Description returns the tool description
func (t *AnalyzeAnomaliesTool) Description() string {
	return "Analyze Prometheus metrics for anomalies using ML-powered KServe models. Returns anomaly scores with confidence levels and natural language explanations."
}

// InputSchema returns the JSON schema for tool inputs
func (t *AnalyzeAnomaliesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"metric": map[string]interface{}{
				"type":        "string",
				"description": "The metric name to analyze (e.g., 'cpu_usage', 'memory_usage', 'pod_restarts')",
			},
			"namespace": map[string]interface{}{
				"type":        "string",
				"description": "Kubernetes namespace to scope the analysis (optional)",
			},
			"time_range": map[string]interface{}{
				"type":        "string",
				"description": "Time range for analysis",
				"enum":        []string{"1h", "6h", "24h", "7d"},
				"default":     "1h",
			},
			"threshold": map[string]interface{}{
				"type":        "number",
				"description": "Anomaly score threshold (0.0-1.0). Values above this are reported.",
				"default":     0.7,
				"minimum":     0.0,
				"maximum":     1.0,
			},
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "KServe model name to use for prediction",
				"default":     "predictive-analytics",
			},
		},
		"required": []string{"metric"},
	}
}

// AnalyzeAnomaliesInput represents the input parameters
type AnalyzeAnomaliesInput struct {
	Metric    string  `json:"metric"`
	Namespace string  `json:"namespace"`
	TimeRange string  `json:"time_range"`
	Threshold float64 `json:"threshold"`
	ModelName string  `json:"model_name"`
}

// AnomalyResult represents a detected anomaly
type AnomalyResult struct {
	Timestamp    string  `json:"timestamp"`
	MetricName   string  `json:"metric_name"`
	Value        float64 `json:"value"`
	AnomalyScore float64 `json:"anomaly_score"`
	Confidence   float64 `json:"confidence"`
	Severity     string  `json:"severity"`
	Explanation  string  `json:"explanation"`
}

// AnalyzeAnomaliesOutput represents the tool output
type AnalyzeAnomaliesOutput struct {
	Status         string          `json:"status"`
	Metric         string          `json:"metric"`
	TimeRange      string          `json:"time_range"`
	Namespace      string          `json:"namespace,omitempty"`
	ModelUsed      string          `json:"model_used"`
	Anomalies      []AnomalyResult `json:"anomalies"`
	AnomalyCount   int             `json:"anomaly_count"`
	MaxScore       float64         `json:"max_score"`
	AverageScore   float64         `json:"average_score"`
	Message        string          `json:"message"`
	Recommendation string          `json:"recommendation,omitempty"`
}

// Execute runs the analyze-anomalies tool
func (t *AnalyzeAnomaliesTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse input arguments with defaults
	input := AnalyzeAnomaliesInput{
		TimeRange: "1h",
		Threshold: 0.7,
		ModelName: "predictive-analytics",
	}

	if argsJSON, err := json.Marshal(args); err == nil {
		json.Unmarshal(argsJSON, &input)
	}

	// Validate required fields
	if input.Metric == "" {
		return nil, fmt.Errorf("metric is required")
	}

	// Build prediction request
	instances := []map[string]interface{}{
		{
			"metric":     input.Metric,
			"namespace":  input.Namespace,
			"time_range": input.TimeRange,
		},
	}

	// Call KServe model for prediction
	prediction, err := t.kserveClient.Predict(ctx, input.ModelName, instances)
	if err != nil {
		return nil, fmt.Errorf("failed to get anomaly predictions: %w", err)
	}

	// Parse predictions and build anomaly results
	anomalies := []AnomalyResult{}
	var totalScore float64
	var maxScore float64

	if predList, ok := prediction.Predictions.([]interface{}); ok {
		for _, pred := range predList {
			if predMap, ok := pred.(map[string]interface{}); ok {
				score := getFloat64(predMap, "anomaly_score")
				confidence := getFloat64(predMap, "confidence")

				// Only include anomalies above threshold
				if score >= input.Threshold {
					anomaly := AnomalyResult{
						Timestamp:    getString(predMap, "timestamp"),
						MetricName:   input.Metric,
						Value:        getFloat64(predMap, "value"),
						AnomalyScore: score,
						Confidence:   confidence,
						Severity:     determineSeverity(score),
						Explanation:  generateExplanation(input.Metric, score, confidence),
					}
					anomalies = append(anomalies, anomaly)
					totalScore += score
					if score > maxScore {
						maxScore = score
					}
				}
			}
		}
	}

	// Calculate statistics
	avgScore := 0.0
	if len(anomalies) > 0 {
		avgScore = totalScore / float64(len(anomalies))
	}

	// Build output
	output := AnalyzeAnomaliesOutput{
		Status:       "success",
		Metric:       input.Metric,
		TimeRange:    input.TimeRange,
		Namespace:    input.Namespace,
		ModelUsed:    input.ModelName,
		Anomalies:    anomalies,
		AnomalyCount: len(anomalies),
		MaxScore:     maxScore,
		AverageScore: avgScore,
	}

	// Generate message and recommendation
	if len(anomalies) == 0 {
		output.Message = fmt.Sprintf("No anomalies detected in %s over the last %s (threshold: %.2f)", input.Metric, input.TimeRange, input.Threshold)
		output.Recommendation = "Cluster metrics appear normal. Continue monitoring."
	} else {
		output.Message = fmt.Sprintf("Detected %d anomalies in %s over the last %s (max score: %.2f)", len(anomalies), input.Metric, input.TimeRange, maxScore)
		output.Recommendation = generateRecommendation(input.Metric, maxScore, len(anomalies))
	}

	return output, nil
}

// Helper functions

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0.0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func determineSeverity(score float64) string {
	if score >= 0.9 {
		return "critical"
	} else if score >= 0.8 {
		return "high"
	} else if score >= 0.7 {
		return "medium"
	}
	return "low"
}

func generateExplanation(metric string, score, confidence float64) string {
	severity := determineSeverity(score)
	return fmt.Sprintf("Metric '%s' shows %s anomaly (score: %.2f, confidence: %.2f). "+
		"This indicates unusual behavior compared to historical patterns.",
		metric, severity, score, confidence)
}

func generateRecommendation(metric string, maxScore float64, count int) string {
	if maxScore >= 0.9 {
		return fmt.Sprintf("CRITICAL: Immediate investigation recommended. %d high-severity anomalies detected in %s. "+
			"Consider scaling resources or triggering remediation workflows.", count, metric)
	} else if maxScore >= 0.8 {
		return fmt.Sprintf("WARNING: Monitor closely. %d anomalies detected in %s. "+
			"Review recent deployments or configuration changes.", count, metric)
	}
	return fmt.Sprintf("INFO: %d minor anomalies detected in %s. "+
		"Monitor trends and investigate if anomalies persist.", count, metric)
}
