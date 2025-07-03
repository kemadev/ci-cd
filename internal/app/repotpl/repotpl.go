package repotpl

import (
	"fmt"
	"os"

	"github.com/kemadev/ci-cd/pkg/ci"
)

var ErrRepoTemplateUpdateTrackerFileDoesNotExist = fmt.Errorf(
	"repo template update tracker file does not exist or is empty",
)

var (
	DayBeforeStale                = 30
	RepoTemplateUpdateTrackerFile = "config/copier/.copier-answers.yml"
)

func CheckRepoTemplateUpdate(args []string) (ci.Finding, error) {
	info, err := os.Stat(RepoTemplateUpdateTrackerFile)
	if err != nil {
		return ci.Finding{}, fmt.Errorf("error checking repo template update tracker file: %w", err)
	}

	if info.IsDir() {
		return ci.Finding{}, fmt.Errorf(
			"expected %s to be a file, but it is a directory",
			RepoTemplateUpdateTrackerFile,
		)
	}
	if info.Size() == 0 {
		return ci.Finding{}, ErrRepoTemplateUpdateTrackerFileDoesNotExist
	}

	if info.ModTime().AddDate(0, 0, DayBeforeStale).Before(info.ModTime()) {
		message := fmt.Sprintf(
			"The repository template has not been updated in the last %d days (last update on %s). Please update the repository template to ensure you have the latest features and fixes.",
			DayBeforeStale,
			info.ModTime().Format("2006-01-02"),
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
