package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/detective-cli/detective/pkg/models"
)

// DetectProjectType analyzes the project structure to determine its type
func DetectProjectType(rootPath string) models.ProjectType {
	project := models.ProjectType{
		DetectedFiles: []string{},
	}

	// Key files to check
	keyFiles := map[string]struct {
		language  string
		framework string
	}{
		"composer.json":      {"PHP", "Laravel/PHP"},
		"artisan":            {"PHP", "Laravel"},
		"package.json":       {"JavaScript", "Node.js"},
		"package-lock.json":  {"JavaScript", "Node.js"},
		"yarn.lock":          {"JavaScript", "Node.js"},
		"next.config.js":     {"JavaScript", "Next.js"},
		"nuxt.config.js":     {"JavaScript", "Nuxt.js"},
		"angular.json":       {"TypeScript", "Angular"},
		"tsconfig.json":      {"TypeScript", "TypeScript"},
		"requirements.txt":   {"Python", "Python"},
		"setup.py":           {"Python", "Python"},
		"Pipfile":            {"Python", "Python/Pipenv"},
		"pyproject.toml":     {"Python", "Python/Poetry"},
		"manage.py":          {"Python", "Django"},
		"Gemfile":            {"Ruby", "Ruby/Rails"},
		"go.mod":             {"Go", "Go"},
		"Cargo.toml":         {"Rust", "Rust"},
		"pom.xml":            {"Java", "Maven"},
		"build.gradle":       {"Java", "Gradle"},
		"*.csproj":           {".NET", ".NET"},
		"Dockerfile":         {"", "Docker"},
		"docker-compose.yml": {"", "Docker Compose"},
	}

	detectedTypes := make(map[string]int) // Count occurrences
	var detectedFrameworks []string

	// Walk directory looking for key files (limit depth for performance)
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip if too deep (> 3 levels)
		rel, _ := filepath.Rel(rootPath, path)
		if strings.Count(rel, string(os.PathSeparator)) > 3 {
			return nil
		}

		name := info.Name()
		for keyFile, typeInfo := range keyFiles {
			if name == keyFile || (strings.Contains(keyFile, "*") && strings.HasSuffix(name, strings.TrimPrefix(keyFile, "*"))) {
				if typeInfo.language != "" {
					detectedTypes[typeInfo.language]++
					project.DetectedFiles = append(project.DetectedFiles, name)
				}
				if typeInfo.framework != "" {
					detectedFrameworks = append(detectedFrameworks, typeInfo.framework)
					project.DetectedFiles = append(project.DetectedFiles, name)
				}
			}
		}

		return nil
	})

	// Determine primary language by count
	maxCount := 0
	for lang, count := range detectedTypes {
		if count > maxCount {
			maxCount = count
			project.PrimaryLanguage = lang
		}
	}

	// Determine framework (prioritize more specific frameworks)
	if len(detectedFrameworks) > 0 {
		project.Framework = detectedFrameworks[0] // First detected framework
		// Prioritize Laravel if found
		for _, fw := range detectedFrameworks {
			if strings.Contains(fw, "Laravel") {
				project.Framework = fw
				break
			}
		}
	}

	// Fallback: detect from file extensions if no key files found
	if project.PrimaryLanguage == "" {
		project.PrimaryLanguage = detectFromExtensions(rootPath)
	}

	return project
}

func detectFromExtensions(rootPath string) string {
	extensionCounts := make(map[string]int)
	languageMap := map[string]string{
		".go":   "Go",
		".js":   "JavaScript",
		".ts":   "TypeScript",
		".py":   "Python",
		".php":  "PHP",
		".rb":   "Ruby",
		".java": "Java",
		".cs":   ".NET",
		".rs":   "Rust",
		".cpp":  "C++",
		".c":    "C",
	}

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if lang, ok := languageMap[ext]; ok {
			extensionCounts[lang]++
		}

		return nil
	})

	// Find most common language
	maxCount := 0
	primaryLang := "Unknown"
	for lang, count := range extensionCounts {
		if count > maxCount {
			maxCount = count
			primaryLang = lang
		}
	}

	return primaryLang
}

// GetFrameworkSpecificAdvice returns context-aware advice for detected frameworks
func GetFrameworkSpecificAdvice(framework string) []string {
	advice := make([]string, 0)

	switch {
	case strings.Contains(framework, "Laravel"):
		advice = append(advice, "✓ Laravel project detected - ensure artisan commands are documented")
		advice = append(advice, "✓ Check for N+1 query problems using Laravel Debugbar or Telescope")
		advice = append(advice, "✓ Validate middleware usage for authentication and authorization")
		advice = append(advice, "✓ Ensure database migrations are properly versioned")

	case strings.Contains(framework, "Node.js"):
		advice = append(advice, "✓ Node.js project - check for outdated dependencies with 'npm audit'")
		advice = append(advice, "✓ Consider using .nvmrc to lock Node version")
		advice = append(advice, "✓ Ensure environment variables are documented in .env.example")

	case strings.Contains(framework, "Django"):
		advice = append(advice, "✓ Django project - ensure SECRET_KEY is not hardcoded")
		advice = append(advice, "✓ Check for proper CSRF protection on forms")
		advice = append(advice, "✓ Validate database migration status")

	case strings.Contains(framework, "Docker"):
		advice = append(advice, "✓ Docker detected - ensure .dockerignore is configured")
		advice = append(advice, "✓ Use multi-stage builds to reduce image size")
		advice = append(advice, "✓ Pin base image versions for reproducibility")
	}

	return advice
}
