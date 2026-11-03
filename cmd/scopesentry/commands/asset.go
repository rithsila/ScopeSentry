package commands

import (
	"context"
	"flag"
	"io"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunAssetCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing asset subcommand. available: list, count, get, add-tag")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "list":
		fs := flag.NewFlagSet("asset list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		assetType := fs.String("type", "", "Asset type (required): asset, subdomain, vulnerability, UrlScan, SensitiveResult, etc.")
		search := fs.String("search", "", "Search DSL expression")
		page := fs.Int("page", 1, "Page number starting from 1")
		size := fs.Int("size", 20, "Items per page")
		filter := fs.String("filter", "", "Exact filter JSON: e.g. {\"project\":[\"id1\"]}")
		sort := fs.String("sort", "", "Sort map or key:direction (e.g. length:descending)")
		sid := fs.String("sid", "", "Sensitive rule name (only for SensitiveResult)")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		if strings.TrimSpace(*assetType) == "" {
			output.PrintError(stderr, "--type is required (e.g. asset, subdomain, vulnerability, UrlScan)")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"asset_type": *assetType,
			"pageIndex":  *page,
			"pageSize":   *size,
		}
		if *search != "" {
			mcpArgs["search"] = *search
		}
		if *sid != "" {
			mcpArgs["sid"] = *sid
		}
		if *filter != "" {
			parsedFilter, err := ParseFilterMap(*filter)
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitUsageError
			}
			mcpArgs["filter"] = parsedFilter
		}
		if *sort != "" {
			parsedSort, err := ParseSortMap(*sort)
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitUsageError
			}
			mcpArgs["sort"] = parsedSort
		}

		res, err := c.CallTool(ctx, "list_assets", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "count":
		fs := flag.NewFlagSet("asset count", flag.ContinueOnError)
		fs.SetOutput(stderr)
		assetType := fs.String("type", "", "Asset type (required): asset, subdomain, vulnerability, etc.")
		search := fs.String("search", "", "Search DSL expression")
		filter := fs.String("filter", "", "Exact filter JSON")
		sid := fs.String("sid", "", "Sensitive rule name (only for SensitiveResult)")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		if strings.TrimSpace(*assetType) == "" {
			output.PrintError(stderr, "--type is required")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"asset_type": *assetType,
		}
		if *search != "" {
			mcpArgs["search"] = *search
		}
		if *sid != "" {
			mcpArgs["sid"] = *sid
		}
		if *filter != "" {
			parsedFilter, err := ParseFilterMap(*filter)
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitUsageError
			}
			mcpArgs["filter"] = parsedFilter
		}

		res, err := c.CallTool(ctx, "count_assets", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "get":
		fs := flag.NewFlagSet("asset get", flag.ContinueOnError)
		fs.SetOutput(stderr)
		assetType := fs.String("type", "", "Asset type (required): asset or vulnerability")
		idFlag := fs.String("id", "", "Asset MongoDB ObjectID or vulnerability hash")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(*assetType) == "" || strings.TrimSpace(targetID) == "" {
			output.PrintError(stderr, "both --type and asset id are required. Usage: scopesentry asset get --type <type> <id>")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"asset_type": *assetType,
			"id":         targetID,
		}
		res, err := c.CallTool(ctx, "get_asset_detail", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "add-tag":
		fs := flag.NewFlagSet("asset add-tag", flag.ContinueOnError)
		fs.SetOutput(stderr)
		assetType := fs.String("type", "", "Asset type (required)")
		idFlag := fs.String("id", "", "Asset MongoDB ObjectID")
		tag := fs.String("tag", "", "Tag to add (required)")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(*assetType) == "" || strings.TrimSpace(targetID) == "" || strings.TrimSpace(*tag) == "" {
			output.PrintError(stderr, "--type, asset id, and --tag are required. Usage: scopesentry asset add-tag --type <type> --id <id> --tag <tag>")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"asset_type": *assetType,
			"id":         targetID,
			"tag":        *tag,
		}
		res, err := c.CallTool(ctx, "add_asset_tag", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown asset subcommand %q. available: list, count, get, add-tag", subCmd)
		return output.ExitUsageError
	}
}
