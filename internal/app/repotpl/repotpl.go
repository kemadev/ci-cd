package repotpl

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/kemadev/ci-cd/pkg/ci"
	kg "github.com/kemadev/ci-cd/pkg/git"
)

var (
	ErrRepoTemplateUpdateTrackerFileDoesNotExist = fmt.Errorf(
		"repo template update tracker file does not exist or is empty",
	)
	ErrGitRepoNil                            = fmt.Errorf("git repository is nil")
	ErrGitHeadNil                            = fmt.Errorf("git repository head is nil")
	ErrRepoTemplateUpdateTrackerFileNoCommit = fmt.Errorf(
		"repo template update tracker file has no commits",
	)
)

var (
	DayBeforeStale                = 30
	RepoTemplateUpdateTrackerFile = "config/copier/.copier-answers.yml"
)

func CheckRepoTemplateUpdate(args []string) (ci.Finding, error) {
	repo, err := kg.GetGitRepo()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting git repository: %w", err)
	}
	if repo == nil {
		return ci.Finding{}, ErrGitRepoNil
	}

	head, err := repo.Head()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository head: %w", err)
	}
	if head == nil {
		return ci.Finding{}, fmt.Errorf("error getting repository head: %w", ErrGitHeadNil)
	}

	info, err := repo.Log(&git.LogOptions{
		From:     head.Hash(),
		FileName: &RepoTemplateUpdateTrackerFile,
	})
	if err != nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting log for repository template update tracker file: %w",
			err,
		)
	}

	commit, err := info.Next()
	if err != nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting next commit for repository template update tracker file: %w",
			err,
		)
	}
	if commit == nil {
		return ci.Finding{}, ErrRepoTemplateUpdateTrackerFileNoCommit
	}

	if commit.Committer.When.AddDate(0, 0, DayBeforeStale).Before(commit.Committer.When) {
		message := fmt.Sprintf(
			"The repository template has not been updated in the last %d days (last update on %s). Please update the repository template to ensure you have the latest features and fixes.",
			DayBeforeStale,
			commit.Committer.When.Format("2006-01-02"),
		)
		return ci.Finding{
			ToolName: "repo-template-updater",
			FilePath: RepoTemplateUpdateTrackerFile,
			Level:    "warning",
			RuleID:   "keep-repo-template-updated",
			Message:  message,
		}, nil
	}

	return ci.Finding{}, nil
}
