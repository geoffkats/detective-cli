# Detective - Forensic Code Investigation Tool

**This is not a linter. This is a forensic investigation tool.**

Detective analyzes codebases as *crime scenes*, generating reasoned forensic findings based on evidence. It frames code quality through investigative narratives, not metric dashboards.

A cross-platform CLI tool that generates professional, evidence-based investigation reports for technical audits, due diligence, and post-mortems.

## What Makes Detective Different

Unlike traditional linters or code analyzers, Detective:
- **Generates forensic narratives** - Interprets evidence through investigator insights, not rules
- **Timeline-aware analysis** - Detects development activity patterns even without git history
- **Context-sensitive evaluation** - Adjusts severity based on project maturity (student vs enterprise)
- **Audit-grade reports** - Includes integrity hashes and professional formatting
- **Reasoned inference** - Transforms raw data into meaningful findings

## Features

### Core Analysis
- **File System Analysis** - Smart categorization of files (source, assets, dependencies, artifacts, config, docs, tests); identifies large files contextually
- **Code Marker Detection** - Identifies TODO, FIXME, HACK, BUG, NOTE markers with severity assessment and actionable resolution steps
- **Git History Analysis** - Commit frequency patterns, top contributors, commit message quality scoring, uncommitted changes, branch count, activity trends
- **Timeline Analysis** - Activity bursts, dormancy patterns, development velocity, file modification windows
- **Project Type Detection** - Auto-detects frameworks (Laravel, Node.js, Django, Docker) with framework-specific recommendations

### Security & Quality
- **Security Analysis** - Hardcoded secret detection (API keys, passwords, tokens), SQL injection patterns, XSS vulnerabilities, insecure crypto flags, weak protocols
- **Health Scoring (Weighted)** - Transparent 100-point system with breakdown: Version Control (20), Code Quality (25), Security (20), Performance (15), Documentation (10), Testing (10)
- **Actionable Recommendations** - Every finding includes specific next steps to remediate (e.g., "Add missing .gitignore", "Use parameterized queries")
- **Context Modes** - Evaluates projects differently based on maturity (student, enterprise, default)

### Reporting & UX
- **Multiple Export Formats** - Text, JSON, and Markdown with optional styling and colors
- **Report Integrity** - SHA256 hash verification for audit trails and tamper detection
- **Interactive Mode** - Rescan findings (R), set watch interval (W), quit (Q) from menu after report
- **Performance-Optimized** - Fast mode (-fast) and custom exclusions (-exclude) to skip heavy directories
- **Watch Mode** - Periodic rescanning with -watch flag or toggle from interactive menu

## Installation

