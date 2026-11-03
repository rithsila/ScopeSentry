package commands

import (
	"context"
	"flag"
	"io"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunProjectCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing project subcommand. available: list, get, create")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "list":
		fs := flag.NewFlagSet("project list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		byTag := fs.Bool("by-tag", false, "Group projects by tag")
		search := fs.String("search", "", "Search projects by name")
		page := fs.Int("page", 1, "Page number starting from 1")
		size := fs.Int("size", 20, "Items per page")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		if *byTag {
			res, err := c.CallTool(ctx, "list_projects", map[string]any{})
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitExecutionError
			}
			_ = output.PrintRawJSON(stdout, client.ExtractText(res))
			return output.ExitSuccess
		}

		mcpArgs := map[string]any{
			"search":    *search,
			"pageIndex": *page,
			"pageSize":  *size,
		}
		res, err := c.CallTool(ctx, "list_projects_data", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "get":
		fs := flag.NewFlagSet("project get", flag.ContinueOnError)
		fs.SetOutput(stderr)
		idFlag := fs.String("id", "", "Project MongoDB ObjectID")
		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(targetID) == "" {
			output.PrintError(stderr, "project id is required. Usage: scopesentry project get <id> or --id <id>")
			return output.ExitUsageError
		}

		res, err := c.CallTool(ctx, "get_project", map[string]any{"id": targetID})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "create":
		fs := flag.NewFlagSet("project create", flag.ContinueOnError)
		fs.SetOutput(stderr)
		name := fs.String("name", "", "Project name (required); an existing name replaces complete scope. Before adding targets, obtain a complete successful read, preserve all existing targets, and stop if the read is incomplete")
		target := fs.String("target", "", "Scan domains/IPs/CIDRs/URLs separated by actual newline characters (not commas or literal backslash-n) (required)")
		tag := fs.String("tag", "", "Project tag for grouping")
		template := fs.String("template", "", "Associated scan template ObjectID")
		node := fs.String("node", "", "Comma-separated scanning node names")
		allNode := fs.Bool("all-node", false, "Use all available nodes")
		ignore := fs.String("ignore", "", "Ignored targets, format same as target")
		duplicates := fs.String("duplicates", "", "Deduplication policy")
		scheduled := fs.Bool("scheduled", false, "Enable recurring scheduled scan")
		hour := fs.Int("hour", 0, "Scheduled scan interval in hours")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		if strings.TrimSpace(*name) == "" || strings.TrimSpace(*target) == "" {
			output.PrintError(stderr, "both --name and --target are required")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"name":   *name,
			"target": *target,
		}
		if *tag != "" {
			mcpArgs["tag"] = *tag
		}
		if *template != "" {
			mcpArgs["template"] = *template
		}
		if *node != "" {
			mcpArgs["node"] = ParseCommaSeparated(*node)
		}
		if *allNode {
			mcpArgs["allNode"] = true
		}
		if *ignore != "" {
			mcpArgs["ignore"] = *ignore
		}
		if *duplicates != "" {
			mcpArgs["duplicates"] = *duplicates
		}
		if *scheduled {
			mcpArgs["scheduledTasks"] = true
			mcpArgs["hour"] = *hour
		}

		res, err := c.CallTool(ctx, "create_project", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown project subcommand %q. available: list, get, create", subCmd)
		return output.ExitUsageError
	}
}
