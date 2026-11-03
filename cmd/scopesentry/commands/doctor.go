package commands

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/config"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

type DoctorReport struct {
	Status         string `json:"status"`
	ServerURL      string `json:"server_url"`
	Live           bool   `json:"live"`
	Ready          bool   `json:"ready"`
	MCPConnected   bool   `json:"mcp_connected"`
	ToolsAvailable int    `json:"tools_available"`
	Message        string `json:"message,omitempty"`
}

func RunDoctorCommand(ctx context.Context, cfg *config.Config, stdout, stderr io.Writer) int {
	output.PrintDiagnostic(stderr, "Running ScopeSentry diagnostics for %s...", cfg.ServerURL)

	report := DoctorReport{
		ServerURL: cfg.ServerURL,
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	// 1. Check /live
	liveURL := strings.TrimRight(cfg.ServerURL, "/") + "/live"
	resp, err := httpClient.Get(liveURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		report.Live = true
		output.PrintDiagnostic(stderr, "  [OK] /live reached")
	} else {
		output.PrintDiagnostic(stderr, "  [FAIL] /live unreachable: %v", err)
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// 2. Check /ready
	readyURL := strings.TrimRight(cfg.ServerURL, "/") + "/ready"
	resp, err = httpClient.Get(readyURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		report.Ready = true
		output.PrintDiagnostic(stderr, "  [OK] /ready healthy (MongoDB & Redis connected)")
	} else {
		if resp != nil {
			var body map[string]any
			_ = json.NewDecoder(resp.Body).Decode(&body)
			output.PrintDiagnostic(stderr, "  [WARN] /ready unavailable: %v", body)
		} else {
			output.PrintDiagnostic(stderr, "  [FAIL] /ready unreachable: %v", err)
		}
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// 3. Connect to MCP
	mcpCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mcpCli, err := client.NewMCPClient(mcpCtx, cfg)
	if err == nil {
		defer mcpCli.Close()
		report.MCPConnected = true
		output.PrintDiagnostic(stderr, "  [OK] /mcp connected")

		tools, err := mcpCli.ListTools(mcpCtx)
		if err == nil {
			report.ToolsAvailable = len(tools)
			output.PrintDiagnostic(stderr, "  [OK] %d MCP tools registered", len(tools))
		} else {
			output.PrintDiagnostic(stderr, "  [WARN] failed to list tools: %v", err)
		}
	} else {
		output.PrintDiagnostic(stderr, "  [FAIL] /mcp connection failed: %v", err)
	}

	if report.Live && report.Ready && report.MCPConnected {
		report.Status = "healthy"
	} else if report.Live || report.MCPConnected {
		report.Status = "degraded"
	} else {
		report.Status = "unreachable"
	}

	_ = output.PrintJSON(stdout, report)

	if report.Status == "unreachable" {
		return output.ExitExecutionError
	}
	return output.ExitSuccess
}
