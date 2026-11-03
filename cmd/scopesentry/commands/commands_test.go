package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockClient struct {
	mu           sync.Mutex
	calledTools  []string
	lastArgs     map[string]any
	customResult func(name string, args any) (*mcp.CallToolResult, error)
	tools        []*mcp.Tool
}

func (m *mockClient) CallTool(ctx context.Context, name string, args any) (*mcp.CallToolResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calledTools = append(m.calledTools, name)

	if m.customResult != nil {
		return m.customResult(name, args)
	}

	if mapArgs, ok := args.(map[string]any); ok {
		m.lastArgs = mapArgs
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: `{"status":"ok"}`},
		},
	}, nil
}

func (m *mockClient) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	if m.tools != nil {
		return m.tools, nil
	}
	return []*mcp.Tool{
		{
			Name:        "test_tool",
			Description: "A test tool",
		},
	}, nil
}

func (m *mockClient) Close() error {
	return nil
}

func TestProjectCommands(t *testing.T) {
	ctx := context.Background()

	t.Run("project create with duplicate name replacement response", func(t *testing.T) {
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				if name != "create_project" {
					return nil, fmt.Errorf("unexpected tool %s", name)
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: `{"success":true,"message":"project scope replaced","id":"proj-123","replaced":true}`},
					},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunProjectCommand(ctx, mock, []string{"create", "--name", "Alpha", "--target", "example.com", "--tag", "core"}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}

		var res map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			t.Fatalf("expected valid JSON on stdout, got: %s (err: %v)", stdout.String(), err)
		}
		if res["replaced"] != true || res["id"] != "proj-123" {
			t.Fatalf("unexpected result: %v", res)
		}
	})

	t.Run("project create missing required flags", func(t *testing.T) {
		mock := &mockClient{}
		var stdout, stderr bytes.Buffer
		code := RunProjectCommand(ctx, mock, []string{"create", "--name", "Alpha"}, &stdout, &stderr)
		if code != output.ExitUsageError {
			t.Fatalf("expected exit code 2 (usage error), got %d", code)
		}
		if !strings.Contains(stderr.String(), "--target") {
			t.Fatalf("expected target error on stderr, got: %s", stderr.String())
		}
	})

	t.Run("project list data and by-tag", func(t *testing.T) {
		mock := &mockClient{}
		var stdout, stderr bytes.Buffer
		code := RunProjectCommand(ctx, mock, []string{"list", "--search", "myproj", "--page", "2"}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d", code)
		}
		if len(mock.calledTools) == 0 || mock.calledTools[0] != "list_projects_data" {
			t.Fatalf("expected list_projects_data called, got: %v", mock.calledTools)
		}

		// by-tag
		mock.calledTools = nil
		stdout.Reset()
		code = RunProjectCommand(ctx, mock, []string{"list", "--by-tag"}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d", code)
		}
		if len(mock.calledTools) == 0 || mock.calledTools[0] != "list_projects" {
			t.Fatalf("expected list_projects called, got: %v", mock.calledTools)
		}
	})
}

func TestAssetCommands(t *testing.T) {
	ctx := context.Background()

	t.Run("asset list with search and filter", func(t *testing.T) {
		var receivedArgs map[string]any
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				if m, ok := args.(map[string]any); ok {
					receivedArgs = m
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: `{"list":[]}`}},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunAssetCommand(ctx, mock, []string{
			"list",
			"--type", "subdomain",
			"--search", `subdomain.domain="example.com"`,
			"--filter", `{"project":["proj-1"]}`,
			"--sort", `length:ascending`,
		}, &stdout, &stderr)

		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}

		if receivedArgs["asset_type"] != "subdomain" {
			t.Fatalf("expected asset_type subdomain, got %v", receivedArgs["asset_type"])
		}
		if receivedArgs["search"] != `subdomain.domain="example.com"` {
			t.Fatalf("expected search query, got %v", receivedArgs["search"])
		}
		filterMap, ok := receivedArgs["filter"].(map[string][]string)
		if !ok || len(filterMap["project"]) != 1 || filterMap["project"][0] != "proj-1" {
			t.Fatalf("expected parsed filter map, got: %v", receivedArgs["filter"])
		}
	})

	t.Run("asset count requires type", func(t *testing.T) {
		mock := &mockClient{}
		var stdout, stderr bytes.Buffer
		code := RunAssetCommand(ctx, mock, []string{"count"}, &stdout, &stderr)
		if code != output.ExitUsageError {
			t.Fatalf("expected exit code 2, got %d", code)
		}
	})
}

