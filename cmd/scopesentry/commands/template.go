package commands

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunTemplateCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing template subcommand. available: list, get, create")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "list":
		fs := flag.NewFlagSet("template list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		query := fs.String("query", "", "Search templates by name")
		search := fs.String("search", "", "Alias for --query")
		page := fs.Int("page", 1, "Page number starting from 1")
		size := fs.Int("size", 20, "Items per page")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		q := *query
		if q == "" {
			q = *search
		}

		mcpArgs := map[string]any{
			"query":     q,
			"pageIndex": *page,
			"pageSize":  *size,
		}
		res, err := c.CallTool(ctx, "list_scan_templates", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "get":
		fs := flag.NewFlagSet("template get", flag.ContinueOnError)
		fs.SetOutput(stderr)
		idFlag := fs.String("id", "", "Template MongoDB ObjectID")
		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(targetID) == "" {
			output.PrintError(stderr, "template id is required. Usage: scopesentry template get <id> or --id <id>")
			return output.ExitUsageError
		}

		res, err := c.CallTool(ctx, "get_scan_template", map[string]any{"id": targetID})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "create":
		fs := flag.NewFlagSet("template create", flag.ContinueOnError)
		fs.SetOutput(stderr)
		name := fs.String("name", "", "Template name")
		modules := fs.String("modules", "", "Mapping of module to plugin hashes (JSON object or @file, e.g. '{\"SubdomainScan\":[\"hash1\"]}')")
		parameters := fs.String("parameters", "", "Parameter overrides (JSON object or @file)")
		vullist := fs.String("vullist", "", "Nuclei POC template IDs, comma-separated or JSON array")
		templateJSON := fs.String("template-json", "", "Complete ScanTemplate JSON, @filepath, or '-' for stdin")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{}

		if *templateJSON != "" {
			var rawBytes []byte
			var err error
			if *templateJSON == "-" {
				rawBytes, err = io.ReadAll(os.Stdin)
			} else if strings.HasPrefix(*templateJSON, "@") {
				rawBytes, err = os.ReadFile(strings.TrimPrefix(*templateJSON, "@"))
			} else if _, statErr := os.Stat(*templateJSON); statErr == nil {
				rawBytes, err = os.ReadFile(*templateJSON)
			} else {
				rawBytes = []byte(*templateJSON)
			}
			if err != nil {
				output.PrintError(stderr, "failed to read template json: %v", err)
				return output.ExitUsageError
			}
			mcpArgs["template_json"] = string(rawBytes)
			if *name != "" {
				mcpArgs["name"] = *name
			}
		} else {
			if strings.TrimSpace(*name) == "" {
				output.PrintError(stderr, "--name is required when --template-json is not provided")
				return output.ExitUsageError
			}
			mcpArgs["name"] = *name

			if *modules != "" {
				parsedMod, err := ParseJSONOrFile(*modules)
				if err != nil {
					output.PrintError(stderr, "invalid modules JSON: %v", err)
					return output.ExitUsageError
				}
				mcpArgs["modules"] = parsedMod
			}

			if *parameters != "" {
				parsedParams, err := ParseJSONOrFile(*parameters)
				if err != nil {
					output.PrintError(stderr, "invalid parameters JSON: %v", err)
					return output.ExitUsageError
				}
				mcpArgs["parameters"] = parsedParams
			}

			if *vullist != "" {
				if strings.HasPrefix(strings.TrimSpace(*vullist), "[") {
					var list []string
					if err := json.Unmarshal([]byte(*vullist), &list); err == nil {
						mcpArgs["vullist"] = list
					} else {
						mcpArgs["vullist"] = ParseCommaSeparated(*vullist)
					}
				} else {
					mcpArgs["vullist"] = ParseCommaSeparated(*vullist)
				}
			}
		}

		res, err := c.CallTool(ctx, "create_scan_template", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown template subcommand %q. available: list, get, create", subCmd)
		return output.ExitUsageError
	}
}
