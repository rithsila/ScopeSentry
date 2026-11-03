package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client interface {
	CallTool(ctx context.Context, name string, args any) (*mcp.CallToolResult, error)
	ListTools(ctx context.Context) ([]*mcp.Tool, error)
	Close() error
}

type ToolExecutionError struct {
	ToolName string
	Message  string
}

func (e *ToolExecutionError) Error() string {
	return fmt.Sprintf("tool %s failed: %s", e.ToolName, e.Message)
}

type mcpClient struct {
	session *mcp.ClientSession
}

type authTransport struct {
	apiKey string
	base   http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	if t.apiKey != "" {
		cloned.Header.Set("X-API-Key", t.apiKey)
		cloned.Header.Set("Authorization", "Bearer "+t.apiKey)
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(cloned)
}

func NewMCPClient(ctx context.Context, cfg *config.Config) (Client, error) {
	endpoint := strings.TrimRight(cfg.ServerURL, "/") + "/mcp"

	httpClient := &http.Client{
		Transport: &authTransport{
			apiKey: cfg.APIKey,
		},
		Timeout: cfg.Timeout,
	}

	transport := &mcp.StreamableClientTransport{
		Endpoint:             endpoint,
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true,
	}

	sdkClient := mcp.NewClient(&mcp.Implementation{
		Name:    "scopesentry-cli",
		Version: "1.0.0",
	}, nil)

	session, err := sdkClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP endpoint (%s): %w", endpoint, err)
	}

	return &mcpClient{session: session}, nil
}

func (c *mcpClient) CallTool(ctx context.Context, name string, args any) (*mcp.CallToolResult, error) {
	res, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return nil, err
	}

	if res.IsError {
		text := ExtractText(res)
		return res, &ToolExecutionError{
			ToolName: name,
			Message:  text,
		}
	}

	return res, nil
}

func (c *mcpClient) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	toolList, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	return toolList.Tools, nil
}

func (c *mcpClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}

// ExtractText concatenates text from all TextContent items in CallToolResult.
func ExtractText(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var sb strings.Builder
	for _, content := range res.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return sb.String()
}
