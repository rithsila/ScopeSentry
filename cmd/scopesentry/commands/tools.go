package commands

import (
	"context"
	"flag"
	"io"
	"strings"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunToolsCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing tools subcommand. available: list, describe, call")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "list":
		tools, err := c.ListTools(ctx)
		if err != nil {
			output.PrintError(stderr, "failed to list tools: %v", err)
			return output.ExitExecutionError
		}

		type toolItem struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			InputSchema any    `json:"inputSchema,omitempty"`
		}
		items := make([]toolItem, 0, len(tools))
		for _, t := range tools {
			items = append(items, toolItem{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.InputSchema,
			})
		}
		_ = output.PrintJSON(stdout, items)
		return output.ExitSuccess

	case "describe":
		if len(subArgs) == 0 {
			output.PrintError(stderr, "tool name is required. Usage: scopesentry tools describe <tool-name>")
			return output.ExitUsageError
		}
		toolName := strings.TrimSpace(subArgs[0])

		tools, err := c.ListTools(ctx)
		if err != nil {
			output.PrintError(stderr, "failed to list tools: %v", err)
			return output.ExitExecutionError
		}

		for _, t := range tools {
			if t.Name == toolName {
				_ = output.PrintJSON(stdout, map[string]any{
					"name":        t.Name,
					"description": t.Description,
					"inputSchema": t.InputSchema,
				})
				return output.ExitSuccess
			}
		}

		output.PrintError(stderr, "tool %q not found", toolName)
		return output.ExitExecutionError

	case "call":
		if len(subArgs) == 0 {
			output.PrintError(stderr, "tool name is required. Usage: scopesentry tools call <tool-name> [--input <json>] [--input-file <path>]")
			return output.ExitUsageError
		}
		toolName := subArgs[0]

		fs := flag.NewFlagSet("tools call", flag.ContinueOnError)
		fs.SetOutput(stderr)
		inputStr := fs.String("input", "", "JSON string arguments")
		argsStr := fs.String("args", "", "JSON string arguments (alias for --input)")
		inputFile := fs.String("input-file", "", "File path containing JSON arguments, or '-' for stdin")

		if err := ParseFlags(fs, subArgs[1:]); err != nil {
			return output.ExitUsageError
		}

		var argsObj any = map[string]any{}
		var rawArgs string
		if *inputFile != "" {
			parsed, err := ParseJSONOrFile(*inputFile)
			if err != nil {
				output.PrintError(stderr, "failed to read input file: %v", err)
				return output.ExitUsageError
			}
			if parsed != nil {
				argsObj = parsed
			}
		} else if *inputStr != "" {
			rawArgs = *inputStr
		} else if *argsStr != "" {
			rawArgs = *argsStr
		}

		if rawArgs != "" {
			parsed, err := ParseJSONOrFile(rawArgs)
			if err != nil {
				output.PrintError(stderr, "failed to parse input JSON: %v", err)
				return output.ExitUsageError
			}
			if parsed != nil {
				argsObj = parsed
			}
		}

		res, err := c.CallTool(ctx, toolName, argsObj)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	default:
		output.PrintError(stderr, "unknown tools subcommand %q. available: list, describe, call", subCmd)
		return output.ExitUsageError
	}
}
