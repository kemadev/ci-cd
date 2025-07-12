// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package repotpl

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/kemadev/ci-cd/pkg/ci"
	kg "github.com/kemadev/ci-cd/pkg/git"
)

var (
	ErrRepoTemplateUpdateTrackerFileDoesNotExist = fmt.Errorf(
		"repo template update tracker file does not exist or is empty",
	)
	ErrGitRepoNil                            = fmt.Errorf("git repository is nil")
	ErrGitHeadNil                            = fmt.Errorf("git repository head is nil")
	ErrGitTagsNil                            = fmt.Errorf("git repository tags is nil")
	ErrGitTagsMalformed                      = fmt.Errorf("git repository tags are malformed")
	ErrRepoTemplateUpdateTrackerFileNoCommit = fmt.Errorf(
		"repo template update tracker file has no commits",
	)
)

const (
	RepoTemplateUpdateTrackerFile = "config/copier/.copier-answers.yml"
	DayBeforeStale                = 30
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

	tplTags, err := tplRepo.Tags()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository tags: %w", err)
	}

	if tplTags == nil {
		return ci.Finding{}, fmt.Errorf("error getting repository tags: %w", ErrGitTagsNil)
	}

	tplLastTag := ""
	semverRegex := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

	for {
		tplTag, err := tplTags.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return ci.Finding{}, fmt.Errorf("error iterating repository tags: %w", err)
		}

		if tplTag == nil {
			break
		}

		// v1.2.3 format
		expectedPartsNumber := 3

		tagName := tplTag.Name().Short()
		if semverRegex.MatchString(tagName) {
			tagParts := strings.Split(tagName, ".")
			if len(tagParts) != expectedPartsNumber {
				return ci.Finding{}, fmt.Errorf(
					"tag %s: expected format vX.Y.Z: %w",
					tagName,
					ErrGitTagsMalformed,
				)
			}

			lastTagParts := strings.Split(tplLastTag, ".")
			if tplLastTag != "" && len(lastTagParts) != expectedPartsNumber {
				return ci.Finding{}, fmt.Errorf(
					"tag %s: expected format vX.Y.Z: %w",
					tagName,
					ErrGitTagsMalformed,
				)
			}

			if tplLastTag == "" || (tagParts[0] > lastTagParts[0] ||
				(tagParts[0] == lastTagParts[0] && tagParts[1] > lastTagParts[1]) ||
				(tagParts[0] == lastTagParts[0] && tagParts[1] == lastTagParts[1] && tagParts[2] > lastTagParts[2])) {
				tplLastTag = tagName
			}
		}
	}

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

	re := regexp.MustCompile(`(?m)^_commit:\s*(.+)$`)

	// The match and exactly one submatch
	expectedMatchesNum := 2
	matches := re.FindStringSubmatch(copierConfContent)

	if len(matches) != expectedMatchesNum {
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

	if lastCommitHash != tplLastTag {
		return ci.Finding{
			ToolName: "repo-template-updater",
			FilePath: RepoTemplateUpdateTrackerFile,
			Level:    "warning",
			RuleID:   "keep-repo-template-updated",
			Message: fmt.Sprintf(
				"New version of repository template is available (%s available, actually got %s). Please update the repository template to ensure you have the latest features and fixes.",
				tplLastTag,
				lastCommitHash,
			),
		}, nil
	}

	return ci.Finding{}, nil
}
