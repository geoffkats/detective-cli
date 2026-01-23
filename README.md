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

- **File System Analysis** - Scans directory structure, file counts, sizes, and type distribution
- **Code Marker Detection** - Identifies TODO, FIXME, HACK, BUG, and NOTE markers
- **Git History Analysis** - Examines commit patterns, activity levels, and contributor data
- **Timeline Analysis** - Detects activity bursts, dormancy patterns, and development velocity
- **Investigator Notes** - Generates forensic insights and observations from evidence patterns
- **Context Modes** - Evaluates projects differently based on maturity level (student, enterprise, default)
- **Report Integrity** - Includes SHA256 hash for audit and verification purposes
- **Multiple Formats** - Export reports as text, JSON, or Markdown

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
- `-path string` - Directory to investigate (default: current)
- `-format string` - Output format: text, json, markdown (default: text)
- `-context string` - Evaluation context: default, student, enterprise (default: default)
- `-severity string` - Filter findings: low, medium, high, critical
- `-export string` - Export report to file
- `-markers` - Show detailed code markers list
- `-verbose` - Show progress information

### Examples

Basic investigation:
```bash
detective
```

Student project analysis:
```bash
detective -context student -verbose
```

Enterprise code audit (high severity only):
```bash
detective -context enterprise -severity high
```

Export as JSON:
```bash
detective -format json -export report.json
```

Markdown report with code markers:
```bash
detective -format markdown -markers -export report.md
```

Specific directory with timeline analysis:
```bash
detective -path C:\Projects\my-app -verbose
```

## Report Sections

1. **Header** - Investigation metadata and target information
2. **Evidence Collection** - Raw data from file system, git, timeline, and code scans
3. **Findings** - Severity-ranked issues based on evidence
4. **Health Assessment** - Numerical scoring (0-100) for overall project health
5. **Investigator Notes** - Forensic insights and inferences from patterns
6. **Report Integrity** - Hash and context information for audit trails

## Severity Levels

- **CRITICAL** - Immediate attention required; blocking production concerns
- **HIGH** - Significant issues impacting project quality or security
- **MEDIUM** - Notable concerns requiring review and action
- **LOW** - Minor observations and suggestions for improvement

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

- 🔍 **Due Diligence** - Evaluate acquired codebases before acquisition
- 📋 **Post-Mortems** - Understand what led to system failures
- 🎓 **Code Review Prep** - Assess project health before manual review
- 📊 **Portfolio Review** - Evaluate your own projects objectively
- ⚖️ **Audit Trails** - Generate timestamped, hash-verified reports
- 🏢 **Enterprise Assessments** - Context-aware evaluation for different environments

## Technical Architecture

```
detective/
├── cmd/detective/          # CLI entry point and main logic
├── internal/
│   ├── scanner/           # File system, timeline, and marker detection
│   ├── git/               # Git repository analysis
│   ├── inference/         # Evidence → findings logic & investigator notes
│   └── reporter/          # Report generation and formatting
└── pkg/models/            # Shared data structures
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
