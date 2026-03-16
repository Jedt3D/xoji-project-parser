package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
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
	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)
	case "-v", "--version", "version":
		fmt.Println("xoji v1.0.0 - AI-Friendly Xojo Project Indexer")
		os.Exit(0)
	default:
		fmt.Printf("Error: Unknown command '%s'\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`xoji v1.0.0 - AI-Friendly Xojo Project Indexer
Reduce AI agent token costs by 5-8x on large Xojo projects (100+ files)

USAGE:
  xoji <command> [options] [path]

COMMANDS:
  setup [PATH]              Configure a Xojo project for indexing
                            Creates CLAUDE.md, .xojo_index/, and hooks
                            PATH: project directory (optional, searches from current dir)

  index [PATH]              Build or rebuild all indexes
                            PATH: project directory (optional, searches from current dir)
                            Generates: codetree.json, manifest.json, dependencies.json, meta.json
    --file PATH             Re-index only a single file (incremental, <100ms)
    --project PATH          Explicit project path (alternative to positional argument)

  check [PATH]              Check if indexes are fresh
                            PATH: project directory (optional, searches from current dir)
                            Exit code: 0 = fresh, 1 = stale
                            Use in scripts: xoji check || xoji index

  serve [PATH]              Watch mode - auto re-index on file changes (TODO)
                            PATH: project directory (optional, searches from current dir)

  help, -h, --help          Show this help message
  version, -v, --version    Show version information

EXAMPLES:
  # One-time project setup
  xoji setup ../my_xojo_project

  # Build indexes
  xoji index
  xoji index ../project
  xoji index --project /path/to/project

  # Incremental update (single file)
  xoji index --file AppSrc/MainWindow.xojo_window

  # Check freshness and rebuild if needed
  xoji check || xoji index

FEATURES:
  • Indexes .xojo_code files (classes, modules, interfaces)
  • Indexes .xojo_window files (controls, methods, events)
  • Extracts exact line numbers for all elements
  • Builds class dependency graphs
  • Automatic freshness checking (sub-50ms)
  • Incremental updates (<100ms)
  • Cross-platform support (macOS, Windows, Linux)

TOKEN SAVINGS:
  Without xoji: ~93% tokens on navigation, ~7% on actual coding
  With xoji:   ~40% tokens on navigation, ~60% on actual coding
  Result: 5-8x token reduction per AI agent task

DOCUMENTATION:
  See README.md and docs/ for full documentation
  GitHub: https://github.com/Jedt3D/xoji-project-parser

SUPPORT:
  Issues: https://github.com/Jedt3D/xoji-project-parser/issues
  Email: support@example.com`)
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