### Prerequisites
- Go 1.21 or higher ([download here](https://golang.org/dl/))

### Quick Install (Recommended)

**Windows:**
1. Clone or download this repository
2. Open the folder in File Explorer
3. Double-click `install.bat`
4. Follow the prompts
5. Open a **new** terminal and run `detective -verbose`

**macOS/Linux:**
```bash
go install github.com/geoffkats/detective-cli/cmd/detective@latest
```

Then add `~/go/bin` to your PATH if not already there.

### Manual Build
```bash
cd detective-case
go mod download
go install ./cmd/detective
```

Then add `$USERPROFILE\go\bin` (Windows) or `~/go/bin` (macOS/Linux) to your system PATH.

## Usage

```bash
detective [flags]
```

**Note:** After installation, open a **new terminal window** before using `detective`.

### Flags

**Investigation Scope & Output**
- `-path string` - Directory to investigate (default: current working directory)
- `-format string` - Report format: `text`, `json`, `markdown` (default: text)
- `-export string` - Export report to file (e.g., report.json, report.md)
- `-context string` - Evaluation context: `default`, `student`, `enterprise` (default: default)

**Filtering & Severity**
- `-severity string` - Filter findings by minimum severity: `low`, `medium`, `high`, `critical`
- `-markers` - Include detailed code markers section in report

**Performance & Optimization**
- `-fast` - Fast mode: skip vendor/, node_modules/, .git/, bin/, and hidden files/dirs
- `-exclude string` - Comma-separated list of directories to exclude (e.g., `-exclude vendor,node_modules,dist`)

**CLI Experience**
- `-verbose` - Show analysis progress and detailed output
- `-no-color` - Disable colored text output
- `-watch int` - Watch mode: auto-rescan every N seconds (interactive or toggle from menu)
- `-no-interactive` - Skip interactive menu; print report and exit

### Examples

**Basic Investigation**
```bash
detective                          # Quick scan current directory
detective -path /path/to/project   # Scan specific directory
detective -verbose                 # Show progress during scan
```

**Performance Optimization**
```bash
detective -fast                    # Skip vendor, node_modules, .git, hidden dirs
detective -exclude vendor,dist     # Custom directory exclusions
detective -fast -exclude build     # Combined: fast mode + extra exclusions
```

**Context-Aware Analysis**
```bash
detective -context student -verbose        # Lenient scoring for learning projects
detective -context enterprise -severity high # Strict audit for production systems
```

**Export & Reporting**
```bash
detective -format json -export report.json                    # JSON export
detective -format markdown -markers -export report.md         # Markdown with markers
detective -format text -severity high -export critical.txt    # High-severity findings only
```

**Security & Compliance**
```bash
detective -severity high            # Show only high and critical issues
detective -fast -context enterprise # Production-grade fast audit
```

**Watch & Interactive**
```bash
detective -watch 5                  # Auto-rescan every 5 seconds (watch mode)
detective -verbose                  # Enter interactive menu: [R]escan, [W]atch, [Q]uit
detective -no-interactive           # Batch mode: print and exit (for CI/CD)
```

**Real-World Scenarios**
```bash
# Laravel project audit
detective -fast -context enterprise -format json -export laravel-audit.json

# Node.js portfolio review
detective -exclude node_modules -markers -format markdown -export review.md

# CI/CD continuous monitoring
detective -fast -no-interactive -severity high

# Development watch mode (rescan every 30 seconds)
detective -watch 30 -verbose
```

## Report Sections

### Text/Markdown Report Structure
1. **Banner** - ASCII art header with tool branding
2. **Header** - Investigation metadata, target path, timestamp
3. **Evidence Summary** - File counts, git stats, code markers, security findings
4. **Findings** - Severity-ranked (CRITICAL → LOW) with:
   - Evidence (concrete data)
   - Recommendations (actionable next steps)
   - Category (Code Quality, Security, Version Control, etc.)
5. **Code Markers Detail** - TODO, FIXME, HACK, BUG locations (optional with -markers)
6. **Health Assessment** - Score (0-100) with weighted breakdown:
   - Version Control (20 pts)
   - Code Quality (25 pts)
   - Security (20 pts)
   - Performance (15 pts)
   - Documentation (10 pts)
   - Testing (10 pts)
7. **Investigator Notes** - Forensic insights gleaned from patterns
8. **Report Integrity** - SHA256 hash, context, and verification note

### JSON Export
Structured data suitable for integration with CI/CD, dashboards, and automated reporting:
```json
{
  "targetPath": "...",
  "investigatedAt": "...",
  "findings": [ { "severity", "title", "description", "recommendations" } ],
  "healthScore": 65,
  "healthBreakdown": { "versionControl": 20, "codeQuality": 18, ... },
  "reportHash": "sha256:..."
}
```

## Severity Levels

- **CRITICAL** - Immediate action required (hardcoded secrets, SQL injection, blocked deployments)
- **HIGH** - Significant quality/security impact (no version control, known bugs, large code files)
- **MEDIUM** - Notable concerns (stale repos, high technical debt, missing tests, XSS patterns)
- **LOW** - Minor improvements (missing docs, high commit frequency, limited history)

**Severity Coloring in Terminal Output**
- CRITICAL: RED
- HIGH: RED
- MEDIUM: YELLOW
- LOW: GREEN

## Design Philosophy

Detective operates as a forensic investigation platform:

- **Forensic Not Prescriptive** - Doesn't tell you what to do, shows what it found
- **Evidence-Based** - All findings tied to concrete evidence, not arbitrary rules
- **Context-Aware** - Adjusts analysis based on project maturity and evaluation context
- **Narrative-Driven** - Generates reasoned inferences, not rule violations
- **Professional** - Suitable for audits, client reports, and legal proceedings

Detective is a non-interactive, stateless tool that:
- Runs independently on each execution
- Produces deterministic, evidence-based output
- Maintains a professional, neutral investigative tone
- Focuses on objective analysis, not subjective opinions

## Context Modes

### Default
Standard evaluation with balanced severity assessment.

### Student
More lenient severity for learning projects:
- Reduces HIGH → MEDIUM for version control and contributor issues
- Focuses on learning opportunities rather than hard requirements

### Enterprise
Stricter evaluation for production systems:
- Elevates MEDIUM → HIGH for version control and repository staleness
- Prioritizes stability, collaboration, and maintenance practices

## When to Use Detective

- **Due Diligence** - Evaluate acquired/inherited codebases before onboarding
- **Post-Mortems** - Forensic analysis after system failures or incidents
- **Code Review Prep** - Comprehensive health assessment before manual review
- **Security Audits** - Identify hardcoded secrets, injection vulnerabilities, weak practices
- **Portfolio Review** - Evaluate your own projects objectively for strengths/gaps
- **Compliance & Audit Trails** - Generate timestamped, hash-verified reports for legal/regulatory requirements
- **Enterprise Assessments** - Context-aware evaluation for different maturity levels
- **CI/CD Integration** - Automated continuous assessment in build pipelines
- **Onboarding** - Quick project health snapshot for new team members

## Technical Architecture

```
Detective/
├── cmd/detective/          # CLI entry point and main logic
├── internal/
│   ├── scanner/            # File system, timeline, marker detection, categorization
│   ├── git/                # Git repository analysis
│   ├── detector/           # Project type detection
│   ├── security/           # Secret and vuln pattern scanning
│   ├── inference/          # Evidence → findings, scoring, recommendations
│   └── reporter/           # Report generation and formatting
└── pkg/models/             # Shared data structures
```

## Cross-Platform Support

Detective uses Go's `filepath` package for cross-platform path handling and works on Windows, macOS, and Linux.

## Comparison with Other Tools

| Feature | Detective | Linter | Code Scanner | Git Analyzer |
|---------|-----------|--------|--------------|--------------|
| Forensic Narrative | ✓ | ✗ | ✗ | ✗ |
| Activity Timeline | ✓ | ✗ | ✗ | ~ |
| Context-Aware | ✓ | ✗ | ✗ | ✗ |
| Investigator Insights | ✓ | ✗ | ✗ | ✗ |
| Report Integrity Hash | ✓ | ✗ | ✗ | ✗ |
| Multiple Export Formats | ✓ | ~ | ~ | ~ |

## License

MIT

## Contributing

Contributions welcome! Focus areas:
- Additional timeline analysis heuristics
- New forensic insight patterns
- Additional export formats
- Extended context modes
