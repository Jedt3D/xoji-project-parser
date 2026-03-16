# xoji — AI-Friendly Xojo Project Indexer

**Reduce AI agent token costs by 5–8× on large Xojo projects (100+ files)**

xoji pre-indexes Xojo projects and serves only relevant context to AI agents, shifting productive work from ~7% of tokens to ~50–60%.

## Why xoji?

When an AI agent works on a 100+ file Xojo project without indexing:

| Task | ~% of tokens | Problem |
|------|-------------|---------|
| Reading full file contents | ~35% | Reads whole 3KB–50KB files to find one method |
| Parsing Xojo's verbose #tag syntax | ~20% | Window files alone can exceed 3,000 lines |
| Finding class dependencies | ~15% | Blindly scans all files to understand relationships |
| Understanding file structure | ~10% | Re-scans on every new task |
| Locating the right file | ~10% | Trial and error across dozens of files |
| **Actual coding work** | **~7%** | **The only part that matters** |
| Verifying results | ~3% | Extra scanning for verification |

**Result**: 90–93% of tokens wasted on overhead; only 7–10% does productive work.

---

## How xoji Solves This

xoji creates four lightweight JSON indexes:

1. **codetree.json** (35% savings)
   Maps every file to its methods, properties, events, constants, enums, and hooks—with exact line numbers. Agents jump to line 78 instead of reading 3,000 lines.

2. **manifest.json** (10% savings)
   Directory of all files with their types and entity names. No more blind scanning.

3. **dependencies.json** (15% savings)
   Class inheritance and interface relationships. Agents know instantly what `OrderForm` depends on.

4. **meta.json** (freshness tracking)
   Project hash and file timestamps. Index is auto-rebuilt only when needed.

---

## Installation

### macOS / Linux

```bash
# Clone the repository
git clone git@github.com:Jedt3D/xoji-project-parser.git
cd xoji-project-parser

# Build cross-platform binaries
./build.sh

# The binary is ready in dist/ or ./xoji (current platform)
./xoji --help
```

### Windows

```cmd
# Clone the repository
git clone git@github.com:Jedt3D/xoji-project-parser.git
cd xoji-project-parser

# Build with batch script
build.bat

# Or with PowerShell
powershell -File build.ps1

# The binary is ready in dist\ or xoji.exe (current platform)
xoji.exe --help
```

---

## Quick Start

### 1. Set up a Xojo project

```bash
xoji setup ../my_xojo_project
```

This:
- Creates `.xojo_index/` directory next to the project
- Adds integration instructions to the project's CLAUDE.md
- Installs `.claude/hooks/pre-task.sh` for automatic freshness checking

### 2. Build the indexes

```bash
xoji index ../my_xojo_project
```

This parses all Xojo files and generates four JSON index files.

### 3. Check freshness (automatic)

```bash
xoji check ../my_xojo_project
```

Returns exit 0 (fresh) or 1 (stale) based on file modification times.

---

## Usage

### Commands

```bash
xoji setup [PROJECT_PATH]       # Configure a Xojo project for indexing
xoji index [PROJECT_PATH]       # Build or rebuild all indexes
xoji index --file RELATIVE_PATH # Incremental: re-index just one file
xoji check [PROJECT_PATH]       # Check if indexes are fresh (exit 0|1)
xoji serve [PROJECT_PATH]       # Watch mode (auto re-index on changes) [TODO]
```

### Examples

```bash
# Set up project from its directory
cd /path/to/MyXojoApp
xoji setup .
xoji index

# Or from elsewhere
xoji setup ../MyXojoApp
xoji check ../MyXojoApp

# Incremental update (fast)
xoji index --file AppSrc/MainWindow.xojo_window

# Check if re-indexing needed
xoji check && echo "Indexes are fresh" || echo "Rebuilding..."
```

---

## Using the Indexes in Your Projects

### For AI Agents (Claude Code, etc.)

Add this to your project's **CLAUDE.md**:

```markdown
## Using xoji for token-efficient indexing

### Before starting any task:
```bash
xoji check || xoji index
```

### Index files in `.xojo_index/`:
- **codetree.json** — #tag structure map with exact line numbers
- **manifest.json** — all files and their types
- **dependencies.json** — class inheritance and relationships
- **meta.json** — project hash and file timestamps

### How to use:

1. **Find a method**: Query codetree.json for the line number
   ```bash
   cat .xojo_index/codetree.json | grep -A 20 "MainWindow.xojo_window"
   ```
   Returns: `"Button1.Pressed": 78` → read lines 70–90 only

2. **Understand dependencies**: Query dependencies.json
   ```bash
   cat .xojo_index/dependencies.json | grep "OrderForm"
   ```
   See what `OrderForm` inherits from instantly

3. **Read only what's needed**: Use sed/head instead of scanning entire files
   ```bash
   sed -n '70,90p' AppSrc/MainWindow.xojo_window
   ```
```

### For Build Pipelines

