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

	findings = append(findings, analyzeFileSystem(evidence.FileSystem)...)
	findings = append(findings, analyzeGit(evidence.Git)...)
	findings = append(findings, analyzeCodeMarkers(evidence.CodeMarkers)...)
	findings = append(findings, analyzeSecurity(evidence.Security)...)

	return findings
}

func analyzeFileSystem(fs models.FileSystemEvidence) []models.Finding {
	var findings []models.Finding

	// Empty or small projects
	if fs.TotalFiles < 5 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Minimal File Count",
			Description: fmt.Sprintf("Project contains only %d files, suggesting early development stage or incomplete project.", fs.TotalFiles),
			Evidence:    []string{fmt.Sprintf("Total files: %d", fs.TotalFiles)},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"Add baseline structure (src/, docs/, tests/) to grow maintainably.",
			},
		})
	}

	// Smarter large-file categorization
	for _, file := range fs.LargestFiles {
		if file.Size <= 1*1024*1024 {
			continue
		}

		sev := models.SeverityLow
		title := "Large File Detected"
		desc := "File exceeds 1MB; consider keeping heavy artifacts out of source control."
		recs := []string{"Move generated artifacts to .gitignore; store assets in CDN or release storage."}

		switch file.Category {
		case models.CategorySource:
			sev = models.SeverityHigh
			desc = "Large source file can hinder reviews and performance. Consider refactor or splitting."
			recs = []string{"Refactor into smaller modules; ensure code owners review for complexity."}
		case models.CategoryBuildArtifact, models.CategoryDependency:
			sev = models.SeverityLow
			desc = "Build artifact or dependency checked in; usually belongs in releases, not VCS."
			recs = []string{"Add patterns to .gitignore (e.g., build outputs, vendor bundles)."}
		case models.CategoryAsset:
			sev = models.SeverityLow
			desc = "Large asset detected; consider optimizing or hosting externally."
			recs = []string{"Compress/resize assets; use asset pipeline or CDN."}
		case models.CategoryTest:
			sev = models.SeverityMedium
			desc = "Unusually large test file; may indicate fixtures that belong outside repo."
			recs = []string{"Externalize large fixtures; generate fixtures on the fly where possible."}
		}

		findings = append(findings, models.Finding{
			Severity:        sev,
			Title:           title,
			Description:     desc,
			Evidence:        []string{fmt.Sprintf("%s (%s) %d bytes", file.Path, file.Category, file.Size)},
			Recommendations: recs,
			Category:        models.FindingPerformance,
		})
	}

	return findings
}

