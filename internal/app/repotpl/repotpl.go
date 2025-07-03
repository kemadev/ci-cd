package repotpl

import "github.com/kemadev/ci-cd/pkg/ci"

var (
	DayBeforeStale = 30
	RepoTemplateUpdateTrackerFile = "config/copier/.copier-answers.yml"
)

func CheckRepoTemplateUpdate(args []string) (ci.Finding, error) {


	return ci.Finding{}, nil
}
