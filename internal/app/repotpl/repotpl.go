// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package repotpl

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v6/plumbing/storer"
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

func compareTagParts(tag, compareTag string) (string, error) {
	finalTag := tag

	// v1.2.3 format
	expectedPartsNumber := 3

	tagParts := strings.Split(tag, ".")
	if len(tagParts) != expectedPartsNumber {
		return "", fmt.Errorf(
			"tag %s: expected format vX.Y.Z: %w",
			tag,
			ErrGitTagsMalformed,
		)
	}

	compareTagParts := strings.Split(compareTag, ".")
	if tag != "" && len(compareTagParts) != expectedPartsNumber {
		return "", fmt.Errorf(
			"tag %s: expected format vX.Y.Z: %w",
			tag,
			ErrGitTagsMalformed,
		)
	}

	if tag == "" {
		return compareTag, nil
	}

	for part := range expectedPartsNumber {
		tagPart, err := strconv.Atoi(tagParts[part])
		if err != nil {
			return "", fmt.Errorf(
				"tag %s: error converting tag part %s to int: %w",
				tag,
				tagParts[part],
				err,
			)
		}

		compareTagPart, err := strconv.Atoi(compareTagParts[part])
		if err != nil {
			return "", fmt.Errorf(
				"tag %s: error converting compare tag part %s to int: %w",
				tag,
				compareTagParts[part],
				err,
			)
		}

		if tagPart == compareTagPart {
			continue
		}

		if tagPart < compareTagPart {
			finalTag = compareTag

			break
		}

		if tagPart > compareTagPart {
			finalTag = tag

			break
		}
	}

	return finalTag, nil
}

func getLastTag(tags storer.ReferenceIter) (string, error) {
	var tag string

	semverRegex := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

	for {
		tplTag, err := tags.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return "", fmt.Errorf("error iterating repository tags: %w", err)
		}

		if tplTag == nil {
			break
		}

		tagName := tplTag.Name().Short()
		if semverRegex.MatchString(tagName) {
			tag, err = compareTagParts(tag, tagName)
			if err != nil {
				return "", fmt.Errorf("error comparing tag parts: %w", err)
			}
		}
	}

	return tag, nil
}

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

	tplLastTag, err := getLastTag(tplTags)
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting last tag: %w", err)
	}

	tree, err := kg.GetGitHeadTree()
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error getting repository tree: %w", err)
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

	lastCommitRef := matches[1]
	if lastCommitRef == "" {
		return ci.Finding{}, fmt.Errorf(
			"error parsing repository template update tracker file: %w",
			ErrRepoTemplateUpdateTrackerFileNoCommit,
		)
	}

	if lastCommitRef != tplLastTag {
		return ci.Finding{
			ToolName: "repo-template-updater",
			FilePath: RepoTemplateUpdateTrackerFile,
			Level:    "warning",
			RuleID:   "keep-repo-template-updated",
			Message: fmt.Sprintf(
				"New version of repository template is available (%s available, actually got %s). Please update the repository template to ensure you have the latest features and fixes.",
				tplLastTag,
				lastCommitRef,
			),
		}, nil
	}

	return ci.Finding{}, nil
}
