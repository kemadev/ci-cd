package branch

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kemadev/ci-cd/pkg/ci"
	"github.com/kemadev/ci-cd/pkg/git"
)

var (
	ErrGitRepoNil  = fmt.Errorf("git repository is nil")
	ErrBranchesNil = fmt.Errorf("branches are nil")
)

// StaleBranchThreshold is the threshold for a branch to be considered stale.
var DayBeforeStale = 0

type StaleBranch struct {
	Name             string
	LastCommitDate   time.Time
	LastCommitAuthor string
}

func CheckStaleBranches(args []string) (ci.Finding, error) {
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
		return ci.Finding{}, fmt.Errorf("current branch is nil")
	}

	var staleBranches []StaleBranch
	err = branches.ForEach(func(branch *plumbing.Reference) error {
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
			return fmt.Errorf("commit object for branch %s is nil", branch.Name())
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
				branch.LastCommitDate.Format("2006-01-02"),
			)
		}
		remote, err := repo.Remote("origin")
		if err != nil {
			return ci.Finding{}, fmt.Errorf("error getting remote origin: %w", err)
		}
		if remote == nil {
			return ci.Finding{}, fmt.Errorf("remote origin is empty")
		}
		repoUrlString := remote.Config().URLs[0]
		repoUrl, err := url.Parse(repoUrlString)
		if err != nil {
			return ci.Finding{}, fmt.Errorf("error parsing remote URL: %w", err)
		}
		message += fmt.Sprintf(
			". Please delete these stale branches. You can view recently deleted branches (and optionally restore them) by navigating to [reposiitory activity](%s)",
			repoUrl.String()+"/activity?activity_type=branch_deletion",
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
