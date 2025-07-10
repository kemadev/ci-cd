package repotpl

import (
	"fmt"
	"regexp"

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

func CheckRepoTemplateUpdate() (ci.Finding, error) {
	tplRepo, err := kg.GetRemoteGitRepo(
		"https://github.com/kemadev/repo-template",
	)
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error opening git repository: %w", err)
	}
	if tplRepo == nil {
		return ci.Finding{}, fmt.Errorf("error opening git repository: %w", ErrGitRepoNil)
	}

	tplHead, err := tplRepo.Head()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository head: %w", err)
	}
	if tplHead == nil {
		return ci.Finding{}, fmt.Errorf("error getting repository head: %w", ErrGitHeadNil)
	}

	tplLastHash := tplHead.Hash().String()
	tplLastHash = tplLastHash[:7]

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

	// Get the commit object
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository commit: %w", err)
	}
	if commit == nil {
		return ci.Finding{}, fmt.Errorf("error getting repository commit: %w", ErrGitHeadNil)
	}

	tree, err := commit.Tree()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository tree: %w", err)
	}
	if tree == nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting repository tree: %w",
			ErrRepoTemplateUpdateTrackerFileDoesNotExist,
		)
	}

	copierConfFile, err := tree.File(RepoTemplateUpdateTrackerFile)
	if err != nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting repository template update tracker file: %w",
			ErrRepoTemplateUpdateTrackerFileDoesNotExist,
		)
	}
	if copierConfFile == nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting repository template update tracker file: %w",
			ErrRepoTemplateUpdateTrackerFileDoesNotExist,
		)
	}

	copierConfContent, err := copierConfFile.Contents()
	if err != nil {
		return ci.Finding{}, fmt.Errorf(
			"error getting repository template update tracker file content: %w",
			ErrRepoTemplateUpdateTrackerFileDoesNotExist,
		)
	}
	if copierConfContent == "" {
		return ci.Finding{}, fmt.Errorf(
			"error getting repository template update tracker file content: %w",
			ErrRepoTemplateUpdateTrackerFileDoesNotExist,
		)
	}

	re := regexp.MustCompile(`(?m)^_commit:\s*([a-fA-F0-9]+)$`)
	matches := re.FindStringSubmatch(copierConfContent)
	if len(matches) != 2 {
		return ci.Finding{}, fmt.Errorf(
			"error parsing repository template update tracker file: %w",
			ErrRepoTemplateUpdateTrackerFileNoCommit,
		)
	}

	lastCommitHash := matches[1]
	if lastCommitHash == "" {
		return ci.Finding{}, fmt.Errorf(
			"error parsing repository template update tracker file: %w",
			ErrRepoTemplateUpdateTrackerFileNoCommit,
		)
	}

	if lastCommitHash != tplLastHash {
		return ci.Finding{
			ToolName: "repo-template-updater",
			FilePath: RepoTemplateUpdateTrackerFile,
			Level:    "warning",
			RuleID:   "keep-repo-template-updated",
			Message: fmt.Sprintf(
				"New version of repository template is available (%s available, actually got %s). Please update the repository template to ensure you have the latest features and fixes.",
				tplLastHash,
				lastCommitHash,
			),
		}, nil
	}

	return ci.Finding{}, nil
}
