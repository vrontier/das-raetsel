package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"das-raetsel/internal/app"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			log.Fatalf("serve failed: %v", err)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "HTTP listen address")
	dbPath := fs.String("db", "data/das-raetsel.db", "SQLite database path")
	storyPath := fs.String("story", "stories/intro.v2.yaml", "Story file path")
	templateGlob := fs.String("templates", "web/templates/*.html", "Template glob")
	staticDir := fs.String("static", "web/static", "Static files directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(*dbPath, *storyPath, *templateGlob, *staticDir)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			log.Printf("close warning: %v", closeErr)
		}
	}()

	if err := application.Serve(ctx, *addr); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func printUsage() {
	fmt.Println("Das Raetsel - Web Textadventure")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  das-raetsel serve [flags]")
	fmt.Println()
	fmt.Println("Flags (serve):")
	fmt.Println("  -addr      HTTP listen address (default :8080)")
	fmt.Println("  -db        SQLite database path (default data/das-raetsel.db)")
	fmt.Println("  -story     Story file path (default stories/intro.v2.yaml)")
	fmt.Println("  -templates Template glob (default web/templates/*.html)")
	fmt.Println("  -static    Static files directory (default web/static)")
}
