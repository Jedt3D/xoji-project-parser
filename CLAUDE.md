# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Xoji** is a Go toolchain that reduces AI agent token costs on large Xojo projects (100+ files) by 5–8×. Instead of having AI agents blindly read entire files and scan all dependencies, xoji pre-indexes projects and serves only relevant context. This shifts productive work from ~7% of tokens to ~50–60%.

The toolchain consists of three indexing tools that generate `.xojo_index/` output files with structured metadata, and a fourth tool (Smart Context Assembler) that uses these indexes at runtime.

### The Problem: Token Budget Without Optimization (100+ file project)

| Task | ~% of tokens | Notes |
|------|-------------|-------|
| Reading full file contents | ~35% | Reads whole files to find one block |
| Parsing #tag syntax / GUI layout | ~20% | Xojo format is verbose |
| Finding class dependencies | ~15% | Scans many files blindly |
| Understanding file structure | ~10% | Re-scans on every new task |
| Locating the right file | ~10% | Trial and error |
| **Actual task: writing code** | **~7%** | **The only part that matters** |
| Verifying result | ~3% | |

**Result**: Only 7–10% of tokens do productive work; 90–93% is overhead.

## Build Commands

Build cross-platform binaries from the repository root:

```bash
# macOS (ARM64 — Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o xoji-mac-arm64 .

# macOS (x86-64 — Intel)
GOOS=darwin GOARCH=amd64 go build -o xoji-mac-amd64 .

# Windows (x86-64)
GOOS=windows GOARCH=amd64 go build -o xoji-windows.exe .

# Linux (x86-64)
GOOS=linux GOARCH=amd64 go build -o xoji-linux .
```

All builds use pure Go stdlib — no CGO or external dependencies required. Cross-compilation works out of the box.

## Architecture & Implementation Order

The xoji binary provides three subcommands (eventually four):

### Tool 1: Project Indexer
**Status**: Plan priority 2nd
**Output**: `.xojo_index/manifest.json`
**Purpose**: Flat manifest of all files in the project with metadata
**Token savings**: ~10% (eliminates blind file listing and structure scanning on every task)

**Contents per file entry**:
- File path and type (`.xojo_code`, `.xojo_window`, `.xojo_menu`, `.xojo_toolbar`)
- File size and last modified timestamp
- Top-level entity name (class name, window name, module name)
- Short description of contents

### Tool 2: Dependency Graph Builder
**Status**: Plan priority 3rd
**Output**: `.xojo_index/dependencies.json`
**Purpose**: Extract class relationships (inheritance, interface implementations, method calls, module usage)
**Token savings**: ~15% (replaces blind search across all class files)

**Contents**:
- Each class and what it inherits from
- Which other classes each class calls methods on
- Which modules are used by each class
- Reverse map: which classes depend on a given class

**Example**: Given a task about `OrderForm`, agent queries the graph and immediately knows it needs to read `BaseForm`, `DatabaseManager`, and `CustomerValidator` — skipping the other 97 class files.

### Tool 3: #tag Code Tree Navigator ⭐ BUILD THIS FIRST
**Status**: Plan priority 1st (highest ROI)
**Output**: `.xojo_index/codetree.json`
**Purpose**: Parse every #tag block in all .xojo_code and .xojo_window files and map them with line numbers
**Token savings**: ~35% (largest single saving)

**Sample output for a window file**:
```json
{
  "MainWindow.xojo_window": {
    "controls": ["Button1", "TextField1", "ListBox1", "Label1"],
    "events": {
      "Opening": 45,
      "Button1.Pressed": 78,
      "ListBox1.Change": 112
    },
    "methods": {
      "LoadData": 156,
      "ValidateInput": 188,
      "ClearForm": 210
    },
    "properties": {
      "Title": 12,
      "Width": 14,
      "Height": 15
    }
  }
}
```

With this map, an agent can jump directly to line 78 to read `Button1.Pressed` without parsing the entire 3,000-line file.

**Implementation steps** (suggested order):
1. Parse `#tag Class` / `#tag EndClass` blocks from `.xojo_code` files
2. Parse `#tag Method`, `#tag Property`, `#tag Event` with line numbers
3. Parse `Begin`/`End` control hierarchy from `.xojo_window` files
4. Parse `#tag Event` and `#tag Method` blocks within window files
5. Serialize to `.xojo_index/codetree.json`
6. Add `--file` flag for incremental single-file re-indexing

### Tool 4: Smart Context Assembler
**Status**: Plan priority 4th (runtime tool, not yet started)
**Purpose**: Use all three indexes to auto-assemble minimal context for a given task
**Note**: Runs at agent runtime — consumes the three index files to answer queries like "what does OrderForm depend on?"

## Subcommand Reference

```bash
xoji index                            # Full rebuild of all three indexes
xoji index --file path/file.xojo_code # Incremental re-index of one file (<100ms)
xoji check                            # Freshness check (exit 0 = fresh, 1 = stale)
xoji serve                            # Watch mode: auto re-index on file changes
```

### xoji check Details
Compares file modification times in the project against timestamps stored in `meta.json`.
- **Exit 0**: Index is fresh — no re-indexing needed
- **Exit 1**: Index is stale — some files have changed since last index

Used by Option A (Pre-Agent Freshness Check) to decide whether a full re-index is needed.

## Index Output Structure

All output lives in a `.xojo_index/` folder next to the `.xojo_project` file:

```
MyApp/
├── MyApp.xojo_project
└── .xojo_index/
    ├── manifest.json      (Tool 1: file manifest)
    ├── dependencies.json  (Tool 2: class relationships)
    ├── codetree.json      (Tool 3: #tag block structure with line numbers)
    └── meta.json          (timestamps, xoji version, project hash)
```