```bash
# CI/CD: Fail if indexes are stale
xoji check || { echo "Stale indexes"; exit 1; }

# Rebuild if needed
xoji check || xoji index

# Verify indexes are valid JSON
jq empty .xojo_index/codetree.json
```

---

## Index File Format

### codetree.json

Maps file paths to parsed structure:

```json
{
  "AppSrc/MainWindow.xojo_window": {
    "type": "DesktopWindow",
    "entity": "MainWindow",
    "controls": [
      { "name": "Button1", "type": "DesktopButton" },
      { "name": "TextField1", "type": "DesktopTextField" }
    ],
    "methods": { "LoadData": 156, "ValidateInput": 188 },
    "events": { "Opening": 45, "Button1.Pressed": 78 },
    "properties": { "mData": 210 }
  },
  "App.xojo_code": {
    "type": "Class",
    "entity": "App",
    "inherits": "Application",
    "implements": ["ILoggable"],
    "methods": { "Constructor": 3, "Run": 12 },
    "properties": { "Name": 20 },
    "events": { "Opening": 5 },
    "constants": { "VERSION": 25 },
    "enums": { "Status": 30 },
    "hooks": {}
  }
}
```

### manifest.json

List of all indexed files:

```json
[
  {
    "path": "App.xojo_code",
    "type": "Class",
    "entity": "App",
    "size": 2341,
    "modTime": "2026-03-17T12:00:00Z",
    "sizeBytes": 2341
  },
  {
    "path": "AppSrc/MainWindow.xojo_window",
    "type": "DesktopWindow",
    "entity": "MainWindow",
    "size": 56539,
    "modTime": "2026-03-17T15:22:14Z",
    "sizeBytes": 56539
  }
]
```

### dependencies.json

Class relationships:

```json
{
  "classes": {
    "MainForm": {
      "inherits": "BaseForm",
      "implements": ["IClosable", "IRefreshable"]
    },
    "DatabaseManager": {
      "inherits": "",
      "implements": []
    }
  },
  "reverse": {
    "BaseForm": ["MainForm", "SettingsForm"],
    "Application": ["App"]
  }
}
```

### meta.json

Metadata for freshness checking:

```json
{
  "version": "1.0.0",
  "indexedAt": "2026-03-17T15:30:00Z",
  "projectHash": "a3b4c5d6e7f8...",
  "files": {
    "App.xojo_code": { "mtime": "2026-03-17T12:00:00Z", "size": 2341 },
    "AppSrc/MainWindow.xojo_window": { "mtime": "2026-03-17T15:22:14Z", "size": 56539 }
  }
}
```

---

## Architecture

### Source Structure

```
xoji-project-parser/
├── src/                    # Go source code
│   ├── main.go
│   ├── cmd_index.go
│   ├── cmd_setup.go
│   ├── parse_code.go       # .xojo_code parser
│   ├── parse_window.go     # .xojo_window parser
│   ├── indexer.go          # Index orchestration
│   ├── types.go
│   └── ...
├── dist/                   # Cross-platform binaries (generated by build.sh)
│   ├── xoji-mac-arm64
│   ├── xoji-mac-amd64
│   ├── xoji-linux
│   └── xoji-windows.exe
├── docs/                   # Documentation
│   ├── README.md           # This file
│   ├── QUICKSTART.md       # Quick start guide
│   └── ARCHITECTURE.md     # Detailed architecture
├── build.sh                # Unix build script
├── build.bat               # Windows batch build
└── go.mod, go.sum
```

---

## Performance

| Metric | Value |
|--------|-------|
| Full index build (40+ files) | ~200ms |
| Incremental update (single file) | <100ms |
| Binary size | 3.7–3.9 MB |
| Index files total | ~100–200 KB |
| Token savings | 5–8× |

---

## Development

### Building from source

```bash
./build.sh      # macOS/Linux
build.bat       # Windows
```

All builds use pure Go stdlib—no external dependencies, no CGO.

### Running tests

```bash
cd src/
go test ./...
```

### Development workflow

1. Edit .go files in `src/`
2. Run `./xoji setup` on a test project
3. Run `./xoji index` to verify
4. Commit and push

---

## Roadmap

- ✅ Tool 1: Project Indexer (manifest.json)
- ✅ Tool 2: Dependency Graph (dependencies.json)
- ✅ Tool 3: #tag Code Tree (codetree.json)
- ✅ Tool 4: CLI setup and check commands
- ⏳ Tool 5: Watch mode (`xoji serve`)
- ⏳ Smart Context Assembler (runtime tool for agents)
- ⏳ IDE integration hooks
- ⏳ Web UI for index visualization

---

## License

MIT

---

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Test on both Windows and Unix
4. Submit a pull request

---

## Support

For issues, feature requests, or questions:
- GitHub Issues: https://github.com/Jedt3D/xoji-project-parser/issues
- Email: support@example.com

---

**Built with ❤️ for Xojo developers and AI agents**
