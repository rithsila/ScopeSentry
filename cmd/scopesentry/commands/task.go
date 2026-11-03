package commands

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"strings"
	"time"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func RunTaskCommand(ctx context.Context, c client.Client, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		output.PrintError(stderr, "missing task subcommand. available: list, get, create, wait, stop, resume")
		return output.ExitUsageError
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "stop", "resume":
		fs := flag.NewFlagSet("task "+subCmd, flag.ContinueOnError)
		fs.SetOutput(stderr)
		var targetID string
		idCount := 0
		fs.Func("id", "Single task MongoDB ObjectID", func(value string) error {
			idCount++
			targetID = value
			return nil
		})
		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}
		if idCount == 0 && fs.NArg() == 1 {
			targetID = fs.Arg(0)
		} else if idCount != 1 || fs.NArg() != 0 {
			output.PrintError(stderr, "exactly one task id is required. Usage: scopesentry task %s <id> or --id <id>", subCmd)
			return output.ExitUsageError
		}
		if decoded, err := hex.DecodeString(targetID); err != nil || len(decoded) != 12 {
			output.PrintError(stderr, "task id must be a single valid MongoDB ObjectID")
			return output.ExitUsageError
		}
		res, err := c.CallTool(ctx, subCmd+"_scan_task", map[string]any{"id": targetID})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "list":
		fs := flag.NewFlagSet("task list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		search := fs.String("search", "", "Search tasks by name")
		page := fs.Int("page", 1, "Page number starting from 1")
		size := fs.Int("size", 20, "Items per page")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"search":    *search,
			"pageIndex": *page,
			"pageSize":  *size,
		}
		res, err := c.CallTool(ctx, "list_tasks", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "get":
		fs := flag.NewFlagSet("task get", flag.ContinueOnError)
		fs.SetOutput(stderr)
		idFlag := fs.String("id", "", "Task MongoDB ObjectID")
		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(targetID) == "" {
			output.PrintError(stderr, "task id is required. Usage: scopesentry task get <id> or --id <id>")
			return output.ExitUsageError
		}

		res, err := c.CallTool(ctx, "get_task", map[string]any{"id": targetID})
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "create":
		fs := flag.NewFlagSet("task create", flag.ContinueOnError)
		fs.SetOutput(stderr)
		name := fs.String("name", "", "Task name, unique (required)")
		node := fs.String("node", "", "Comma-separated online scanning node names (required unless --all-node)")
		target := fs.String("target", "", "Scan targets separated by actual newline characters (not commas or literal backslash-n)")
		template := fs.String("template", "", "Scan template ObjectID")
		allNode := fs.Bool("all-node", false, "Include all online nodes")
		ignore := fs.String("ignore", "", "Ignored targets")
		duplicates := fs.String("duplicates", "", "Deduplication policy")
		targetSource := fs.String("target-source", "general", "Target source: general, project, asset, etc.")
		project := fs.String("project", "", "Comma-separated project ObjectIDs (when target-source is project)")
		targetTp := fs.String("target-tp", "", "Selection type: search or select")
		search := fs.String("search", "", "Search DSL expression for target filtering")
		filter := fs.String("filter", "", "Filter JSON map: e.g. {\"project\":[\"id1\"]}")
		targetNumber := fs.Int("target-number", 0, "Max target count limit in search mode")
		targetIds := fs.String("target-ids", "", "Comma-separated target asset ObjectIDs in select mode")
		bindProject := fs.String("bind-project", "", "Associated project ObjectID to bind findings to")
		scheduled := fs.Bool("scheduled", false, "Unsupported: recurring scan requests are rejected by MCP")
		hour := fs.Int("hour", 0, "Scheduled interval - hours")
		minute := fs.Int("minute", 0, "Scheduled interval - minutes")
		day := fs.Int("day", 0, "Scheduled interval - days")
		week := fs.Int("week", 0, "Scheduled interval - week")
		cycleType := fs.String("cycle-type", "", "Cycle type: nhours, daily, weekly")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		if strings.TrimSpace(*name) == "" || (!*allNode && strings.TrimSpace(*node) == "") {
			output.PrintError(stderr, "--name and --node (or --all-node) are required")
			return output.ExitUsageError
		}

		mcpArgs := map[string]any{
			"name":         *name,
			"targetSource": *targetSource,
		}
		if nodes := ParseCommaSeparated(*node); len(nodes) > 0 {
			mcpArgs["node"] = nodes
		}
		if *target != "" {
			mcpArgs["target"] = *target
		}
		if *template != "" {
			mcpArgs["template"] = *template
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
		if *project != "" {
			mcpArgs["project"] = ParseCommaSeparated(*project)
		}
		if *targetTp != "" {
			mcpArgs["targetTp"] = *targetTp
		}
		if *search != "" {
			mcpArgs["search"] = *search
		}
		if *targetNumber > 0 {
			mcpArgs["targetNumber"] = *targetNumber
		}
		if *targetIds != "" {
			mcpArgs["targetIds"] = ParseCommaSeparated(*targetIds)
		}
		if *bindProject != "" {
			mcpArgs["bindProject"] = *bindProject
		}
		if *scheduled {
			mcpArgs["scheduledTasks"] = true
			mcpArgs["hour"] = *hour
			mcpArgs["minute"] = *minute
			mcpArgs["day"] = *day
			mcpArgs["week"] = *week
			mcpArgs["cycleType"] = *cycleType
		}
		if *filter != "" {
			parsedFilter, err := ParseFilterMap(*filter)
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitUsageError
			}
			mcpArgs["filter"] = parsedFilter
		}

		res, err := c.CallTool(ctx, "create_scan_task", mcpArgs)
		if err != nil {
			output.PrintError(stderr, "%v", err)
			return output.ExitExecutionError
		}
		_ = output.PrintRawJSON(stdout, client.ExtractText(res))
		return output.ExitSuccess

	case "wait":
		fs := flag.NewFlagSet("task wait", flag.ContinueOnError)
		fs.SetOutput(stderr)
		idFlag := fs.String("id", "", "Task MongoDB ObjectID")
		timeout := fs.Duration("timeout", 10*time.Minute, "Maximum wait time (e.g. 10m, 30s)")
		interval := fs.Duration("interval", 5*time.Second, "Polling interval (e.g. 5s, 2s)")

		if err := ParseFlags(fs, subArgs); err != nil {
			return output.ExitUsageError
		}

		targetID := *idFlag
		if targetID == "" && fs.NArg() > 0 {
			targetID = fs.Arg(0)
		}

		if strings.TrimSpace(targetID) == "" {
			output.PrintError(stderr, "task id is required. Usage: scopesentry task wait <id> or --id <id>")
			return output.ExitUsageError
		}

		if *interval <= 0 {
			*interval = 2 * time.Second
		}

		output.PrintDiagnostic(stderr, "waiting for task %s (timeout: %v, interval: %v)...", targetID, *timeout, *interval)

		deadline := time.Now().Add(*timeout)
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		for {
			res, err := c.CallTool(ctx, "get_task", map[string]any{"id": targetID})
			if err != nil {
				output.PrintError(stderr, "%v", err)
				return output.ExitExecutionError
			}

			rawText := client.ExtractText(res)
			var taskData struct {
				Status   int `json:"status"`
				Progress int `json:"progress"`
			}
			if err := json.Unmarshal([]byte(rawText), &taskData); err == nil {
				output.PrintDiagnostic(stderr, "[task %s] status: %d, progress: %d%%", targetID, taskData.Status, taskData.Progress)
				// Status 3 is complete, or progress >= 100
				if taskData.Status == 3 || taskData.Progress >= 100 {
					_ = output.PrintRawJSON(stdout, rawText)
					return output.ExitSuccess
				}
			} else {
				output.PrintDiagnostic(stderr, "[task %s] polling...", targetID)
			}

			if time.Now().After(deadline) {
				output.PrintError(stderr, "timeout: task %s did not complete within %v (exiting without task cancellation)", targetID, *timeout)
				return output.ExitTimeout
			}

			select {
			case <-ctx.Done():
				output.PrintError(stderr, "context cancelled while waiting for task: %v", ctx.Err())
				return output.ExitExecutionError
			case <-ticker.C:
			}
		}

	default:
		output.PrintError(stderr, "unknown task subcommand %q. available: list, get, create, wait, stop, resume", subCmd)
		return output.ExitUsageError
	}
}
