package models

import "time"

// Evidence represents raw data collected during investigation
type Evidence struct {
	FileSystem        FileSystemEvidence
	Git               GitEvidence
	CodeMarkers       []CodeMarker
	Timeline          TimelineEvidence
	InvestigatorNotes []string
	ProjectType       ProjectType
	Security          SecurityEvidence
}

// FileSystemEvidence contains file system analysis data
type FileSystemEvidence struct {
	TotalFiles       int
	TotalDirectories int
	TotalSize        int64
	FileTypes        map[string]int
	LargestFiles     []FileInfo
	SkippedDirs      []string
	SkippedDirsCount int
	CategorizedFiles CategorizedFiles
}

// FileInfo represents information about a single file
type FileInfo struct {
	Path     string
	Size     int64
	Type     string
	Category FileCategory
}

// FileCategory represents the purpose/type of a file
type FileCategory string

const (
	CategorySource        FileCategory = "source"
	CategoryAsset         FileCategory = "asset"
	CategoryDependency    FileCategory = "dependency"
	CategoryBuildArtifact FileCategory = "build-artifact"
	CategoryConfig        FileCategory = "config"
	CategoryDocumentation FileCategory = "documentation"
	CategoryTest          FileCategory = "test"
	CategoryOther         FileCategory = "other"
)

// CategorizedFiles holds files grouped by category
type CategorizedFiles struct {
	SourceFiles        int
	AssetFiles         int
	DependencyFiles    int
	BuildArtifacts     int
	ConfigFiles        int
	DocumentationFiles int
	TestFiles          int
	OtherFiles         int
}

// GitEvidence contains git repository analysis data
type GitEvidence struct {
	IsRepository         bool
	TotalCommits         int
	Contributors         int
	FirstCommitDate      time.Time
	LastCommitDate       time.Time
	RecentActivity       []CommitInfo
	CommitFrequency      CommitFrequency
	TopContributors      []ContributorInfo
	UncommittedChanges   bool
	BranchCount          int
	CommitMessageQuality float64 // 0.0-1.0 score
}

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
}

// CommitFrequency represents commit activity patterns
type CommitFrequency struct {
	Last7Days      int
	Last30Days     int
	Last90Days     int
	AveragePerWeek float64
}

// ContributorInfo represents contributor statistics
type ContributorInfo struct {
	Name    string
	Email   string
	Commits int
	Percent float64
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
	Severity        Severity
	Title           string
	Description     string
	Evidence        []string
	Recommendations []string // Actionable next steps
	Category        FindingCategory
}

// FindingCategory represents the type of finding
type FindingCategory string

const (
	FindingCodeQuality     FindingCategory = "code-quality"
	FindingSecurity        FindingCategory = "security"
	FindingPerformance     FindingCategory = "performance"
	FindingMaintainability FindingCategory = "maintainability"
	FindingVersionControl  FindingCategory = "version-control"
	FindingDocumentation   FindingCategory = "documentation"
)

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
	TargetPath      string
	InvestigatedAt  time.Time
	Evidence        Evidence
	Findings        []Finding
	HealthScore     int // 0-100
	HealthBreakdown HealthBreakdown
	ReportHash      string
	Context         string // "student", "enterprise", "default"
}

// HealthBreakdown shows weighted scoring components
type HealthBreakdown struct {
	VersionControl int // 0-20
	CodeQuality    int // 0-25
	Security       int // 0-20
	Performance    int // 0-15
	Documentation  int // 0-10
	Testing        int // 0-10
}

// ProjectType represents detected project framework/type
type ProjectType struct {
	PrimaryLanguage string
	Framework       string
	DetectedFiles   []string // Key files that determined the type
}

// SecurityEvidence contains security analysis findings
type SecurityEvidence struct {
	HardcodedSecrets  []SecretFinding
	SQLInjectionRisks []SecurityRisk
	XSSRisks          []SecurityRisk
	InsecurePatterns  []SecurityRisk
}

// SecretFinding represents a potential hardcoded secret
type SecretFinding struct {
	File    string
	Line    int
	Type    string // "api-key", "password", "token", etc.
	Pattern string
}

// SecurityRisk represents a security vulnerability pattern
type SecurityRisk struct {
	File        string
	Line        int
	Type        string
	Description string
	Severity    Severity
}
