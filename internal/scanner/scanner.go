package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/detective-cli/detective/pkg/models"
)

// ScanOptions controls scanning behavior
type ScanOptions struct {
	ExcludeDirs []string        // directory names to skip entirely
	OnlyExts    map[string]bool // optional allowed file extensions (for code marker scan)
	SkipHidden  bool            // skip hidden directories/files (names starting with .)
}

func shouldSkipDir(name string, opts ScanOptions) bool {
	if opts.SkipHidden && strings.HasPrefix(name, ".") {
		return true
	}
	for _, ex := range opts.ExcludeDirs {
		if name == ex {
			return true
		}
	}
	return false
}

var markerPatterns = map[string]*regexp.Regexp{
	"TODO":  regexp.MustCompile(`(?i)//\s*TODO:?\s*(.+)`),
	"FIXME": regexp.MustCompile(`(?i)//\s*FIXME:?\s*(.+)`),
	"HACK":  regexp.MustCompile(`(?i)//\s*HACK:?\s*(.+)`),
	"BUG":   regexp.MustCompile(`(?i)//\s*BUG:?\s*(.+)`),
	"NOTE":  regexp.MustCompile(`(?i)//\s*NOTE:?\s*(.+)`),
}

// ScanFileSystem analyzes the file system at the given path
func ScanFileSystem(rootPath string, opts ScanOptions) (models.FileSystemEvidence, error) {
	evidence := models.FileSystemEvidence{
		FileTypes:    make(map[string]int),
		LargestFiles: []models.FileInfo{},
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip excluded directories
		if info.IsDir() {
			if shouldSkipDir(info.Name(), opts) {
				// record skipped dir
				evidence.SkippedDirs = append(evidence.SkippedDirs, info.Name())
				evidence.SkippedDirsCount++
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			evidence.TotalDirectories++
		} else {
			evidence.TotalFiles++
			evidence.TotalSize += info.Size()

			// Track file types
			ext := filepath.Ext(path)
			if ext == "" {
				ext = "no-extension"
			}
			evidence.FileTypes[ext]++

			// Track large files (top 10)
			fileInfo := models.FileInfo{
				Path: path,
				Size: info.Size(),
				Type: ext,
			}
			evidence.LargestFiles = append(evidence.LargestFiles, fileInfo)
		}

		return nil
	})

	// Sort and keep only top 10 largest files
	if len(evidence.LargestFiles) > 1 {
		// Simple bubble sort for top 10
		for i := 0; i < len(evidence.LargestFiles)-1; i++ {
			for j := 0; j < len(evidence.LargestFiles)-i-1; j++ {
				if evidence.LargestFiles[j].Size < evidence.LargestFiles[j+1].Size {
					evidence.LargestFiles[j], evidence.LargestFiles[j+1] =
						evidence.LargestFiles[j+1], evidence.LargestFiles[j]
				}
			}
		}
		if len(evidence.LargestFiles) > 10 {
			evidence.LargestFiles = evidence.LargestFiles[:10]
		}
	}

	return evidence, err
}

// ScanCodeMarkers searches for code maintenance markers in source files
func ScanCodeMarkers(rootPath string, opts ScanOptions) ([]models.CodeMarker, error) {
	var markers []models.CodeMarker

	codeExtensions := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".py": true, ".java": true,
		".c": true, ".cpp": true, ".h": true, ".rs": true, ".rb": true,
		".php": true, ".cs": true, ".swift": true, ".kt": true,
	}

	// If opts.OnlyExts provided, override default set
	if opts.OnlyExts != nil && len(opts.OnlyExts) > 0 {
		codeExtensions = opts.OnlyExts
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories per options
		if info.IsDir() {
			if shouldSkipDir(info.Name(), opts) {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan code files
		ext := filepath.Ext(path)
		if !codeExtensions[ext] {
			return nil
		}

		// Scan file for markers
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

			for markerType, pattern := range markerPatterns {
				if matches := pattern.FindStringSubmatch(line); matches != nil {
					content := strings.TrimSpace(line)
					markers = append(markers, models.CodeMarker{
						Type:    markerType,
						File:    path,
						Line:    lineNum,
						Content: content,
					})
				}
			}
		}

		return nil
	})

	return markers, err
}

// AnalyzeTimeline analyzes file modification patterns
func AnalyzeTimeline(rootPath string, opts ScanOptions) (models.TimelineEvidence, error) {
	timeline := models.TimelineEvidence{}
	var modTimes []time.Time

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories per options
		if info.IsDir() {
			if shouldSkipDir(info.Name(), opts) {
				return filepath.SkipDir
			}
			return nil
		}

		modTime := info.ModTime()
		modTimes = append(modTimes, modTime)

		// Track oldest and newest files
		if timeline.OldestFile.IsZero() || modTime.Before(timeline.OldestFile) {
			timeline.OldestFile = modTime
		}
		if timeline.NewestFile.IsZero() || modTime.After(timeline.NewestFile) {
			timeline.NewestFile = modTime
		}

		return nil
	})

	if err != nil || len(modTimes) == 0 {
		return timeline, err
	}

	// Sort times to analyze distribution
	sort.Slice(modTimes, func(i, j int) bool {
		return modTimes[i].Before(modTimes[j])
	})

	// Calculate days since most recent modification
	timeline.MostRecentDay = int(time.Since(timeline.NewestFile).Hours() / 24)

	// Detect activity bursts (files modified in narrow time window)
	if len(modTimes) > 1 {
		timeSpan := timeline.NewestFile.Sub(timeline.OldestFile)
		daySpan := int(timeSpan.Hours() / 24)
		timeline.BurstDaySpan = daySpan

		// If 80% of files modified within 7 days, it's a burst
		if daySpan <= 7 && len(modTimes) > 2 {
			timeline.ActivityBurst = true
		}
	}

	return timeline, nil
}
