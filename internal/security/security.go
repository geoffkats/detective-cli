package security

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/detective-cli/detective/pkg/models"
)

// Secret detection patterns
var secretPatterns = map[string]*regexp.Regexp{
	"api-key":        regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]([a-zA-Z0-9_\-]{20,})['"]`),
	"password":       regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*['"]([^'"]{8,})['"]`),
	"token":          regexp.MustCompile(`(?i)(token|auth[_-]?token)\s*[:=]\s*['"]([a-zA-Z0-9_\-]{20,})['"]`),
	"private-key":    regexp.MustCompile(`-----BEGIN (RSA |EC |DSA )?PRIVATE KEY-----`),
	"aws-access-key": regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key[_-]?id|aws_access_key)\s*[:=]\s*['"]?(AKIA[0-9A-Z]{16})['"]?`),
	"github-token":   regexp.MustCompile(`(?i)(github[_-]?token|gh_token)\s*[:=]\s*['"]?(ghp_[a-zA-Z0-9]{36})['"]?`),
	"slack-token":    regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[a-zA-Z0-9-]{24,}`),
	"stripe-key":     regexp.MustCompile(`(?i)(sk_live_|pk_live_)[a-zA-Z0-9]{24,}`),
	"database-url":   regexp.MustCompile(`(?i)(database[_-]?url|db[_-]?url)\s*[:=]\s*['"]?(postgres|mysql|mongodb)://[^'"\\s]+['"]?`),
}

// SQL injection patterns
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)query\s*\(?\s*["']?\s*SELECT.*\+.*["']?`),                        // String concatenation in SQL
	regexp.MustCompile(`(?i)execQuery\s*\(?\s*["']?\s*SELECT.*\$.*["']?`),                    // Variable interpolation
	regexp.MustCompile(`(?i)(SELECT|UPDATE|DELETE|INSERT).*\+\s*(req\.body|params|request)`), // Direct user input
	regexp.MustCompile(`(?i)execute\s*\(\s*["']SELECT.*\+`),
}

// XSS vulnerability patterns
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)innerHTML\s*=\s*.*\+`),                   // Direct innerHTML manipulation
	regexp.MustCompile(`(?i)document\.write\s*\(\s*.*\+`),            // document.write with concatenation
	regexp.MustCompile(`(?i)eval\s*\(\s*(req\.|request\.|params\.)`), // eval with user input
	regexp.MustCompile(`(?i)dangerouslySetInnerHTML`),                // React unsafe pattern
}

// ScanSecurity performs security analysis on the codebase
func ScanSecurity(rootPath string, excludeDirs []string) (models.SecurityEvidence, error) {
	evidence := models.SecurityEvidence{
		HardcodedSecrets:  []models.SecretFinding{},
		SQLInjectionRisks: []models.SecurityRisk{},
		XSSRisks:          []models.SecurityRisk{},
		InsecurePatterns:  []models.SecurityRisk{},
	}

	sourceExtensions := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true, ".php": true,
		".java": true, ".rb": true, ".cs": true, ".jsx": true, ".tsx": true,
		".vue": true, ".html": true, ".env": true, ".config": true, ".yml": true,
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if info.Name() == exclude || strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only scan source and config files
		ext := filepath.Ext(path)
		if !sourceExtensions[ext] && !strings.HasSuffix(path, ".env") {
			return nil
		}

		// Scan file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check for hardcoded secrets
			for secretType, pattern := range secretPatterns {
				if pattern.MatchString(line) {
					evidence.HardcodedSecrets = append(evidence.HardcodedSecrets, models.SecretFinding{
						File:    path,
						Line:    lineNum,
						Type:    secretType,
						Pattern: strings.TrimSpace(line),
					})
				}
			}

			// Check for SQL injection risks
			for _, pattern := range sqlInjectionPatterns {
				if pattern.MatchString(line) {
					evidence.SQLInjectionRisks = append(evidence.SQLInjectionRisks, models.SecurityRisk{
						File:        path,
						Line:        lineNum,
						Type:        "sql-injection",
						Description: "Potential SQL injection vulnerability from string concatenation",
						Severity:    models.SeverityHigh,
					})
				}
			}

			// Check for XSS risks
			for _, pattern := range xssPatterns {
				if pattern.MatchString(line) {
					evidence.XSSRisks = append(evidence.XSSRisks, models.SecurityRisk{
						File:        path,
						Line:        lineNum,
						Type:        "xss",
						Description: "Potential XSS vulnerability from unsafe HTML rendering",
						Severity:    models.SeverityMedium,
					})
				}
			}

			// Check for other insecure patterns
			if strings.Contains(line, "crypto.MD5") || strings.Contains(line, "hashlib.md5") {
				evidence.InsecurePatterns = append(evidence.InsecurePatterns, models.SecurityRisk{
					File:        path,
					Line:        lineNum,
					Type:        "weak-crypto",
					Description: "Weak hashing algorithm (MD5) detected",
					Severity:    models.SeverityMedium,
				})
			}

			if regexp.MustCompile(`(?i)http://`).MatchString(line) && !strings.Contains(line, "localhost") {
				evidence.InsecurePatterns = append(evidence.InsecurePatterns, models.SecurityRisk{
					File:        path,
					Line:        lineNum,
					Type:        "insecure-protocol",
					Description: "Insecure HTTP protocol usage (should use HTTPS)",
					Severity:    models.SeverityLow,
				})
			}
		}

		return nil
	})

	return evidence, err
}
