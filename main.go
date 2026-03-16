package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: xoji <command> [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  setup      Add xoji instructions to CLAUDE.md and create hooks")
		fmt.Println("  index      Build or rebuild indexes (codetree, manifest, dependencies)")
		fmt.Println("  check      Check if indexes are fresh (exit 0) or stale (exit 1)")
		fmt.Println("  serve      Watch mode - auto re-index on file changes")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "setup":
		runSetup()
	case "index":
		runIndex()
	case "check":
		runCheck()
	case "serve":
		runServe()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runSetup() {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	projectPath := fs.String("project", "", "Path to Xojo project (default: search from current dir)")

	fs.Parse(os.Args[2:])

	// If projectPath not provided via flag, use first positional argument
	if *projectPath == "" && fs.NArg() > 0 {
		*projectPath = fs.Arg(0)
	}

	if err := CmdSetup(*projectPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runIndex() {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	projectPath := fs.String("project", "", "Path to .xojo_project file (default: search from current dir)")
	filePath := fs.String("file", "", "Optional: re-index only this single file (relative path)")

	fs.Parse(os.Args[2:])

	// If projectPath not provided via flag, use first positional argument
	if *projectPath == "" && fs.NArg() > 0 {
		*projectPath = fs.Arg(0)
	}

	if err := CmdIndex(*projectPath, *filePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCheck() {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	projectPath := fs.String("project", "", "Path to .xojo_project file (default: search from current dir)")

	fs.Parse(os.Args[2:])

	// If projectPath not provided via flag, use first positional argument
	if *projectPath == "" && fs.NArg() > 0 {
		*projectPath = fs.Arg(0)
	}

	fresh, err := CmdCheck(*projectPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !fresh {
		os.Exit(1) // Index is stale
	}
	// Exit 0 if fresh
}

func runServe() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	projectPath := fs.String("project", "", "Path to .xojo_project file (default: search from current dir)")

	fs.Parse(os.Args[2:])

	// If projectPath not provided via flag, use first positional argument
	if *projectPath == "" && fs.NArg() > 0 {
		*projectPath = fs.Arg(0)
	}

	if err := CmdServe(*projectPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
