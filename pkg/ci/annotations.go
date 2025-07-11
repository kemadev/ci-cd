package ci

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrInvalidArg    = fmt.Errorf("invalid argument")
	ErrInvalidFormat = fmt.Errorf("invalid format")
)

func PrintFindings(findings []Finding, format string) error {
	cwd := os.Getenv("PWD")

	var pfindings []*Finding
	for i := range findings {
		pfindings = append(pfindings, &findings[i])
		pfindings[i].FilePath = strings.TrimPrefix(pfindings[i].FilePath, cwd+"/")
	}

	err := validateFindings(pfindings)
	if err != nil {
		return fmt.Errorf("error validating findings: %w", err)
	}

	if len(pfindings) == 0 {
		slog.Info("no finding found")
		return nil
	}

	switch format {
	case "human":
		for _, annotation := range pfindings {
			fmt.Printf("Tool: %s\n", annotation.ToolName)
			fmt.Printf("Rule ID: %s\n", annotation.RuleID)
			fmt.Printf("Level: %s\n", annotation.Level)
			fmt.Printf("File: %s", annotation.FilePath)

			if annotation.StartLine > 0 {
				fmt.Printf(":%d", annotation.StartLine)
			}

			fmt.Printf("\n")
			fmt.Printf("Message: %s\n", annotation.Message)
			fmt.Println()
		}
	case "json":
		output, err := json.MarshalIndent(pfindings, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling findings to JSON: %w", err)
		}

		fmt.Println(string(output))
	case "github":
		for _, annotation := range pfindings {
			githubAnnotation := fmt.Sprintf(
				"::%s title=%s,file=%s",
				annotation.Level,
				annotation.ToolName,
				annotation.FilePath,
			)

			if annotation.StartLine > 0 {
				githubAnnotation += fmt.Sprintf(",line=%d", annotation.StartLine)
			}

			if annotation.EndLine > annotation.StartLine {
				githubAnnotation += fmt.Sprintf(",endLine=%d", annotation.EndLine)
			}

			if annotation.StartCol > 0 {
				githubAnnotation += fmt.Sprintf(",col=%d", annotation.StartCol)
			}

			if annotation.EndCol > annotation.StartCol {
				githubAnnotation += fmt.Sprintf(",endColumn=%d", annotation.EndCol)
			}

			githubAnnotation += "::" + strconv.Quote(annotation.Message)
			fmt.Println(githubAnnotation)
		}
	default:
		return fmt.Errorf("unknown output format %s: %w", format, ErrInvalidFormat)
	}

	return nil
}

func validateFindings(f []*Finding) error {
	for _, annotation := range f {
		if annotation.ToolName == "" {
			return fmt.Errorf("tool name is required for finding %v: %w", annotation, ErrInvalidArg)
		}

		if annotation.RuleID == "" {
			return fmt.Errorf("rule ID is required for finding %v: %w", annotation, ErrInvalidArg)
		}

		// NOTE if changing this, update comment in jsonToFindingsMappings.Level definition
		// Based on GitHub workflow commands, see https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#setting-a-debug-message
		validFindingLevels := []string{"debug", "notice", "warning", "error"}

		annotation.Level = strings.ToLower(annotation.Level)
		switch annotation.Level {
		case "info":
			annotation.Level = "notice"
		case "low":
			annotation.Level = "notice"
		case "medium":
			annotation.Level = "warning"
		case "critical":
			annotation.Level = "error"
		case "high":
			annotation.Level = "error"
		}

		if !slices.Contains(validFindingLevels, annotation.Level) {
			return fmt.Errorf(
				"invalid level %s for finding %v: %w",
				annotation.Level,
				annotation,
				ErrInvalidArg,
			)
		}

		if annotation.FilePath == "" {
			return fmt.Errorf("file path is required for finding %v: %w", annotation, ErrInvalidArg)
		}

		if annotation.Message == "" {
			return fmt.Errorf("message is required for finding %v: %w", annotation, ErrInvalidArg)
		}
	}

	return nil
}