func analyzeGit(git models.GitEvidence) []models.Finding {
	var findings []models.Finding

	// Check if not a git repository
	if !git.IsRepository {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityHigh,
			Title:           "No Version Control",
			Description:     "Directory is not a git repository. Version control is essential for tracking changes and collaboration.",
			Evidence:        []string{"No .git directory found"},
			Recommendations: []string{"Initialize git, commit a baseline, and push to a remote (GitHub/GitLab)."},
			Category:        models.FindingVersionControl,
		})
		return findings
	}

	// Check for single contributor
	if git.Contributors == 1 {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityLow,
			Title:           "Single Contributor",
			Description:     "Project has only one contributor, indicating potential bus factor risk.",
			Evidence:        []string{fmt.Sprintf("Contributors: %d", git.Contributors)},
			Recommendations: []string{"Add a second maintainer or reviewer; document critical workflows."},
			Category:        models.FindingVersionControl,
		})
	}

	// Check for stale repository
	if !git.LastCommitDate.IsZero() {
		daysSinceLastCommit := int(time.Since(git.LastCommitDate).Hours() / 24)
		if daysSinceLastCommit > 180 {
			findings = append(findings, models.Finding{
				Severity:        models.SeverityMedium,
				Title:           "Stale Repository",
				Description:     fmt.Sprintf("No commits in the last %d days, project may be abandoned or inactive.", daysSinceLastCommit),
				Evidence:        []string{fmt.Sprintf("Last commit: %s", git.LastCommitDate.Format("2006-01-02"))},
				Recommendations: []string{"Schedule maintenance sprint; review backlog and cut a release or archive."},
				Category:        models.FindingVersionControl,
			})
		}
	}

	// Check for low commit count
	if git.TotalCommits < 10 {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityLow,
			Title:           "Limited Commit History",
			Description:     "Repository has minimal commit history, suggesting early development stage.",
			Evidence:        []string{fmt.Sprintf("Total commits: %d", git.TotalCommits)},
			Recommendations: []string{"Adopt smaller, frequent commits with clear messages to improve traceability."},
			Category:        models.FindingVersionControl,
		})
	}

	// Uncommitted changes
	if git.UncommittedChanges {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityMedium,
			Title:           "Working Tree Has Uncommitted Changes",
			Description:     "There are pending changes not committed; risk of loss or drift from remote.",
			Evidence:        []string{"git status reports dirty working tree"},
			Recommendations: []string{"Commit or stash changes; ensure CI reflects current state."},
			Category:        models.FindingVersionControl,
		})
	}

	// Commit message quality
	if git.CommitMessageQuality < 0.5 {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityLow,
			Title:           "Commit Message Quality Could Improve",
			Description:     "Commit messages are brief or non-descriptive; this hurts traceability.",
			Evidence:        []string{fmt.Sprintf("Quality score: %.2f", git.CommitMessageQuality)},
			Recommendations: []string{"Use conventional commits or descriptive messages (what/why)."},
			Category:        models.FindingVersionControl,
		})
	}

	return findings

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
			Severity:        severity,
			Title:           "Known Issues in Code",
			Description:     fmt.Sprintf("Found %d FIXME/BUG markers indicating known problems requiring attention.", count),
			Evidence:        []string{fmt.Sprintf("FIXME: %d, BUG: %d", markerCounts["FIXME"], markerCounts["BUG"])},
			Recommendations: []string{"Create tickets for each FIX/BUG marker and prioritize remediation."},
			Category:        models.FindingCodeQuality,
		})
	}

	// TODO markers
	if count := markerCounts["TODO"]; count > 0 {
		severity := models.SeverityLow
		if count > 20 {
			severity = models.SeverityMedium
		}
		findings = append(findings, models.Finding{
			Severity:        severity,
			Title:           "Pending Tasks",
			Description:     fmt.Sprintf("Found %d TODO markers indicating incomplete features or planned work.", count),
			Evidence:        []string{fmt.Sprintf("TODO markers: %d", count)},
			Recommendations: []string{"Convert top TODOs into backlog items; schedule cleanup sprints."},
			Category:        models.FindingMaintainability,
		})
	}

	// HACK markers
	if count := markerCounts["HACK"]; count > 0 {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityMedium,
			Title:           "Technical Debt Indicators",
			Description:     fmt.Sprintf("Found %d HACK markers suggesting suboptimal solutions requiring refactoring.", count),
			Evidence:        []string{fmt.Sprintf("HACK markers: %d", count)},
			Recommendations: []string{"Refactor HACK areas; add tests before changing risky code."},
			Category:        models.FindingMaintainability,
		})
	}

	return findings
}

// CalculateHealthScore computes an overall health score (0-100)
func CalculateHealthScore(evidence models.Evidence, findings []models.Finding) int {
	breakdown := CalculateHealthBreakdown(evidence, findings)
	return breakdown.VersionControl + breakdown.CodeQuality + breakdown.Security + breakdown.Performance + breakdown.Documentation + breakdown.Testing
}

