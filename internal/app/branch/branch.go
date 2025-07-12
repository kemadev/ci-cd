// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package branch

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/kemadev/ci-cd/pkg/ci"
	"github.com/kemadev/ci-cd/pkg/git"
)

var (
	ErrGitRepoNil      = fmt.Errorf("git repository is nil")
	ErrBranchesNil     = fmt.Errorf("branches are nil")
	ErrCurrBrancheNil  = fmt.Errorf("current branch is nil")
	ErrCommitNil       = fmt.Errorf("commit object is nil")
	ErrRemoteOriginNil = fmt.Errorf("remote is nil")
)

// StaleBranchThreshold is the threshold for a branch to be considered stale.
const DayBeforeStale = 0

type StaleBranch struct {
	Name             string
	LastCommitDate   time.Time
	LastCommitAuthor string
}

func CheckStaleBranches() (ci.Finding, error) {
	repo, err := git.GetGitRepo()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting git repository: %w", err)
	}

	if repo == nil {
		return ci.Finding{}, ErrGitRepoNil
	}

	branches, err := repo.Branches()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting branches: %w", err)
	}

	if branches == nil {
		return ci.Finding{}, ErrBranchesNil
	}

	currentBranch, err := repo.Head()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting current branch: %w", err)
	}

	if currentBranch == nil {
		return ci.Finding{}, ErrCurrBrancheNil
	}

	var staleBranches []StaleBranch

	err = branches.ForEach(func(branch *plumbing.Reference) error {
		slog.Debug("checking branch", slog.String("branch", branch.Name().Short()))
		// Branch which the workflow is running on is not considered stale
		if branch.Name() == currentBranch.Name() {
			return nil
		}

		commit, err := repo.CommitObject(branch.Hash())
		if err != nil {
			return fmt.Errorf(
				"error getting commit object for branch %s: %w",
				branch.Name(),
				err,
			)
		}

		if commit == nil {
			return fmt.Errorf("branch name %s: %w", branch.Name(), ErrCommitNil)
		}

		if commit.Committer.When.Before(time.Now().AddDate(0, 0, -DayBeforeStale)) {
			staleBranches = append(staleBranches, StaleBranch{
				Name:             branch.Name().Short(),
				LastCommitDate:   commit.Committer.When,
				LastCommitAuthor: commit.Committer.Name,
			})
		}

		return nil
	})
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error iterating branches: %w", err)
	}

	if len(staleBranches) > 0 {
		message := "The following branches are stale: "

		for i, branch := range staleBranches {
			if i > 0 {
				message += ", "
			}

			message += fmt.Sprintf(
				"%s (last commit by %s on %s)",
				branch.Name,
				branch.LastCommitAuthor,
				branch.LastCommitDate.Format(time.DateOnly),
			)
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			return ci.Finding{}, fmt.Errorf("error getting remote origin: %w", err)
		}

		if remote == nil {
			return ci.Finding{}, ErrRemoteOriginNil
		}

		repoURLString := remote.Config().URLs[0]

		repoURL, err := url.Parse(repoURLString)
		if err != nil {
			return ci.Finding{}, fmt.Errorf("error parsing remote URL: %w", err)
		}

		message += fmt.Sprintf(
			". Please delete these stale branches. You can view recently deleted branches (and optionally restore them) by navigating to [reposiitory activity](%s)",
			repoURL.String()+"/activity?activity_type=branch_deletion",
		)

		return ci.Finding{
			ToolName: "stale-branch-checker",
			Level:    "error",
			RuleID:   "no-stale-branch",
			FilePath: "stale-branch",
			Message:  message,
		}, nil
	}

	return ci.Finding{}, nil
}
