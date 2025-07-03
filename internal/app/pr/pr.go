package pr

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/kemadev/ci-cd/pkg/ci"
)

func CheckPRTitle(args []string) (ci.Finding, error) {
	if len(args) < 3 {
		return ci.Finding{}, fmt.Errorf("pr title is required")
	}

	prTitle := strings.Join(args[2:], " ")
	// Highly inspired by https://gist.github.com/marcojahn/482410b728c31b221b70ea6d2c433f0c
	// Reference: https://www.conventionalcommits.org
	prTitleRegex := `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test){1}(\([[:alnum:]._-]+\))?(!)?: ([[:alnum:]])+([[:space:][:print:]])*$`

	exp, err := regexp.Compile(prTitleRegex)
	if err != nil {
		return ci.Finding{}, fmt.Errorf("failed to compile regex: %w", err)
	}

	if !exp.MatchString(prTitle) {
		return ci.Finding{
			ToolName: "pr-title-checker",
			Level:    "error",
			RuleID:   "pr-title-conventional-commit",
			FilePath: "pr-title",
			Message:  "PR title does not follow conventional commit format - See https://www.conventionalcommits.org",
		}, nil
	}

	slog.Info("PR title is valid", "title", prTitle)

	return ci.Finding{}, nil
}
