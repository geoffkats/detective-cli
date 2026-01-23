package git

import (
	"os"
	"path/filepath"
	"time"

	"github.com/detective-cli/detective/pkg/models"
	"github.com/go-git/go-git/v5"
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

	contributors := make(map[string]bool)
	var recentCommits []models.CommitInfo
	commitCount := 0
	var firstCommit, lastCommit time.Time

	err = commitIter.ForEach(func(c *object.Commit) error {
		commitCount++

		// Track contributors
		contributors[c.Author.Email] = true

		// Track first and last commit dates
		if firstCommit.IsZero() || c.Author.When.Before(firstCommit) {
			firstCommit = c.Author.When
		}
		if lastCommit.IsZero() || c.Author.When.After(lastCommit) {
			lastCommit = c.Author.When
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

	return evidence, nil
}

// IsGitRepository checks if a path is a git repository
func IsGitRepository(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	return err == nil && info.IsDir()
}
