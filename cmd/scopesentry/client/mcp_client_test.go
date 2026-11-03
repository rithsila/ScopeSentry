package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockToolInput struct {
	Msg string `json:"msg"`
}

func TestMCPClientAuthAndToolCall(t *testing.T) {
	var receivedAuthHeader string
	var receivedAPIKeyHeader string

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "echo tool",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input mockToolInput) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: `{"echo":"` + input.Msg + `"}`},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fail_tool",
		Description: "failing tool",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "simulated failure"},
			},
		}, nil, nil
	})

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(_ *http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		receivedAPIKeyHeader = r.Header.Get("X-API-Key")
		mcpHandler.ServeHTTP(w, r)
	}))
	mux.Handle("/mcp/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		receivedAPIKeyHeader = r.Header.Get("X-API-Key")
		mcpHandler.ServeHTTP(w, r)
	}))

	ts := httptest.NewServer(mux)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := &config.Config{
		ServerURL: ts.URL,
		APIKey:    "test-api-key-xyz",
		Timeout:   3 * time.Second,
	}

	client, err := NewMCPClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// 1. Verify headers received
	if receivedAPIKeyHeader != "test-api-key-xyz" {
		t.Fatalf("expected X-API-Key 'test-api-key-xyz', got %q", receivedAPIKeyHeader)
	}
	if receivedAuthHeader != "Bearer test-api-key-xyz" {
		t.Fatalf("expected Authorization 'Bearer test-api-key-xyz', got %q", receivedAuthHeader)
	}

	// 2. Call echo tool
	res, err := client.CallTool(ctx, "echo", mockToolInput{Msg: "hello"})
	if err != nil {
		t.Fatalf("unexpected error on CallTool: %v", err)
	}
	extracted := ExtractText(res)
	if extracted != `{"echo":"hello"}` {
		t.Fatalf("expected echo result, got: %s", extracted)
	}

	// 3. Call tool that returns isError: true
	_, err = client.CallTool(ctx, "fail_tool", map[string]any{})
	if err == nil {
		t.Fatal("expected error when tool returns IsError: true, got nil")
	}
	toolErr, ok := err.(*ToolExecutionError)
	if !ok {
		t.Fatalf("expected *ToolExecutionError, got %T: %v", err, err)
	}
	if toolErr.ToolName != "fail_tool" || toolErr.Message != "simulated failure" {
		t.Fatalf("unexpected tool error payload: %+v", toolErr)
	}

	// 4. List tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("failed to list tools: %v", err)
	}
	if len(tools) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(tools))
	}
}
