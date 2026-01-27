package inference

import (
	"fmt"
	"strings"

	"github.com/detective-cli/detective/pkg/models"
)

// GenerateFindingsEnhanced analyzes evidence and produces comprehensive findings with actionable recommendations
func GenerateFindingsEnhanced(evidence models.Evidence) []models.Finding {
	var findings []models.Finding

	// Analyze file system with smart categorization
	findings = append(findings, analyzeFileSystemEnhanced(evidence.FileSystem)...)

	// Analyze git with enhanced metrics
	findings = append(findings, analyzeGitEnhanced(evidence.Git)...)

	// Analyze code markers with recommendations
	findings = append(findings, analyzeCodeMarkersEnhanced(evidence.CodeMarkers)...)

	// Analyze security evidence
	findings = append(findings, analyzeSecurityEvidence(evidence.Security)...)

	// Project-specific findings
	findings = append(findings, analyzeProjectType(evidence.ProjectType)...)

	return findings
}

func analyzeFileSystemEnhanced(fs models.FileSystemEvidence) []models.Finding {
	var findings []models.Finding

	// Check for empty or very small projects
	if fs.TotalFiles < 5 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Minimal File Count",
			Description: fmt.Sprintf("Project contains only %d files, suggesting early development stage.", fs.TotalFiles),
			Evidence:    []string{fmt.Sprintf("Total files: %d", fs.TotalFiles)},
			Category:    models.FindingCodeQuality,
			Recommendations: []string{
				"This is normal for new projects",
				"Consider adding a README.md to document your project",
				"Add .gitignore to exclude unnecessary files",
			},
		})
	}

	// Smart large file analysis based on category
	sourceCodeIssues := 0
	buildArtifactIssues := 0
	for _, file := range fs.LargestFiles {
		if file.Size > 1024*1024 { // > 1MB
			switch file.Category {
			case models.CategorySource:
				sourceCodeIssues++
			case models.CategoryBuildArtifact:
				buildArtifactIssues++
			case models.CategoryAsset:
				// Expected for assets, skip
				continue
			}
		}
	}

	if sourceCodeIssues > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Large Source Code Files",
			Description: fmt.Sprintf("Found %d source code files exceeding 1MB. Large files are harder to review and maintain.", sourceCodeIssues),
			Evidence:    []string{fmt.Sprintf("%d large source files detected", sourceCodeIssues)},
			Category:    models.FindingCodeQuality,
			Recommendations: []string{
				"Consider breaking large files into smaller, focused modules",
				"Extract reusable components into separate files",
				"Review for potential code duplication",
			},
		})
	}

	if buildArtifactIssues > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Build Artifacts in Repository",
			Description: fmt.Sprintf("Found %d build artifacts (compiled binaries, .exe, .dll). These should typically be excluded from version control.", buildArtifactIssues),
			Evidence:    []string{fmt.Sprintf("%d build artifacts found", buildArtifactIssues)},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"Add build artifacts to .gitignore",
				"Remove existing artifacts with: git rm --cached <file>",
				"Use CI/CD for artifact generation instead",
			},
		})
	}

	// Check for documentation
	if fs.CategorizedFiles.DocumentationFiles == 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Missing Documentation",
			Description: "No documentation files (README.md, etc.) detected. Good documentation is essential for project maintainability.",
			Evidence:    []string{"No .md or .txt files found"},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"Create a README.md with project overview, setup instructions, and usage examples",
				"Add CONTRIBUTING.md for collaboration guidelines",
				"Consider API documentation if building a library",
			},
		})
	}

	// Check for tests
	if fs.CategorizedFiles.TestFiles == 0 && fs.CategorizedFiles.SourceFiles > 10 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "No Test Files Detected",
			Description: "Project has source code but no test files. Tests are crucial for code quality and maintainability.",
			Evidence:    []string{fmt.Sprintf("%d source files, 0 test files", fs.CategorizedFiles.SourceFiles)},
			Category:    models.FindingCodeQuality,
			Recommendations: []string{
				"Start with testing critical business logic",
				"Aim for at least 70% code coverage on core functionality",
				"Set up automated testing in CI/CD pipeline",
				"Use framework-specific testing tools (Jest, pytest, go test, etc.)",
			},
		})
	}

	return findings
}