func TestTemplateCommands(t *testing.T) {
	ctx := context.Background()

	t.Run("template create with module json", func(t *testing.T) {
		var receivedArgs map[string]any
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				if m, ok := args.(map[string]any); ok {
					receivedArgs = m
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: `{"success":true,"id":"tmpl-1"}`}},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunTemplateCommand(ctx, mock, []string{
			"create",
			"--name", "MyTemplate",
			"--modules", `{"SubdomainScan":["hash123"]}`,
			"--vullist", "cve-2023-1,cve-2024-2",
		}, &stdout, &stderr)

		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}
		if receivedArgs["name"] != "MyTemplate" {
			t.Fatalf("expected name MyTemplate, got %v", receivedArgs["name"])
		}
		modulesMap, ok := receivedArgs["modules"].(map[string]any)
		if !ok || modulesMap["SubdomainScan"] == nil {
			t.Fatalf("expected modules map, got: %v", receivedArgs["modules"])
		}
	})
}

func TestTaskWaitPollingAndTimeout(t *testing.T) {
	ctx := context.Background()

	t.Run("task wait completes when status is 3", func(t *testing.T) {
		pollCount := 0
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				pollCount++
				status := 1
				prog := 50
				if pollCount >= 2 {
					status = 3
					prog = 100
				}
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf(`{"id":"task-99","status":%d,"progress":%d}`, status, prog)},
					},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunTaskCommand(ctx, mock, []string{
			"wait",
			"task-99",
			"--timeout", "5s",
			"--interval", "20ms",
		}, &stdout, &stderr)

		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}
		if pollCount < 2 {
			t.Fatalf("expected at least 2 polls, got %d", pollCount)
		}
		if !strings.Contains(stderr.String(), "status: 1") {
			t.Fatalf("expected intermediate progress on stderr, got: %s", stderr.String())
		}
		if !strings.Contains(stdout.String(), `"status": 3`) && !strings.Contains(stdout.String(), `"status":3`) {
			t.Fatalf("expected final task JSON on stdout, got: %s", stdout.String())
		}
	})

	t.Run("task wait exits code 3 on timeout without cancelling", func(t *testing.T) {
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: `{"id":"task-never-finishes","status":1,"progress":10}`},
					},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunTaskCommand(ctx, mock, []string{
			"wait",
			"task-never-finishes",
			"--timeout", "60ms",
			"--interval", "15ms",
		}, &stdout, &stderr)

		if code != output.ExitTimeout {
			t.Fatalf("expected exit code 3 (ExitTimeout), got %d. stderr: %s", code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "timeout:") {
			t.Fatalf("expected timeout error on stderr, got: %s", stderr.String())
		}
	})
}

func TestToolsCommands(t *testing.T) {
	ctx := context.Background()

	t.Run("tools list and describe", func(t *testing.T) {
		mock := &mockClient{
			tools: []*mcp.Tool{
				{
					Name:        "demo_tool",
					Description: "demo tool for unit test",
					InputSchema: map[string]any{"type": "object"},
				},
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunToolsCommand(ctx, mock, []string{"list"}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d", code)
		}

		var list []map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &list); err != nil {
			t.Fatalf("expected json array on stdout: %v, got %s", err, stdout.String())
		}
		if len(list) != 1 || list[0]["name"] != "demo_tool" {
			t.Fatalf("unexpected tools list: %v", list)
		}

		// describe
		stdout.Reset()
		code = RunToolsCommand(ctx, mock, []string{"describe", "demo_tool"}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d", code)
		}
		var desc map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &desc); err != nil {
			t.Fatalf("expected json object: %v", err)
		}
		if desc["name"] != "demo_tool" {
			t.Fatalf("unexpected description: %v", desc)
		}
	})

	t.Run("tools call with input json", func(t *testing.T) {
		var receivedTool string
		var receivedArgs any
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				receivedTool = name
				receivedArgs = args
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: `{"called":true}`}},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunToolsCommand(ctx, mock, []string{"call", "custom_mcp_tool", "--input", `{"foo":"bar"}`}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}
		if receivedTool != "custom_mcp_tool" {
			t.Fatalf("expected tool custom_mcp_tool, got %s", receivedTool)
		}
		argsMap, ok := receivedArgs.(map[string]any)
		if !ok || argsMap["foo"] != "bar" {
			t.Fatalf("expected args foo=bar, got: %v", receivedArgs)
		}
	})

	t.Run("tools call with --args alias", func(t *testing.T) {
		var receivedTool string
		var receivedArgs any
		mock := &mockClient{
			customResult: func(name string, args any) (*mcp.CallToolResult, error) {
				receivedTool = name
				receivedArgs = args
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: `{"called":true}`}},
				}, nil
			},
		}

		var stdout, stderr bytes.Buffer
		code := RunToolsCommand(ctx, mock, []string{"call", "custom_mcp_tool", "--args", `{"hello":"world"}`}, &stdout, &stderr)
		if code != output.ExitSuccess {
			t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
		}
		if receivedTool != "custom_mcp_tool" {
			t.Fatalf("expected tool custom_mcp_tool, got %s", receivedTool)
		}
		argsMap, ok := receivedArgs.(map[string]any)
		if !ok || argsMap["hello"] != "world" {
			t.Fatalf("expected args hello=world, got: %v", receivedArgs)
		}
	})
}
