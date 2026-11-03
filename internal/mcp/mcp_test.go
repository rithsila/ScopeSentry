package mcp

import (
	"context"
	"testing"
	"time"

	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewServerRegistersAllTools(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tServer, tClient := mcpSDK.NewInMemoryTransports()
	srv := newServer()

	go func() {
		_, _ = srv.Connect(ctx, tServer, nil)
	}()

	client := mcpSDK.NewClient(&mcpSDK.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := client.Connect(ctx, tClient, nil)
	if err != nil {
		t.Fatalf("failed to connect to in-memory server: %v", err)
	}
	defer session.Close()

	toolList, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}

	expectedTools := []string{
		"list_projects",
		"list_projects_data",
		"get_project",
		"create_project",
		"list_tasks",
		"get_task",
		"list_scan_templates",
		"get_scan_template",
		"list_plugin_modules",
		"list_plugins",
		"create_scan_template",
		"create_scan_task",
		"list_assets",
		"count_assets",
		"get_asset_detail",
		"add_asset_tag",
		"list_nodes",
	}

	foundMap := make(map[string]bool)
	for _, tool := range toolList.Tools {
		foundMap[tool.Name] = true
	}

	if len(toolList.Tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(toolList.Tools))
	}

	var missing []string
	for _, exp := range expectedTools {
		if !foundMap[exp] {
			missing = append(missing, exp)
		}
	}

	if len(missing) > 0 {
		t.Errorf("missing tools: %v", missing)
	}
}

func TestNormalizeAssetIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"asset", "asset", false},
		{"web", "asset", false},
		{"RootDomain", "RootDomain", false},
		{"rootdomain", "RootDomain", false},
		{"root_domain", "RootDomain", false},
		{"subdomain", "subdomain", false},
		{"app", "app", false},
		{"mp", "mp", false},
		{"miniprogram", "mp", false},
		{"url", "UrlScan", false},
		{"UrlScan", "UrlScan", false},
		{"SensitiveResult", "SensitiveResult", false},
		{"DirScanResult", "DirScanResult", false},
		{"crawler", "crawler", false},
		{"vulnerability", "vulnerability", false},
		{"PageMonitoring", "PageMonitoring", false},
		{"IPAsset", "IPAsset", false},
		{"SubdomainTakerResult", "SubdomainTakerResult", false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		got, err := normalizeAssetIndex(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("normalizeAssetIndex(%q) err = %v, wantErr = %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("normalizeAssetIndex(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