func analyzeGitEnhanced(git models.GitEvidence) []models.Finding {
	var findings []models.Finding

	// Check if not a git repository
	if !git.IsRepository {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityHigh,
			Title:       "No Version Control",
			Description: "Directory is not a git repository. Version control is essential for tracking changes, collaboration, and rollback capability.",
			Evidence:    []string{"No .git directory found"},
			Category:    models.FindingVersionControl,
			Recommendations: []string{
				"Initialize git: git init",
				"Create .gitignore file for your language/framework",
				"Make initial commit: git add . && git commit -m 'Initial commit'",
				"Consider pushing to GitHub/GitLab for backup and collaboration",
			},
		})
		return findings
	}

	// Check for uncommitted changes
	if git.UncommittedChanges {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Uncommitted Changes Detected",
			Description: "Working directory has uncommitted changes. Regular commits help track progress and enable easy rollback.",
			Evidence:    []string{"Git status shows modified files"},
			Category:    models.FindingVersionControl,
			Recommendations: []string{
				"Review changes with: git status",
				"Commit meaningful chunks: git add <files> && git commit -m 'descriptive message'",
				"Push to remote regularly to back up work",
			},
		})
	}

	// Analyze commit frequency
	if git.CommitFrequency.Last30Days == 0 && git.TotalCommits > 10 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Inactive Repository",
			Description: "No commits in the last 30 days. Regular commits indicate active development and maintenance.",
			Evidence:    []string{fmt.Sprintf("Last commit: %s", git.LastCommitDate.Format("2006-01-02"))},
			Category:    models.FindingVersionControl,
			Recommendations: []string{
				"If project is complete, add documentation noting stable/production status",
				"If inactive, consider archiving the repository",
				"For active projects, commit at least weekly to track progress",
			},
		})
	} else if git.CommitFrequency.Last7Days > 20 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "High Commit Frequency",
			Description: fmt.Sprintf("%d commits in the last 7 days. Very frequent commits may indicate work-in-progress or lack of local testing before committing.", git.CommitFrequency.Last7Days),
			Evidence:    []string{fmt.Sprintf("Average %.1f commits/week", git.CommitFrequency.AveragePerWeek)},
			Category:    models.FindingVersionControl,
			Recommendations: []string{
				"Ensure commits are meaningful and tested before pushing",
				"Consider using feature branches for experimental work",
				"Use 'git commit --amend' to fix recent commits instead of making new ones",
			},
		})
	}

	// Analyze commit message quality
	if git.CommitMessageQuality < 0.5 && git.TotalCommits > 10 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Low Commit Message Quality",
			Description: fmt.Sprintf("%.0f%% of commit messages are unclear or too brief. Good commit messages help track project history.", git.CommitMessageQuality*100),
			Evidence:    []string{"Many commits with messages like 'fix', 'WIP', 'update'"},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"Use format: 'type(scope): description' (e.g., 'feat(auth): add login validation')",
				"Describe WHAT changed and WHY, not HOW",
				"Keep first line under 50 chars, add details in body if needed",
				"Set up commit message templates or conventional commits",
			},
		})
	}

	// Check contributor diversity
	if git.Contributors == 1 && git.TotalCommits > 50 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Single Contributor (Bus Factor Risk)",
			Description: "Project has only one contributor despite substantial work. This creates knowledge concentration risk.",
			Evidence:    []string{fmt.Sprintf("%d commits by 1 person", git.TotalCommits)},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"Document key architecture decisions and design patterns",
				"Consider pair programming or code review practices",
				"Invite collaborators if it's an open-source project",
				"Write comprehensive README and contribution guidelines",
			},
		})
	}

	// Low commit history
	if git.TotalCommits < 10 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Limited Commit History",
			Description: "Repository has minimal commit history, suggesting early development stage.",
			Evidence:    []string{fmt.Sprintf("Total commits: %d", git.TotalCommits)},
			Category:    models.FindingVersionControl,
			Recommendations: []string{
				"This is normal for new projects",
				"Commit frequently to capture incremental progress",
				"Each commit should be a logical, working state",
			},
		})
	}

	return findings
}

