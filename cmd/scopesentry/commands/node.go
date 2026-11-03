package commands

import (
	"context"
	"flag"
	"io"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunNodeCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing node subcommand. available: list")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "list":
		fs := flag.NewFlagSet("node list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		onlineOnly := fs.Bool("online-only", false, "Only return online nodes")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		res, err := c.CallTool(ctx, "list_nodes", map[string]any{"online_only": *onlineOnly})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown node subcommand %q. available: list", subCmd)
		return output.ExitUsageError
	}
}
