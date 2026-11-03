package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/constants"
)

func TestAdminCommands(t *testing.T) {
	t.Run("pluginkey subcommand outputs json with plugin key", func(t *testing.T) {
		config.GlobalConfig.System.PluginKey = "test-plugin-key-12345"
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"pluginkey"}, &stdout, &stderr)
		if !handled {
			t.Fatal("expected pluginkey command to be handled")
		}

		var res map[string]string
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			t.Fatalf("expected JSON output on stdout, got: %s (err: %v)", stdout.String(), err)
		}
		if res["plugin_key"] != "test-plugin-key-12345" {
			t.Fatalf("expected plugin_key test-plugin-key-12345, got %v", res["plugin_key"])
		}
	})

	t.Run("version subcommand outputs json with version", func(t *testing.T) {
		constants.Version = "2.0.0-headless"
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"version"}, &stdout, &stderr)
		if !handled {
			t.Fatal("expected version command to be handled")
		}

		var res map[string]string
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			t.Fatalf("expected JSON output on stdout, got: %s (err: %v)", stdout.String(), err)
		}
		if res["version"] != "2.0.0-headless" {
			t.Fatalf("expected version 2.0.0-headless, got %v", res["version"])
		}
	})

	t.Run("apikey missing subcommand returns usage error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"apikey"}, &stdout, &stderr)
		if !handled {
			t.Fatal("expected apikey command to be handled")
		}
		if !strings.Contains(stderr.String(), "usage:") && !strings.Contains(stdout.String(), "usage") {
			t.Fatalf("expected usage message, got stderr: %s, stdout: %s", stderr.String(), stdout.String())
		}
	})

	t.Run("apikey create missing name flag returns error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"apikey", "create"}, &stdout, &stderr)
		if !handled {
			t.Fatal("expected apikey create command to be handled")
		}
		if !strings.Contains(stderr.String(), "name is required") {
			t.Fatalf("expected 'name is required' error on stderr, got: %s", stderr.String())
		}
	})

	t.Run("apikey delete missing id flag returns error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"apikey", "delete"}, &stdout, &stderr)
		if !handled {
			t.Fatal("expected apikey delete command to be handled")
		}
		if !strings.Contains(stderr.String(), "id is required") {
			t.Fatalf("expected 'id is required' error on stderr, got: %s", stderr.String())
		}
	})

	t.Run("unrecognized command returns false", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		handled := runAdminCommand([]string{"some-unknown-cmd"}, &stdout, &stderr)
		if handled {
			t.Fatal("expected unknown command to not be handled")
		}
	})
}