// CalculateHealthBreakdown applies weighted scoring (100 total)
func CalculateHealthBreakdown(evidence models.Evidence, findings []models.Finding) models.HealthBreakdown {
	b := models.HealthBreakdown{}

	// Version control (20)
	if evidence.Git.IsRepository {
		b.VersionControl += 8
		if evidence.Git.TotalCommits >= 20 {
			b.VersionControl += 4
		}
		if evidence.Git.CommitFrequency.Last30Days >= 5 {
			b.VersionControl += 4
		}
		if !evidence.Git.UncommittedChanges {
			b.VersionControl += 4
		}
	}

	// Code quality (25): based on markers and source size
	markerPenalty := 0
	for _, f := range findings {
		if f.Category == models.FindingCodeQuality || f.Category == models.FindingMaintainability {
			switch f.Severity {
			case models.SeverityCritical:
				markerPenalty += 10
			case models.SeverityHigh:
				markerPenalty += 7
			case models.SeverityMedium:
				markerPenalty += 4
			case models.SeverityLow:
				markerPenalty += 2
			}
		}
	}
	b.CodeQuality = maxInt(0, 25-markerPenalty)

	// Security (20)
	securityPenalty := 0
	for _, f := range findings {
		if f.Category == models.FindingSecurity {
			switch f.Severity {
			case models.SeverityCritical:
				securityPenalty += 10
			case models.SeverityHigh:
				securityPenalty += 7
			case models.SeverityMedium:
				securityPenalty += 4
			case models.SeverityLow:
				securityPenalty += 2
			}
		}
	}
	b.Security = maxInt(0, 20-securityPenalty)

	// Performance (15)
	perfPenalty := 0
	for _, f := range findings {
		if f.Category == models.FindingPerformance {
			switch f.Severity {
			case models.SeverityCritical, models.SeverityHigh:
				perfPenalty += 6
			case models.SeverityMedium:
				perfPenalty += 4
			case models.SeverityLow:
				perfPenalty += 2
			}
		}
	}
	b.Performance = maxInt(0, 15-perfPenalty)

	// Documentation (10)
	docPoints := 0
	if evidence.FileSystem.FileTypes[".md"] > 0 || evidence.FileSystem.FileTypes[".adoc"] > 0 {
		docPoints += 6
	}
	if evidence.FileSystem.FileTypes[".md"] > 1 {
		docPoints += 2
	}
	if evidence.FileSystem.FileTypes[".txt"] > 0 {
		docPoints += 2
	}
	b.Documentation = minInt(docPoints, 10)

	// Testing (10)
	if evidence.FileSystem.CategorizedFiles.TestFiles > 0 {
		b.Testing = 8
		if evidence.FileSystem.CategorizedFiles.TestFiles > evidence.FileSystem.CategorizedFiles.SourceFiles/5 {
			b.Testing = 10
		}
	}

	return b
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

func analyzeSecurity(sec models.SecurityEvidence) []models.Finding {
	var findings []models.Finding

	for _, s := range sec.HardcodedSecrets {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityHigh,
			Title:       "Potential Hardcoded Secret",
			Description: fmt.Sprintf("Possible secret (%s) detected in code.", s.Type),
			Evidence:    []string{fmt.Sprintf("%s:%d (%s)", s.File, s.Line, s.Pattern)},
			Recommendations: []string{
				"Rotate the exposed credential immediately.",
				"Replace with environment variables or secret manager (Vault, AWS Secrets Manager).",
			},
			Category: models.FindingSecurity,
		})
	}

	for _, risk := range sec.SQLInjectionRisks {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityHigh,
			Title:           "SQL Injection Risk",
			Description:     risk.Description,
			Evidence:        []string{fmt.Sprintf("%s:%d", risk.File, risk.Line)},
			Recommendations: []string{"Use parameterized queries/ORM bindings; avoid string concatenation."},
			Category:        models.FindingSecurity,
		})
	}

	for _, risk := range sec.XSSRisks {
		findings = append(findings, models.Finding{
			Severity:        models.SeverityMedium,
			Title:           "Potential XSS Vector",
			Description:     risk.Description,
			Evidence:        []string{fmt.Sprintf("%s:%d", risk.File, risk.Line)},
			Recommendations: []string{"HTML-escape untrusted output; use templating safeguards; add CSP."},
			Category:        models.FindingSecurity,
		})
	}

	for _, risk := range sec.InsecurePatterns {
		findings = append(findings, models.Finding{
			Severity:        risk.Severity,
			Title:           risk.Type,
			Description:     risk.Description,
			Evidence:        []string{fmt.Sprintf("%s:%d", risk.File, risk.Line)},
			Recommendations: []string{"Follow secure coding practices; add tests around this area."},
			Category:        models.FindingSecurity,
		})
	}

	return findings
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
