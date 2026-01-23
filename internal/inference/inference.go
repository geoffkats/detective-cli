package inference

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/detective-cli/detective/pkg/models"
)

// GenerateFindings analyzes evidence and produces findings
func GenerateFindings(evidence models.Evidence) []models.Finding {
	var findings []models.Finding

	// Analyze file system evidence
	findings = append(findings, analyzeFileSystem(evidence.FileSystem)...)

	// Analyze git evidence
	findings = append(findings, analyzeGit(evidence.Git)...)

	// Analyze code markers
	findings = append(findings, analyzeCodeMarkers(evidence.CodeMarkers)...)

	return findings
}

func analyzeFileSystem(fs models.FileSystemEvidence) []models.Finding {
	var findings []models.Finding

	// Check for empty or very small projects
	if fs.TotalFiles < 5 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Minimal File Count",
			Description: fmt.Sprintf("Project contains only %d files, suggesting early development stage or incomplete project.", fs.TotalFiles),
			Evidence:    []string{fmt.Sprintf("Total files: %d", fs.TotalFiles)},
		})
	}

	// Check for large files
	for _, file := range fs.LargestFiles {
		if file.Size > 1024*1024 { // > 1MB
			findings = append(findings, models.Finding{
				Severity:    models.SeverityMedium,
				Title:       "Large File Detected",
				Description: fmt.Sprintf("File exceeds 1MB, may impact repository performance and code review efficiency."),
				Evidence:    []string{fmt.Sprintf("%s: %d bytes", file.Path, file.Size)},
			})
		}
	}

	return findings
}

func analyzeGit(git models.GitEvidence) []models.Finding {
	var findings []models.Finding

	// Check if not a git repository
	if !git.IsRepository {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityHigh,
			Title:       "No Version Control",
			Description: "Directory is not a git repository. Version control is essential for tracking changes and collaboration.",
			Evidence:    []string{"No .git directory found"},
		})
		return findings
	}

	// Check for single contributor
	if git.Contributors == 1 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Single Contributor",
			Description: "Project has only one contributor, indicating potential bus factor risk.",
			Evidence:    []string{fmt.Sprintf("Contributors: %d", git.Contributors)},
		})
	}

	// Check for stale repository
	if !git.LastCommitDate.IsZero() {
		daysSinceLastCommit := int(time.Since(git.LastCommitDate).Hours() / 24)
		if daysSinceLastCommit > 180 {
			findings = append(findings, models.Finding{
				Severity:    models.SeverityMedium,
				Title:       "Stale Repository",
				Description: fmt.Sprintf("No commits in the last %d days, project may be abandoned or inactive.", daysSinceLastCommit),
				Evidence:    []string{fmt.Sprintf("Last commit: %s", git.LastCommitDate.Format("2006-01-02"))},
			})
		}
	}

	// Check for low commit count
	if git.TotalCommits < 10 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Limited Commit History",
			Description: "Repository has minimal commit history, suggesting early development stage.",
			Evidence:    []string{fmt.Sprintf("Total commits: %d", git.TotalCommits)},
		})
	}

	return findings
}

func analyzeCodeMarkers(markers []models.CodeMarker) []models.Finding {
	var findings []models.Finding

	if len(markers) == 0 {
		return findings
	}

	// Count by type
	markerCounts := make(map[string]int)
	for _, marker := range markers {
		markerCounts[marker.Type]++
	}

	// FIXME and BUG are higher severity
	if count := markerCounts["FIXME"] + markerCounts["BUG"]; count > 0 {
		severity := models.SeverityMedium
		if count > 10 {
			severity = models.SeverityHigh
		}
		findings = append(findings, models.Finding{
			Severity:    severity,
			Title:       "Known Issues in Code",
			Description: fmt.Sprintf("Found %d FIXME/BUG markers indicating known problems requiring attention.", count),
			Evidence:    []string{fmt.Sprintf("FIXME: %d, BUG: %d", markerCounts["FIXME"], markerCounts["BUG"])},
		})
	}

	// TODO markers
	if count := markerCounts["TODO"]; count > 0 {
		severity := models.SeverityLow
		if count > 20 {
			severity = models.SeverityMedium
		}
		findings = append(findings, models.Finding{
			Severity:    severity,
			Title:       "Pending Tasks",
			Description: fmt.Sprintf("Found %d TODO markers indicating incomplete features or planned work.", count),
			Evidence:    []string{fmt.Sprintf("TODO markers: %d", count)},
		})
	}

	// HACK markers
	if count := markerCounts["HACK"]; count > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Technical Debt Indicators",
			Description: fmt.Sprintf("Found %d HACK markers suggesting suboptimal solutions requiring refactoring.", count),
			Evidence:    []string{fmt.Sprintf("HACK markers: %d", count)},
		})
	}

	return findings
}

