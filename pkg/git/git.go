package git

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/caarlos0/svu/pkg/svu"
	"github.com/go-git/go-git/v5"
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

func TagSemver() error {
	nextVersion, err := svu.Next()
	if err != nil {
		return fmt.Errorf("error getting next version: %w", err)
	}

	slog.Debug("next version", slog.String("version", nextVersion))

	repo, err := GetGitRepo()
	if err != nil {
		return fmt.Errorf("error getting git repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD reference: %w", err)
	}

	ref, err := repo.CreateTag(nextVersion, head.Hash(), nil)
	if err != nil {
		return fmt.Errorf("error creating tag: %w", err)
	}

	slog.Info("tag created", slog.String("tag", ref.Name().Short()))

	repo.Push(&git.PushOptions{
		RemoteName: "origin",
		FollowTags: true,
	})

	slog.Debug("pushed tag", slog.String("tag", ref.Name().Short()))

	return nil
}