### meta.json
Contains metadata for freshness checking:
- File modification timestamps for all indexed files
- xoji version used to generate indexes
- Project hash (to detect project changes)

Used by `xoji check` to determine if any files have changed since the last index.

**Note**: Add `.xojo_index/` to `.gitignore` — it is generated output, not source.

## Key Xojo Format Details

Xojo files mix multiple concerns in a single file:
- `.xojo_window` files contain GUI property assignments, event handlers inside #tag blocks, control hierarchy with Begin/End blocks, and metadata
- A single 100-control window can exceed 3000 lines, but an agent typically needs only 20 lines for a given task
- #tag blocks are the primary structuring mechanism — they isolate code by purpose (e.g., `#tag Method`, `#tag Event`, `#tag Property`)

## Implementation Notes

- **Language choice**: Go was chosen for cross-platform static binaries, excellent stdlib for file I/O and JSON, fast compilation, zero runtime dependencies, and binary size ~3–5 MB
- **File parsing**: Use Go's `bufio`, `regexp`, `filepath.Walk`, and `encoding/json` from stdlib only
- **No external packages** — keep dependencies at zero
- Incremental re-indexing via `--file` flag should complete in <100ms

## Trigger Strategy: When to Run xoji

Three-tier approach for keeping indexes fresh:

### Option A: Pre-Agent Freshness Check (Start Here)
Simplest approach. Before Claude Code begins any task:
```bash
xoji check || xoji index
```
- If index is fresh (all file mtimes match stored timestamps), `xoji check` exits 0 — nothing runs
- If stale, a full or incremental re-index runs in milliseconds
- Zero changes required to the Xojo project

**Recommended**: Start with this approach.

### Option B: XojoScript Shell Hook (Add Later)
For maximum freshness during active development, add a hook in Xojo that fires after each file save. In a module or App class:
```xojo
Var sh As New Shell
sh.Execute("xoji index --file " + savedFilePath)
```
Keeps index perfectly in sync with every save. Requires adding one small module to the Xojo project.

### Option C: File Watcher / Serve Mode (Optional)
Run `xoji serve` in the background — watches the project folder using OS file events (inotify on Linux, FSEvents on macOS, etc.) and automatically re-indexes any changed file. Zero changes to Xojo project; ~1–2 second lag after save. Ideal for developers who want fire-and-forget freshness without touching the project.

## Workflow Example: Using the Indexes

**Scenario**: Agent is asked to fix a crash in `OrderForm` when `ListBox1` is empty.

1. **Query dependencies**: Read `.xojo_index/dependencies.json` → find `OrderForm` depends on `DatabaseManager`, `BaseForm`
2. **Find the code**: Read `.xojo_index/codetree.json` → locate `ListBox1.Change` at line 112 in `OrderForm.xojo_window`
3. **Read only what's needed**: Read lines 100–140 of `OrderForm.xojo_window` (skip the entire 3000-line file)
4. **Write the fix**: Make the targeted edit

**Without the index**, steps 1 and 2 would each require scanning multiple files, and step 3 would require reading the entire 3000-line file.

## Integration with Claude Code

### Option A: Manual Setup (Recommended for now)

Add this section to your **project's** CLAUDE.md file:

```markdown
## Using xoji for token-efficient indexing

### Before starting any task:
```bash
# From project root, run freshness check
xoji check || xoji index
```

### Index files in `.xojo_index/`:
- **codetree.json** — Maps file paths to {entity, methods, properties, events, line numbers}
- **manifest.json** — All files with their types and entity names
- **dependencies.json** — Class inheritance and interface relationships
- **meta.json** — Project hash and file modification times (for freshness)

### How to use them:

1. **Find a method/property**: Query `codetree.json`
   ```bash
   cat .xojo_index/codetree.json | grep -A 20 "MainWindow.xojo_window"
   ```
   Returns: `"Button1.Pressed": 78` → method at line 78

2. **Read only what's needed**:
   ```bash
   sed -n '70,90p' AppSrc/MainWindow.xojo_window
   ```
   Skip scanning the entire 3000-line file

3. **Understand relationships**: Query `dependencies.json`
   ```bash
   cat .xojo_index/dependencies.json | grep -A 5 "OrderForm"
   ```
   See what `OrderForm` inherits from and which classes depend on it

### Why this saves tokens:
- Instead of reading entire 3KB–50KB files, read only 20–50 lines
- Instead of blindly scanning all classes, query the dependency graph
- Instead of re-parsing file structure, use pre-computed line numbers
- **Result**: 5–8× fewer tokens per task
```

### Option B: Automated Freshness Check (Shell Hook)

If you use Claude Code with a shell integration, create a pre-task hook:

**File: `.claude/hooks/pre-task.sh`**

```bash
#!/bin/bash
# Auto-refresh xoji indexes before each Claude Code task

PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$PROJECT_ROOT"

# Ensure xoji is in PATH or provide full path
XOJI="./xoji"  # or wherever you keep the binary

if [ -f "$XOJI" ]; then
    "$XOJI" check || "$XOJI" index
fi
```

Make executable:
```bash
chmod +x .claude/hooks/pre-task.sh
```

This hook runs automatically before each task, keeping indexes always fresh without manual intervention.

### Option C: IDE Integration (XojoScript)

If you want indexes to update on every file save in Xojo IDE:

Add to a module in your Xojo project:
```xojo
Sub FileWasSaved(filePath As String)
  Var sh As New Shell
  sh.Execute("xoji index --file """ + filePath + """")
End Sub
```

Keeps indexes perfectly in sync with zero latency.