func analyzeCodeMarkersEnhanced(markers []models.CodeMarker) []models.Finding {
	var findings []models.Finding

	if len(markers) == 0 {
		return findings
	}

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
			Category:    models.FindingCodeQuality,
			Recommendations: []string{
				fmt.Sprintf("Create GitHub issues for top %d critical markers", min(count, 10)),
				"Prioritize fixing bugs before adding new features",
				"Set up automated TODO tracking with tools like todo-tree",
				"Schedule regular 'technical debt' sprints to address markers",
			},
		})
	}

	// TODO markers
	if count := markerCounts["TODO"]; count > 20 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "High TODO Count",
			Description: fmt.Sprintf("Found %d TODO markers. Many pending tasks may indicate incomplete features or ambitious roadmap.", count),
			Evidence:    []string{fmt.Sprintf("TODO markers: %d", count)},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				fmt.Sprintf("Review and convert %d high-priority TODOs into tracked issues", min(count, 20)),
				"Remove obsolete TODOs that are no longer relevant",
				"Link TODOs to specific issue numbers: // TODO(#123): description",
				"Set deadlines for feature completion",
			},
		})
	} else if count > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityLow,
			Title:       "Pending Tasks",
			Description: fmt.Sprintf("Found %d TODO markers indicating planned work.", count),
			Evidence:    []string{fmt.Sprintf("TODO markers: %d", count)},
			Category:    models.FindingMaintainability,
			Recommendations: []string{
				"This is normal for active projects",
				"Ensure TODOs have clear descriptions and owners",
			},
		})
	}

	// HACK markers
	if count := markerCounts["HACK"]; count > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Technical Debt Indicators",
			Description: fmt.Sprintf("Found %d HACK markers suggesting suboptimal solutions requiring refactoring.", count),
			Evidence:    []string{fmt.Sprintf("HACK markers: %d", count)},
			Category:    models.FindingCodeQuality,
			Recommendations: []string{
				"Schedule refactoring sessions to eliminate hacks",
				"Document why the hack exists and ideal solution",
				"Prioritize removing hacks in critical paths",
				"Consider pair programming to find better solutions",
			},
		})
	}

	return findings
}

func analyzeSecurityEvidence(security models.SecurityEvidence) []models.Finding {
	var findings []models.Finding

	// Hardcoded secrets
	if len(security.HardcodedSecrets) > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityCritical,
			Title:       "Hardcoded Secrets Detected",
			Description: fmt.Sprintf("Found %d potential hardcoded secrets (API keys, passwords, tokens). This is a critical security risk.", len(security.HardcodedSecrets)),
			Evidence:    buildSecretEvidence(security.HardcodedSecrets),
			Category:    models.FindingSecurity,
			Recommendations: []string{
				"IMMEDIATE: Remove secrets from code and rotate compromised credentials",
				"Use environment variables or secret management tools (Vault, AWS Secrets Manager)",
				"Add .env to .gitignore and provide .env.example template",
				"Use git-secrets or pre-commit hooks to prevent future leaks",
				"Review git history and purge secrets if already committed",
			},
		})
	}

	// SQL injection risks
	if len(security.SQLInjectionRisks) > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityHigh,
			Title:       "SQL Injection Vulnerability Patterns",
			Description: fmt.Sprintf("Found %d potential SQL injection vulnerabilities from string concatenation in queries.", len(security.SQLInjectionRisks)),
			Evidence:    buildSecurityRiskEvidence(security.SQLInjectionRisks),
			Category:    models.FindingSecurity,
			Recommendations: []string{
				"CRITICAL: Use parameterized queries or prepared statements",
				"Never concatenate user input directly into SQL",
				"Use ORM frameworks (Eloquent, Sequelize, SQLAlchemy) with built-in protection",
				"Validate and sanitize all user inputs",
				"Run automated security scans with tools like SQLMap",
			},
		})
	}

	// XSS risks
	if len(security.XSSRisks) > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Cross-Site Scripting (XSS) Risks",
			Description: fmt.Sprintf("Found %d potential XSS vulnerabilities from unsafe HTML rendering.", len(security.XSSRisks)),
			Evidence:    buildSecurityRiskEvidence(security.XSSRisks),
			Category:    models.FindingSecurity,
			Recommendations: []string{
				"Escape all user-generated content before rendering",
				"Use framework built-in escaping (React auto-escapes, use {{ }} in templates)",
				"Implement Content Security Policy (CSP) headers",
				"Avoid innerHTML, document.write, and eval() with user data",
				"Use DOMPurify for sanitizing rich text",
			},
		})
	}

	// Other insecure patterns
	if len(security.InsecurePatterns) > 0 {
		findings = append(findings, models.Finding{
			Severity:    models.SeverityMedium,
			Title:       "Insecure Coding Patterns",
			Description: fmt.Sprintf("Found %d security concerns (weak crypto, insecure protocols, etc.).", len(security.InsecurePatterns)),
			Evidence:    buildSecurityRiskEvidence(security.InsecurePatterns),
			Category:    models.FindingSecurity,
			Recommendations: []string{
				"Replace MD5/SHA1 with bcrypt, argon2, or PBKDF2 for passwords",
				"Use HTTPS instead of HTTP for all external communication",
				"Keep dependencies updated to patch known vulnerabilities",
				"Enable security linting in your IDE",
			},
		})
	}

	return findings
}

func analyzeProjectType(projectType models.ProjectType) []models.Finding {
	var findings []models.Finding

	if projectType.Framework == "" {
		return findings
	}

	// Generic project type detection finding
	findings = append(findings, models.Finding{
		Severity:        models.SeverityLow,
		Title:           fmt.Sprintf("%s Project Detected", projectType.Framework),
		Description:     fmt.Sprintf("Detected %s project. Framework-specific best practices will be applied.", projectType.Framework),
		Evidence:        projectType.DetectedFiles,
		Category:        models.FindingCodeQuality,
		Recommendations: getFrameworkRecommendations(projectType.Framework),
	})

	return findings
}

