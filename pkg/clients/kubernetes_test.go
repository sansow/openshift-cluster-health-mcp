package clients

import (
	"context"
	"testing"
	"time"
)

func TestNewK8sClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *K8sClientConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false, // Should succeed with kubeconfig
		},
		{
			name: "custom config",
			config: &K8sClientConfig{
				QPS:     100,
				Burst:   200,
				Timeout: 60 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewK8sClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewK8sClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewK8sClient() returned nil client")
			}
			if client != nil {
				defer client.Close()
			}
		})
	}
}

func TestK8sClient_HealthCheck(t *testing.T) {
	client, err := NewK8sClient(nil)
	if err != nil {
		t.Skipf("Skipping: unable to create Kubernetes client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() failed: %v", err)
	}
}

func TestK8sClient_GetServerVersion(t *testing.T) {
	client, err := NewK8sClient(nil)
	if err != nil {
		t.Skipf("Skipping: unable to create Kubernetes client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.GetServerVersion(ctx)
	if err != nil {
		t.Errorf("GetServerVersion() failed: %v", err)
	}
	if version == "" {
		t.Error("GetServerVersion() returned empty version")
	}
	t.Logf("Server version: %s", version)
}

func TestK8sClient_ListNodes(t *testing.T) {
	client, err := NewK8sClient(nil)
	if err != nil {
		t.Skipf("Skipping: unable to create Kubernetes client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	nodes, err := client.ListNodes(ctx)
	if err != nil {
		t.Errorf("ListNodes() failed: %v", err)
	}
	if len(nodes.Items) == 0 {
		t.Error("ListNodes() returned no nodes")
	}
	t.Logf("Found %d nodes", len(nodes.Items))
}

func TestK8sClient_GetClusterHealth(t *testing.T) {
	client, err := NewK8sClient(nil)
	if err != nil {
		t.Skipf("Skipping: unable to create Kubernetes client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	health, err := client.GetClusterHealth(ctx)
	if err != nil {
		t.Errorf("GetClusterHealth() failed: %v", err)
	}
	if health == nil {
		t.Error("GetClusterHealth() returned nil")
	}

	t.Logf("Cluster Health:")
	t.Logf("  Status: %s", health.Status)
	t.Logf("  Nodes: %d total, %d ready, %d not ready",
		health.Nodes.Total, health.Nodes.Ready, health.Nodes.NotReady)
	t.Logf("  Pods: %d total, %d running, %d pending, %d failed",
		health.Pods.Total, health.Pods.Running, health.Pods.Pending, health.Pods.Failed)
}
