package git

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/detective-cli/detective/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// AnalyzeRepository analyzes a git repository at the given path
func AnalyzeRepository(rootPath string) (models.GitEvidence, error) {
	evidence := models.GitEvidence{
		IsRepository: false,
	}

	// Check if it's a git repository
	repo, err := git.PlainOpen(rootPath)
	if err != nil {
		// Not a git repository or can't open it
		return evidence, nil
	}

	evidence.IsRepository = true

	// Get commit history
	ref, err := repo.Head()
	if err != nil {
		return evidence, nil
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return evidence, nil
	}

	contributors := make(map[string]*models.ContributorInfo)
	var recentCommits []models.CommitInfo
	commitCount := 0
	var firstCommit, lastCommit time.Time
	var commits7Days, commits30Days, commits90Days int
	now := time.Now()
	totalMessageLength := 0
	goodMessages := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		commitCount++

		// Track contributors
		if contrib, exists := contributors[c.Author.Email]; exists {
			contrib.Commits++
		} else {
			contributors[c.Author.Email] = &models.ContributorInfo{
				Name:    c.Author.Name,
				Email:   c.Author.Email,
				Commits: 1,
			}
		}

		// Track first and last commit dates
		if firstCommit.IsZero() || c.Author.When.Before(firstCommit) {
			firstCommit = c.Author.When
		}
		if lastCommit.IsZero() || c.Author.When.After(lastCommit) {
			lastCommit = c.Author.When
		}

		// Count commits by time period
		daysSince := now.Sub(c.Author.When).Hours() / 24
		if daysSince <= 7 {
			commits7Days++
		}
		if daysSince <= 30 {
			commits30Days++
		}
		if daysSince <= 90 {
			commits90Days++
		}

		// Analyze commit message quality
		msg := strings.TrimSpace(c.Message)
		totalMessageLength += len(msg)
		if len(msg) > 10 && !strings.HasPrefix(msg, "WIP") && !strings.HasPrefix(msg, "fix") {
			goodMessages++
		}

		// Keep recent commits (last 10)
		if len(recentCommits) < 10 {
			recentCommits = append(recentCommits, models.CommitInfo{
				Hash:    c.Hash.String()[:8],
				Author:  c.Author.Name,
				Date:    c.Author.When,
				Message: c.Message,
			})
		}

		return nil
	})

	evidence.TotalCommits = commitCount
	evidence.Contributors = len(contributors)
	evidence.FirstCommitDate = firstCommit
	evidence.LastCommitDate = lastCommit
	evidence.RecentActivity = recentCommits

	// Calculate commit frequency
	if commitCount > 0 && !firstCommit.IsZero() {
		weeksSinceFirstCommit := now.Sub(firstCommit).Hours() / 24 / 7
		if weeksSinceFirstCommit > 0 {
			evidence.CommitFrequency.AveragePerWeek = float64(commitCount) / weeksSinceFirstCommit
		}
	}
	evidence.CommitFrequency.Last7Days = commits7Days
	evidence.CommitFrequency.Last30Days = commits30Days
	evidence.CommitFrequency.Last90Days = commits90Days

	// Calculate commit message quality score
	if commitCount > 0 {
		evidence.CommitMessageQuality = float64(goodMessages) / float64(commitCount)
	}

	// Get top contributors
	var contributorList []models.ContributorInfo
	for _, c := range contributors {
		c.Percent = float64(c.Commits) / float64(commitCount) * 100
		contributorList = append(contributorList, *c)
	}
	// Sort by commits (simple bubble sort for top 5)
	for i := 0; i < len(contributorList)-1 && i < 5; i++ {
		for j := 0; j < len(contributorList)-i-1; j++ {
			if contributorList[j].Commits < contributorList[j+1].Commits {
				contributorList[j], contributorList[j+1] = contributorList[j+1], contributorList[j]
			}
		}
	}
	if len(contributorList) > 5 {
		evidence.TopContributors = contributorList[:5]
	} else {
		evidence.TopContributors = contributorList
	}

	// Check for uncommitted changes
	worktree, err := repo.Worktree()
	if err == nil {
		status, err := worktree.Status()
		if err == nil {
			evidence.UncommittedChanges = !status.IsClean()
		}
	}

	// Count branches
	branches, err := repo.Branches()
	if err == nil {
		branchCount := 0
		branches.ForEach(func(ref *plumbing.Reference) error {
			branchCount++
			return nil
		})
		evidence.BranchCount = branchCount
	}

	return evidence, nil
}

// IsGitRepository checks if a path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	return err == nil && info.IsDir()
}
