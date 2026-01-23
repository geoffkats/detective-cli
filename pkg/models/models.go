package models

import "time"

// Evidence represents raw data collected during investigation
type Evidence struct {
	FileSystem        FileSystemEvidence
	Git               GitEvidence
	CodeMarkers       []CodeMarker
	Timeline          TimelineEvidence
	InvestigatorNotes []string
}

// FileSystemEvidence contains file system analysis data
type FileSystemEvidence struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	FileTypes        map[string]int
	LargestFiles     []FileInfo
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path string
	Size int64
	Type string
}

// GitEvidence contains git repository analysis data
type GitEvidence struct {
	IsRepository    bool
	TotalCommits    int
	Contributors    int
	FirstCommitDate time.Time
	LastCommitDate  time.Time
	RecentActivity  []CommitInfo
}

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}

// TimelineEvidence represents file activity timeline analysis
type TimelineEvidence struct {
	OldestFile    time.Time
	NewestFile    time.Time
	MostRecentDay int  // Days since most recent file modification
	ActivityBurst bool // True if files modified in narrow time window
	BurstDaySpan  int  // Days spanned by activity burst
}

// CodeMarker represents a code maintenance marker (TODO, FIXME, etc.)
type CodeMarker struct {
	Type    string // TODO, FIXME, HACK, BUG, NOTE
	File    string
	Line    int
	Content string
}

// Finding represents an inference made from evidence
type Finding struct {
	Severity    Severity
	Title       string
	Description string
	Evidence    []string
}

// Severity levels for findings
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Report represents the complete investigation report
type Report struct {
	TargetPath     string
	InvestigatedAt time.Time
	Evidence       Evidence
	Findings       []Finding
	HealthScore    int // 0-100
	ReportHash     string
	Context        string // "student", "enterprise", "default"
}
