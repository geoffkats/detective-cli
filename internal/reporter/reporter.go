package reporter

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/detective-cli/detective/pkg/models"
	"github.com/fatih/color"
)

type styler struct {
	enabled bool
	header  func(string, ...interface{}) string
	section func(string, ...interface{}) string
	label   func(string, ...interface{}) string
	dim     func(string, ...interface{}) string
	high    func(string, ...interface{}) string
	medium  func(string, ...interface{}) string
	low     func(string, ...interface{}) string
	info    func(string, ...interface{}) string
}

func newStyler(enabled bool) styler {
	if !enabled {
		return styler{
			enabled: false,
			header:  fmt.Sprintf,
			section: fmt.Sprintf,
			label:   fmt.Sprintf,
			dim:     fmt.Sprintf,
			high:    fmt.Sprintf,
			medium:  fmt.Sprintf,
			low:     fmt.Sprintf,
			info:    fmt.Sprintf,
		}
	}

	return styler{
		enabled: true,
		header:  color.New(color.FgCyan, color.Bold).SprintfFunc(),
		section: color.New(color.FgHiWhite, color.Bold).SprintfFunc(),
		label:   color.New(color.FgHiWhite).SprintfFunc(),
		dim:     color.New(color.FgHiBlack).SprintfFunc(),
		high:    color.New(color.FgRed, color.Bold).SprintfFunc(),
		medium:  color.New(color.FgYellow).SprintfFunc(),
		low:     color.New(color.FgCyan).SprintfFunc(),
		info:    color.New(color.FgWhite).SprintfFunc(),
	}
}

// GenerateReport creates a formatted forensic investigation report (plain)
func GenerateReport(report models.Report) string {
	return GenerateReportStyled(report, false)
}

// GenerateReportStyled renders report with optional color styling
func GenerateReportStyled(report models.Report, colorEnabled bool) string {
	style := newStyler(colorEnabled)
	var sb strings.Builder

	// Header
	sb.WriteString(generateHeaderStyled(report, style))
	sb.WriteString("\n")

	// Evidence Collection
	sb.WriteString(generateEvidenceSectionStyled(report.Evidence, style))
	sb.WriteString("\n")

	// Findings
	sb.WriteString(generateFindingsSectionStyled(report.Findings, style))
	sb.WriteString("\n")

	// Health Assessment
	sb.WriteString(generateHealthSectionStyled(report.HealthScore, style))
	sb.WriteString("\n")

	// Investigator Notes
	if len(report.Evidence.InvestigatorNotes) > 0 {
		sb.WriteString(generateNotesSectionStyled(report.Evidence.InvestigatorNotes, style))
		sb.WriteString("\n")
	}

	// Report Integrity
	sb.WriteString(generateIntegritySectionStyled(report.ReportHash, report.Context, style))

	return sb.String()
}

func generateHeaderStyled(report models.Report, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.header("◼ FORENSIC CODE INVESTIGATION REPORT\n"))
	sb.WriteString(style.label("Target: %s\n", report.TargetPath))
	sb.WriteString(style.label("Timestamp: %s\n", report.InvestigatedAt.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(style.label("Status: COMPLETED\n"))

	return sb.String()
}

func generateEvidenceSectionStyled(evidence models.Evidence, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.section("◼ EVIDENCE COLLECTION\n\n"))

	// File System Evidence
	sb.WriteString(style.label("▸ FILE SYSTEM ANALYSIS\n"))
	sb.WriteString(style.label("  Total Files: %d\n", evidence.FileSystem.TotalFiles))
	sb.WriteString(style.label("  Total Directories: %d\n", evidence.FileSystem.TotalDirectories))
	sb.WriteString(style.label("  Total Size: %s\n", formatBytes(evidence.FileSystem.TotalSize)))

	sb.WriteString("\n  File Type Distribution:\n")
	// Sort file types by count
	types := make([]string, 0, len(evidence.FileSystem.FileTypes))
	for ext := range evidence.FileSystem.FileTypes {
		types = append(types, ext)
	}
	sort.Slice(types, func(i, j int) bool {
		return evidence.FileSystem.FileTypes[types[i]] > evidence.FileSystem.FileTypes[types[j]]
	})
	for i, ext := range types {
		if i < 10 {
			sb.WriteString(style.label("    %s: %d files\n", ext, evidence.FileSystem.FileTypes[ext]))
		}
	}

	if len(evidence.FileSystem.LargestFiles) > 0 {
		sb.WriteString("\n  Largest Files:\n")
		for i, file := range evidence.FileSystem.LargestFiles {
			if i < 5 {
				sb.WriteString(style.label("    %s (%s)\n", file.Path, formatBytes(file.Size)))
			}
		}
	}

	// Git Evidence
	sb.WriteString("\n")
	sb.WriteString(style.label("▸ GIT REPOSITORY ANALYSIS\n"))
	if evidence.Git.IsRepository {
		sb.WriteString(style.label("  Repository Status: ACTIVE\n"))
		sb.WriteString(style.label("  Total Commits: %d\n", evidence.Git.TotalCommits))
		sb.WriteString(style.label("  Contributors: %d\n", evidence.Git.Contributors))
		if !evidence.Git.FirstCommitDate.IsZero() {
			sb.WriteString(style.label("  First Commit: %s\n", evidence.Git.FirstCommitDate.Format("2006-01-02")))
			sb.WriteString(style.label("  Last Commit: %s\n", evidence.Git.LastCommitDate.Format("2006-01-02")))
			daysSinceLastCommit := int(time.Since(evidence.Git.LastCommitDate).Hours() / 24)
			sb.WriteString(style.label("  Days Since Last Commit: %d\n", daysSinceLastCommit))
		}
	} else {
		sb.WriteString(style.label("  Repository Status: NOT A GIT REPOSITORY\n"))
	}

	// Code Markers
	sb.WriteString("\n")
	sb.WriteString(style.label("▸ CODE MARKER DETECTION\n"))
	if len(evidence.CodeMarkers) > 0 {
		markerCounts := make(map[string]int)
		for _, marker := range evidence.CodeMarkers {
			markerCounts[marker.Type]++
		}
		sb.WriteString(style.label("  Total Markers: %d\n", len(evidence.CodeMarkers)))
		for markerType, count := range markerCounts {
			sb.WriteString(style.label("    %s: %d\n", markerType, count))
		}
	} else {
		sb.WriteString(style.label("  No code markers detected\n"))
	}

	return sb.String()
}