// CalculateHealthScore computes an overall health score (0-100)
func CalculateHealthScore(findings []models.Finding) int {
	score := 100

	for _, finding := range findings {
		switch finding.Severity {
		case models.SeverityCritical:
			score -= 20
		case models.SeverityHigh:
			score -= 10
		case models.SeverityMedium:
			score -= 5
		case models.SeverityLow:
			score -= 2
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// GenerateInvestigatorNotes creates forensic narrative insights
func GenerateInvestigatorNotes(evidence models.Evidence) []string {
	var notes []string

	// Binary presence analysis
	for _, file := range evidence.FileSystem.LargestFiles {
		if strings.HasSuffix(strings.ToLower(file.Path), ".exe") || strings.HasSuffix(strings.ToLower(file.Path), ".dll") {
			notes = append(notes, "⚠️  Compiled binary artifacts detected in project root. Suggests either manual post-build commits or automated build artifact inclusion. Consider adding to .gitignore.")
			break
		}
	}

	// Version control narrative
	if !evidence.Git.IsRepository {
		notes = append(notes, "📋 No version control detected. Limits traceability of changes, authorship attribution, and collaborative development practices.")
	} else if evidence.Git.TotalCommits == 0 {
		notes = append(notes, "📋 Git repository initialized but empty. Project is in pre-development state or commits not yet recorded.")
	}

	// Activity pattern analysis
	if evidence.Timeline.ActivityBurst {
		notes = append(notes, fmt.Sprintf("⏱️  Activity burst detected: %d%% of files modified within %d days. Suggests concentrated development effort.",
			calculateBurstPercentage(evidence), evidence.Timeline.BurstDaySpan))
	} else if evidence.Timeline.MostRecentDay > 180 {
		notes = append(notes, fmt.Sprintf("⏸️  Dormancy detected: No file modifications for %d days. Project may be stable/complete or inactive.",
			evidence.Timeline.MostRecentDay))
	}

	// Code quality narrative
	todoCount := countMarkerType(evidence.CodeMarkers, "TODO")
	fixmeCount := countMarkerType(evidence.CodeMarkers, "FIXME")
	hackCount := countMarkerType(evidence.CodeMarkers, "HACK")

	if todoCount > 20 {
		notes = append(notes, "📝 High TODO marker density suggests ambitious roadmap or incomplete feature set.")
	}
	if fixmeCount+hackCount > 5 {
		notes = append(notes, "🔧 Elevated technical debt indicators. Consider prioritizing refactoring and code quality improvements.")
	}

	// Project maturity signals
	if evidence.FileSystem.TotalFiles > 100 && evidence.Git.TotalCommits > 50 && evidence.Git.Contributors > 2 {
		notes = append(notes, "✨ Maturity indicators present: Substantial codebase with collaborative history suggests established project.")
	}

	return notes
}

// ContextualizeFindings adjusts severity based on context
func ContextualizeFindings(findings []models.Finding, context string) []models.Finding {
	if context == "" {
		context = "default"
	}

	contextualized := make([]models.Finding, len(findings))
	copy(contextualized, findings)

	switch strings.ToLower(context) {
	case "student":
		// More lenient for learning projects
		for i := range contextualized {
			if contextualized[i].Severity == models.SeverityHigh &&
				(strings.Contains(contextualized[i].Title, "Single Contributor") ||
					strings.Contains(contextualized[i].Title, "Version Control")) {
				contextualized[i].Severity = models.SeverityMedium
			}
		}
	case "enterprise":
		// Stricter for production systems
		for i := range contextualized {
			if contextualized[i].Severity == models.SeverityMedium &&
				(strings.Contains(contextualized[i].Title, "No Version Control") ||
					strings.Contains(contextualized[i].Title, "Stale Repository")) {
				contextualized[i].Severity = models.SeverityHigh
			}
		}
	}

	return contextualized
}

func calculateBurstPercentage(evidence models.Evidence) int {
	if evidence.Timeline.BurstDaySpan == 0 {
		return 100
	}
	return 80 // Conservative estimate for activity burst
}

func countMarkerType(markers []models.CodeMarker, markerType string) int {
	count := 0
	for _, m := range markers {
		if m.Type == markerType {
			count++
		}
	}
	return count
}

// ComputeReportHash generates SHA256 hash of report for integrity
func ComputeReportHash(report models.Report) string {
	data := fmt.Sprintf("%s|%d|%d|%s",
		report.TargetPath,
		report.InvestigatedAt.Unix(),
		len(report.Findings),
		report.Evidence.Git.IsRepository)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}
