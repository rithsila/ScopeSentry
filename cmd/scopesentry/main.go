package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	code := commands.Execute(ctx, os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}