func generateFindingsSectionStyled(findings []models.Finding, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.section("◼ FINDINGS\n\n"))

	if len(findings) == 0 {
		sb.WriteString(style.info("No significant issues detected.\n"))
		return sb.String()
	}

	// Sort findings by severity (highest first)
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].Severity > findings[j].Severity
	})

	for i, finding := range findings {
		label := severityTag(finding.Severity, style)
		sb.WriteString(fmt.Sprintf("%s %s\n", label, finding.Title))
		sb.WriteString(style.label("  %s\n", finding.Description))
		if len(finding.Evidence) > 0 {
			sb.WriteString(style.dim("  Evidence:\n"))
			for _, evidence := range finding.Evidence {
				sb.WriteString(style.label("    - %s\n", evidence))
			}
		}
		if i < len(findings)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func generateHealthSectionStyled(healthScore int, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.section("◼ HEALTH ASSESSMENT\n\n"))
	sb.WriteString(style.label("Overall Health Score: %d/100\n\n", healthScore))

	var assessment string
	switch {
	case healthScore >= 90:
		assessment = "EXCELLENT - Project demonstrates strong health indicators"
	case healthScore >= 75:
		assessment = "GOOD - Minor issues present, generally well-maintained"
	case healthScore >= 60:
		assessment = "FAIR - Several concerns require attention"
	case healthScore >= 40:
		assessment = "POOR - Significant issues impacting project quality"
	default:
		assessment = "CRITICAL - Major problems require immediate intervention"
	}

	sb.WriteString(style.label("Assessment: %s\n", assessment))
	return sb.String()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func generateNotesSectionStyled(notes []string, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.section("◼ INVESTIGATOR NOTES\n\n"))

	for _, note := range notes {
		sb.WriteString(style.info("%s\n", note))
	}

	return sb.String()
}

func generateIntegritySectionStyled(hash string, context string, style styler) string {
	var sb strings.Builder

	sb.WriteString(style.section("◼ REPORT INTEGRITY\n\n"))
	sb.WriteString(style.dim("Report Signature:\n"))
	sb.WriteString(style.dim("SHA256: %s\n", hash))
	if context != "" && context != "default" {
		sb.WriteString(style.label("Context: %s\n", strings.ToUpper(context)))
	}
	sb.WriteString("\n" + style.header("═══════════════════════════════════════════════════════════════\n"))
	sb.WriteString(style.header("                    END OF REPORT\n"))
	sb.WriteString(style.header("═══════════════════════════════════════════════════════════════\n"))

	return sb.String()
}

func severityTag(sev models.Severity, style styler) string {
	switch sev {
	case models.SeverityCritical, models.SeverityHigh:
		return style.high("[ %s ]", sev.String())
	case models.SeverityMedium:
		return style.medium("[ %s ]", sev.String())
	case models.SeverityLow:
		return style.low("[ %s ]", sev.String())
	default:
		return style.info("[ %s ]", sev.String())
	}
}
