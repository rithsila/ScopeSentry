package commands

import (
	"context"
	"flag"
	"io"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunPluginCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing plugin subcommand. available: modules, list")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "modules":
		res, err := c.CallTool(ctx, "list_plugin_modules", map[string]any{})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "list":
		fs := flag.NewFlagSet("plugin list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		module := fs.String("module", "", "Filter plugins by module name (see 'scopesentry plugin modules')")
		search := fs.String("search", "", "Search plugins by name")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{}
		if *module != "" {
			mcpArgs["module"] = *module
		}
		if *search != "" {
			mcpArgs["search"] = *search
		}

		res, err := c.CallTool(ctx, "list_plugins", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown plugin subcommand %q. available: modules, list", subCmd)
		return output.ExitUsageError
	}
}
