// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package pr

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/kemadev/ci-cd/pkg/ci"
)

var (
	ErrPRTitleInvalid = fmt.Errorf("invalid PR title")
	ErrPRTitleNil     = fmt.Errorf("PR title is nil")
)

func CheckPRTitle(title string) (ci.Finding, error) {
	// Highly inspired by https://gist.github.com/marcojahn/482410b728c31b221b70ea6d2c433f0c
	// Reference: https://www.conventionalcommits.org
	prTitleRegex := `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test){1}(\([[:alnum:]._-]+\))?(!)?: ([[:alnum:]])+([[:space:][:print:]])*$`

	exp, err := regexp.Compile(prTitleRegex)
	if err != nil {
		return ci.Finding{}, fmt.Errorf("failed to compile regex: %w", err)
	}

	if !exp.MatchString(title) {
		return ci.Finding{
			ToolName: "pr-title-checker",
			Level:    "error",
			RuleID:   "pr-title-conventional-commit",
			FilePath: "pr-title",
			Message:  "PR title does not follow conventional commit format - See https://www.conventionalcommits.org",
		}, nil
	}

	slog.Info("PR title is valid", slog.String("title", title))

	return ci.Finding{}, nil
}
