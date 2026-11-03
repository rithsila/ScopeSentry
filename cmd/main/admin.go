package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/constants"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/services/apikey"
)

func handleAdminCommands() bool {
	if len(os.Args) < 2 {
		return false
	}
	handled, code := runAdminCommandWithCode(os.Args[1:], os.Stdout, os.Stderr)
	if handled {
		os.Exit(code)
	}
	return false
}

func runAdminCommand(args []string, stdout, stderr io.Writer) bool {
	handled, _ := runAdminCommandWithCode(args, stdout, stderr)
	return handled
}

func runAdminCommandWithCode(args []string, stdout, stderr io.Writer) (bool, int) {
	if len(args) == 0 {
		return false, 0
	}

	cmd := args[0]
	switch cmd {
	case "pluginkey":
		resp := map[string]string{
			"plugin_key": config.GlobalConfig.System.PluginKey,
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Fprintln(stdout, string(data))
		return true, 0

	case "version", "--version", "-v":
		ver := constants.Version
		if ver == "" {
			ver = "dev"
		}
		resp := map[string]string{
			"version": ver,
		}
		data, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Fprintln(stdout, string(data))
		return true, 0

	case "apikey":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "usage: ScopeSentry apikey <create|list|delete> [flags]")
			return true, 1
		}

		subCmd := args[1]
		subArgs := args[2:]
		ctx := context.Background()

		switch subCmd {
		case "create":
			fs := flag.NewFlagSet("apikey create", flag.ContinueOnError)
			fs.SetOutput(stderr)
			name := fs.String("name", "", "Name of the API key")
			if err := fs.Parse(subArgs); err != nil {
				return true, 1
			}
			if strings.TrimSpace(*name) == "" {
				fmt.Fprintln(stderr, "error: --name is required")
				return true, 1
			}

			if !mongodb.IsConnected || mongodb.DB == nil {
				fmt.Fprintln(stderr, "error: mongodb client not connected. verify database connection in config.yaml")
				return true, 1
			}

			svc := apikey.NewService()
			result, err := svc.Create(ctx, *name, "admin-cli")
			if err != nil {
				fmt.Fprintf(stderr, "error creating api key: %v\n", err)
				return true, 1
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(stdout, string(data))
			return true, 0

		case "list":
			if !mongodb.IsConnected || mongodb.DB == nil {
				fmt.Fprintln(stderr, "error: mongodb client not connected. verify database connection in config.yaml")
				return true, 1
			}

			svc := apikey.NewService()
			keys, err := svc.List(ctx)
			if err != nil {
				fmt.Fprintf(stderr, "error listing api keys: %v\n", err)
				return true, 1
			}
			data, _ := json.MarshalIndent(keys, "", "  ")
			fmt.Fprintln(stdout, string(data))
			return true, 0

		case "delete":
			fs := flag.NewFlagSet("apikey delete", flag.ContinueOnError)
			fs.SetOutput(stderr)
			id := fs.String("id", "", "ID of the API key to delete")
			if err := fs.Parse(subArgs); err != nil {
				return true, 1
			}
			if strings.TrimSpace(*id) == "" {
				fmt.Fprintln(stderr, "error: --id is required")
				return true, 1
			}

			if !mongodb.IsConnected || mongodb.DB == nil {
				fmt.Fprintln(stderr, "error: mongodb client not connected. verify database connection in config.yaml")
				return true, 1
			}

			svc := apikey.NewService()
			if err := svc.Delete(ctx, *id); err != nil {
				fmt.Fprintf(stderr, "error deleting api key: %v\n", err)
				return true, 1
			}
			resp := map[string]any{
				"success": true,
				"deleted": *id,
			}
			data, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Fprintln(stdout, string(data))
			return true, 0

		default:
			fmt.Fprintf(stderr, "unknown apikey subcommand %q. available: create, list, delete\n", subCmd)
			return true, 1
		}

	default:
		return false, 0
	}
}
