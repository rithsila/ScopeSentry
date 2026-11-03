package commands

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/config"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

const ClientVersion = "2.0.0-headless"

func RunVersionCommand(ctx context.Context, cfg *config.Config, stdout, stderr io.Writer) int {
	type versionInfo struct {
		ClientVersion   string `json:"client_version"`
		ServerURL       string `json:"server_url"`
		ServerReachable bool   `json:"server_reachable"`
	}

	info := versionInfo{
		ClientVersion: ClientVersion,
		ServerURL:     cfg.ServerURL,
	}

	httpClient := &http.Client{Timeout: 3 * time.Second}
	liveURL := strings.TrimRight(cfg.ServerURL, "/") + "/live"
	resp, err := httpClient.Get(liveURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		info.ServerReachable = true
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	_ = output.PrintJSON(stdout, info)
	return output.ExitSuccess
}