func getFrameworkRecommendations(framework string) []string {
	switch {
	case strings.Contains(framework, "Laravel"):
		return []string{
			"Ensure .env is in .gitignore and use .env.example as template",
			"Run 'php artisan optimize' before deploying to production",
			"Use Laravel Telescope or Debugbar to detect N+1 queries",
			"Validate all requests with Form Requests",
			"Keep Laravel updated: composer update",
		}
	case strings.Contains(framework, "Node.js"):
		return []string{
			"Run 'npm audit' regularly to check for vulnerable dependencies",
			"Use .nvmrc file to lock Node version",
			"Add 'engines' field in package.json to specify compatible versions",
			"Consider using Typescript for better type safety",
			"Use ESLint and Prettier for code quality",
		}
	case strings.Contains(framework, "Django"):
		return []string{
			"Never commit SECRET_KEY - use environment variables",
			"Enable DEBUG=False in production",
			"Run 'python manage.py check --deploy' before deployment",
			"Use Django's built-in CSRF protection",
			"Keep dependencies updated with pip-audit",
		}
	case strings.Contains(framework, "Docker"):
		return []string{
			"Create .dockerignore to exclude unnecessary files",
			"Use multi-stage builds to minimize image size",
			"Pin base image versions (e.g., node:18.16-alpine, not node:latest)",
			"Run containers as non-root user for security",
			"Scan images with 'docker scan' or Trivy",
		}
	default:
		return []string{
			"Follow framework-specific security best practices",
			"Keep dependencies updated",
			"Use framework's built-in security features",
		}
	}
}

func buildSecretEvidence(secrets []models.SecretFinding) []string {
	evidence := []string{}
	for i, secret := range secrets {
		if i >= 5 { // Limit to first 5
			evidence = append(evidence, fmt.Sprintf("... and %d more", len(secrets)-5))
			break
		}
		evidence = append(evidence, fmt.Sprintf("%s:%d - %s", secret.File, secret.Line, secret.Type))
	}
	return evidence
}

func buildSecurityRiskEvidence(risks []models.SecurityRisk) []string {
	evidence := []string{}
	for i, risk := range risks {
		if i >= 5 {
			evidence = append(evidence, fmt.Sprintf("... and %d more", len(risks)-5))
			break
		}
		evidence = append(evidence, fmt.Sprintf("%s:%d - %s", risk.File, risk.Line, risk.Description))
	}
	return evidence
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CalculateHealthScoreWeighted computes health score with weighted categories
func CalculateHealthScoreWeighted(findings []models.Finding, evidence models.Evidence) (int, models.HealthBreakdown) {
	breakdown := models.HealthBreakdown{
		VersionControl: 20,
		CodeQuality:    25,
		Security:       20,
		Performance:    15,
		Documentation:  10,
		Testing:        10,
	}

	// Deduct points by category
	for _, finding := range findings {
		deduction := 0
		switch finding.Severity {
		case models.SeverityCritical:
			deduction = 10
		case models.SeverityHigh:
			deduction = 5
		case models.SeverityMedium:
			deduction = 3
		case models.SeverityLow:
			deduction = 1
		}

		switch finding.Category {
		case models.FindingVersionControl:
			breakdown.VersionControl = max(0, breakdown.VersionControl-deduction)
		case models.FindingCodeQuality, models.FindingMaintainability:
			breakdown.CodeQuality = max(0, breakdown.CodeQuality-deduction)
		case models.FindingSecurity:
			breakdown.Security = max(0, breakdown.Security-deduction)
		case models.FindingPerformance:
			breakdown.Performance = max(0, breakdown.Performance-deduction)
		default:
			breakdown.CodeQuality = max(0, breakdown.CodeQuality-deduction)
		}
	}

	// Bonus points for good practices
	if evidence.Git.IsRepository && evidence.Git.TotalCommits > 10 {
		breakdown.VersionControl = min(20, breakdown.VersionControl+2)
	}
	if evidence.FileSystem.CategorizedFiles.TestFiles > 0 {
		breakdown.Testing = min(10, breakdown.Testing+5)
	}
	if evidence.FileSystem.CategorizedFiles.DocumentationFiles > 0 {
		breakdown.Documentation = min(10, breakdown.Documentation+5)
	}

	totalScore := breakdown.VersionControl + breakdown.CodeQuality + breakdown.Security +
		breakdown.Performance + breakdown.Documentation + breakdown.Testing

	return totalScore, breakdown
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
