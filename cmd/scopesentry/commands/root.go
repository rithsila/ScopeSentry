package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/config"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func PrintUsage(w io.Writer) {
	usage := `ScopeSentry CLI - Headless Security Scanning & Asset Mapping Client

Usage:
  scopesentry [global flags] <command> [subcommand] [flags]

Commands:
  project     Manage scanning projects (list, get, create)
  task        Manage scan tasks (list, get, create, wait, stop, resume)
  template    Manage scan templates (list, get, create)
  plugin      Inspect scan plugins and pipeline modules (modules, list)
  asset       Search, filter, and inspect discovered assets (list, count, get, add-tag)
  node        List distributed scan workers (list)
  tools       Inspect and invoke MCP tools directly (list, describe, call)
  doctor      Diagnose connectivity to ScopeSentry server, health endpoints, and MCP
  version     Display CLI version and server reachability

Global Flags:
  --url, -u       ScopeSentry server URL (default: http://localhost:8082 or $SCOPESENTRY_URL)
  --api-key, -k   ScopeSentry API Key ($SCOPESENTRY_API_KEY)
  --help, -h      Show help message
  --version, -v   Show version

Output:
  stdout is strictly machine-readable JSON for scripting and AI agents.
  stderr contains progress, diagnostic banners, and error logs.
`
	fmt.Fprint(w, usage)
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	var flagURL string
	var flagAPIKey string
	var cmdArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			PrintUsage(stderr)
			return output.ExitSuccess
		case arg == "--version" || arg == "-v":
			cfg := config.LoadConfig(flagURL, flagAPIKey)
			return RunVersionCommand(ctx, cfg, stdout, stderr)
		case arg == "--url" || arg == "-u":
			if i+1 < len(args) {
				flagURL = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--url="):
			flagURL = strings.TrimPrefix(arg, "--url=")
		case arg == "--api-key" || arg == "-k":
			if i+1 < len(args) {
				flagAPIKey = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--api-key="):
			flagAPIKey = strings.TrimPrefix(arg, "--api-key=")
		case !strings.HasPrefix(arg, "-"):
			cmdArgs = args[i:]
			i = len(args)
		default:
			output.PrintError(stderr, "unknown global flag %q. Run 'scopesentry --help'", arg)
			return output.ExitUsageError
		}
	}

	if len(cmdArgs) == 0 {
		PrintUsage(stderr)
		return output.ExitUsageError
	}

	cfg := config.LoadConfig(flagURL, flagAPIKey)
	cmd := cmdArgs[0]
	subArgs := cmdArgs[1:]

	switch cmd {
	case "version":
		return RunVersionCommand(ctx, cfg, stdout, stderr)

	case "doctor":
		return RunDoctorCommand(ctx, cfg, stdout, stderr)

	case "project", "task", "template", "plugin", "asset", "node", "tools":
		mcpCli, err := client.NewMCPClient(ctx, cfg)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		defer mcpCli.Close()

		switch cmd {
		case "project":
			return RunProjectCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "task":
			return RunTaskCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "template":
			return RunTemplateCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "plugin":
			return RunPluginCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "asset":
			return RunAssetCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "node":
			return RunNodeCommand(ctx, mcpCli, subArgs, stdout, stderr)
		case "tools":
			return RunToolsCommand(ctx, mcpCli, subArgs, stdout, stderr)
		}

	default:
		output.PrintError(stderr, "unknown command %q. Run 'scopesentry --help' for available commands", cmd)
		return output.ExitUsageError
	}

	return output.ExitSuccess
}
