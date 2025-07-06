package git

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/svu/pkg/svu"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
)

var ErrRemoteURLNotFound = fmt.Errorf("remote URL not found")

func GetGitRepo() (*git.Repository, error) {
	repo, err := git.PlainOpenWithOptions(
		".",
		&git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: false},
	)
	if err != nil {
		return nil, fmt.Errorf("error opening git repository: %w", err)
	}

	return repo, nil
}

func GetGitBasePath() (string, error) {
	repo, err := GetGitRepo()
	if err != nil {
		return "", fmt.Errorf("error getting git repository: %w", err)
	}
	return GetGitBasePathWithRepo(repo)
}

func GetGitBasePathWithRepo(repo *git.Repository) (string, error) {
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("error getting remote: %w", err)
	}

	if len(remote.Config().URLs) == 0 {
		return "", ErrRemoteURLNotFound
	}

	basePath := strings.TrimPrefix(remote.Config().URLs[0], "git@")
	basePath = strings.TrimPrefix(basePath, "https://")
	basePath = strings.TrimSuffix(basePath, ".git")

	return basePath, nil
}

func TagSemver() (bool, error) {
	currentVersion, err := svu.Current()
	if err != nil {
		return false, fmt.Errorf("error getting current version: %w", err)
	}

	slog.Debug("got version", slog.String("current-version", currentVersion))

	nextVersion, err := svu.Next()
	if err != nil {
		return false, fmt.Errorf("error getting next version: %w", err)
	}

	slog.Debug("got version", slog.String("next-version", nextVersion))

	if currentVersion == nextVersion {
		return true, nil
	}

	repo, err := GetGitRepo()
	if err != nil {
		return false, fmt.Errorf("error getting git repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return false, fmt.Errorf("error getting HEAD reference: %w", err)
	}

	ref, err := repo.CreateTag(nextVersion, head.Hash(), nil)
	if err != nil {
		return false, fmt.Errorf("error creating tag: %w", err)
	}

	slog.Info("tag created", slog.String("tag", ref.Name().Short()))

	return true, nil
}

func PushTag() error {
	repo, err := GetGitRepo()
	if err != nil {
		return fmt.Errorf("error getting git repository: %w", err)
	}

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		FollowTags: true,
		Auth: &http.BasicAuth{
			Username: "git",
			Password: os.Getenv("GITHUB_TOKEN"),
		},
	})
	if err != nil {
		return fmt.Errorf("error pushing tags: %w", err)
	}

	slog.Debug("pushed tag")

	return nil
}

func GetRemoteGitRepo(remoteURL string) (*git.Repository, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: remoteURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening git repository: %w", err)
	}

	return repo, nil
}
