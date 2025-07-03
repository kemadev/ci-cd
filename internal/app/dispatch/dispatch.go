//nolint:exhaustruct // we puprosefully avoid initializing all fields as it would be a complete mess
package dispatch

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/kemadev/ci-cd/internal/app/branch"
	"github.com/kemadev/ci-cd/internal/app/config"
	"github.com/kemadev/ci-cd/internal/app/lint"
	"github.com/kemadev/ci-cd/internal/app/pr"
	"github.com/kemadev/ci-cd/internal/app/repotpl"
	"github.com/kemadev/ci-cd/pkg/ci"
	"github.com/kemadev/ci-cd/pkg/filesfind"
	"github.com/kemadev/ci-cd/pkg/git"
)

var (
	ErrUnknownCommand    = fmt.Errorf("unknown command")
	ErrNoCommandProvided = fmt.Errorf("no command provided")
)

//nolint:cyclop,funlen // the enormous switch is (hopefully) easily understandable for a human
func DispatchCommand(config *config.Config, args []string) (int, error) {
	gitRepoBasePath, err := git.GetGitBasePath()
	if err != nil {
		return 1, fmt.Errorf("error getting git base path: %w", err)
	}

	if len(args) == 0 {
		return 0, ErrNoCommandProvided
	}

	goModList, err := filesfind.FindFilesByExtension(filesfind.FilesFindingArgs{
		Extension: "go.mod",
		Recursive: true,
	})
	if err != nil {
		return 1, fmt.Errorf("error finding go.mod files: %w", err)
	}

	slog.Debug("Go mod list", slog.Any("goModList", goModList))

	goRc := 0
	goErr := error(nil)

	switch args[0] {
	case "docker":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "hadolint",
				Ext: "Dockerfile",
				Paths: []string{
					filesfind.FilesFindingRootPath,
				},
				CliArgs: []string{
					"--format",
					"json",
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "hadolint",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "code",
						},
						Level: ci.JsonMappingInfo{
							Key: "level",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "file",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "line",
						},
						Message: ci.JsonMappingInfo{
							Key: "message",
						},
					},
				},
			})

		return rc, err

	case "gha":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "actionlint",
				Ext: ".yaml",
				Paths: []string{
					filesfind.FilesFindingRootPath + "/.github/workflows",
					filesfind.FilesFindingRootPath + "/.github/actions",
				},
				CliArgs: []string{
					"-format",
					"{{json .}}",
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "gha-actionlint",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "kind",
						},
						Level: ci.JsonMappingInfo{
							OverrideValue: "warning",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "filepath",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "line",
						},
						StartCol: ci.JsonMappingInfo{
							Key: "col",
						},
						Message: ci.JsonMappingInfo{
							Key: "message",
						},
					},
				},
			})

		return rc, err

	case "secrets":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "gitleaks",
				CliArgs: []string{
					"git",
					"--no-banner",
					"--max-decode-depth",
					"3",
					"--redact=80",
					"--report-format",
					"json",
					"--gitleaks-ignore-path",
					"config/gitleaks/.gitleaksignore",
					"--report-path",
					"-",
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "secrets-gitleaks",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "RuleID",
						},
						Level: ci.JsonMappingInfo{
							OverrideValue: "error",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "File",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "StartLine",
						},
						EndLine: ci.JsonMappingInfo{
							Key: "EndLine",
						},
						StartCol: ci.JsonMappingInfo{
							Key: "StartColumn",
						},
						EndCol: ci.JsonMappingInfo{
							Key: "EndColumn",
						},
						Message: ci.JsonMappingInfo{
							Key: "Description",
						},
					},
				},
			})

		return rc, err

	case "sast":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "semgrep",
				CliArgs: []string{
					"scan",
					"--metrics=off",
					"--error",
					"--json",
					"--config",
					"p/default",
					"--config",
					"p/gitlab",
					"--config",
					"p/golang",
					"--config",
					"p/cwe-top-25",
					"--config",
					"p/owasp-top-ten",
					"--config",
					"p/r2c-security-audit",
					"--config",
					"p/kubernetes",
					"--config",
					"p/dockerfile",
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						BaseArrayKey: "results",
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "sast-semgrep",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "check_id",
						},
						Level: ci.JsonMappingInfo{
							Key: "extra.severity",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "path",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "start.line",
						},
						EndLine: ci.JsonMappingInfo{
							Key: "end.line",
						},
						StartCol: ci.JsonMappingInfo{
							Key: "start.col",
						},
						EndCol: ci.JsonMappingInfo{
							Key: "end.col",
						},
						Message: ci.JsonMappingInfo{
							Key: "extra.message",
						},
					},
				},
			})

		return rc, err

	case "go-test":
		for _, mod := range goModList {
			retCode, _, _, err := lint.RunLinter(
				config,
				lint.LinterArgs{
					Workdir: strings.Split(mod, "go.mod")[0],
					Bin:     "go",
					CliArgs: []string{
						"test",
						"-bench=.",
						"-benchmem",
						"-json",
						"-race",
						"./...",
					},
					JsonInfo: ci.JsonInfos{
						Type: "stream",
						Mappings: ci.JsonToFindingsMappings{
							ToolName: ci.JsonMappingInfo{
								OverrideValue: "go-test",
							},
							RuleID: ci.JsonMappingInfo{
								OverrideValue: "no-failing-test",
							},
							Level: ci.JsonMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JsonMappingInfo{
								Key: "Package",
								// Get path relative to git repo base path
								ValueTransformerRegex: gitRepoBasePath + "/(.*)",
								Suffix: &ci.JsonMappingInfo{
									// Add a /
									OverrideValue: "/",
									Suffix: &ci.JsonMappingInfo{
										Key: "Output",
										// Name of test file producing the finding
										ValueTransformerRegex: `\s*(\w+_test.go):`,
									},
								},
							},
							StartLine: ci.JsonMappingInfo{
								Key:                   "Output",
								ValueTransformerRegex: `\s*\w_test.go:(\d+):`,
							},
							Message: ci.JsonMappingInfo{
								Key:                   "Output",
								GlobalSelectorRegex:   `\s*(\w_test.go:\d+):`,
								ValueTransformerRegex: `\s*\w_test.go:\d+:\s*(.*)`,
							},
						},
					},
				})
			goRc += retCode

			if err != nil {
				return goRc, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case "go-cover":
		for _, mod := range goModList {
			retCode, _, _, err := lint.RunLinter(
				config,
				lint.LinterArgs{
					Workdir: strings.Split(mod, "go.mod")[0],
					Bin:     "go",
					CliArgs: []string{
						"test",
						"-covermode=atomic",
						"-json",
						"./...",
					},
					FailOnAtLeastOneFinding: true,
					JsonInfo: ci.JsonInfos{
						Type: "stream",
						Mappings: ci.JsonToFindingsMappings{
							ToolName: ci.JsonMappingInfo{
								OverrideValue: "go-cover",
							},
							RuleID: ci.JsonMappingInfo{
								OverrideValue: "no-cover-below-70",
							},
							Level: ci.JsonMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JsonMappingInfo{
								Key: "Package",
								// Get path relative to git repo base path
								ValueTransformerRegex: gitRepoBasePath + "/(.*)",
							},
							Message: ci.JsonMappingInfo{
								Key: "Output",
								// Failed test or coverage lesser than 70%
								GlobalSelectorRegex:   `coverage:\s*([0-6](\d)?(\.\d)?)\% of statements`,
								ValueTransformerRegex: `coverage:\s*([0-6](\d)?(\.\d)?\%) of statements`,
								Suffix: &ci.JsonMappingInfo{
									OverrideValue: " package coverage is below 70%",
								},
							},
						},
					},
				})
			goRc += retCode

			if err != nil {
				return goRc, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case "go-mod-tidy":
		for _, mod := range goModList {
			retCode, _, _, err := lint.RunLinter(
				config,
				lint.LinterArgs{
					Workdir: strings.Split(mod, "go.mod")[0],
					Bin:     "go",
					CliArgs: []string{
						"mod",
						"tidy",
						"-diff",
					},
					JsonInfo: ci.JsonInfos{
						Type: "plain",
						Mappings: ci.JsonToFindingsMappings{
							ToolName: ci.JsonMappingInfo{
								OverrideValue: "go-mod-tidy",
							},
							RuleID: ci.JsonMappingInfo{
								OverrideValue: "no-unused-dependency",
							},
							Level: ci.JsonMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JsonMappingInfo{
								OverrideValue: strings.TrimPrefix(
									strings.Join(
										strings.Split(mod, filesfind.FilesFindingRootPath)[1:],
										"",
									),
									"/",
								),
							},
							Message: ci.JsonMappingInfo{
								OverrideValue: "Unused dependencies found in go.mod",
							},
						},
					},
				})
			goRc += retCode

			if err != nil {
				return goRc, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case "go-mod-name":
		for _, mod := range goModList {
			expectedGoModName := gitRepoBasePath + strings.Split(strings.Join(strings.Split(mod, filesfind.FilesFindingRootPath)[1:], ""), "/go.mod")[0]
			retCode, _, _, err := lint.RunLinter(
				config,
				lint.LinterArgs{
					Workdir: strings.Split(mod, "go.mod")[0],
					Bin:     "go",
					CliArgs: []string{
						"mod",
						"edit",
						"-json",
					},
					JsonInfo: ci.JsonInfos{
						Type: "object",
						Mappings: ci.JsonToFindingsMappings{
							ToolName: ci.JsonMappingInfo{
								OverrideValue: "go-mod-name",
							},
							RuleID: ci.JsonMappingInfo{
								OverrideValue: "mod-name-must-match-repo-structure",
							},
							Level: ci.JsonMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JsonMappingInfo{
								OverrideValue: strings.TrimPrefix(
									strings.Join(
										strings.Split(mod, filesfind.FilesFindingRootPath)[1:],
										"",
									),
									"/",
								),
							},
							Message: ci.JsonMappingInfo{
								Key: "Module.Path",
								GlobalSelectorRegex: strings.ReplaceAll(
									expectedGoModName,
									".",
									`\.`,
								) + "$",
								InvertGlobalSelector: true,
								Suffix: &ci.JsonMappingInfo{
									OverrideValue: " does not match the repository structure, module name should be ",
									Suffix: &ci.JsonMappingInfo{
										OverrideValue: expectedGoModName,
									},
								},
							},
						},
					},
				})

			if retCode != 0 {
				goRc = 1
			}

			if err != nil {
				return goRc, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case "lint":
		var fixEnabled bool
		if len(args) > 1 && args[1] == "--fix" {
			fixEnabled = true
		}

		lintArgs := []string{
			"run",
			"--config",
			"config/golangci-lint/.golangci.yaml",
			"--show-stats=false",
			"--output.json.path",
			"stdout",
		}
		if fixEnabled {
			lintArgs = append(lintArgs, "--fix")
		}

		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin:     "golangci-lint",
				CliArgs: lintArgs,
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						BaseArrayKey: "Issues",
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "golangci-lint",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "FromLinter",
						},
						Level: ci.JsonMappingInfo{
							Key: "Severity",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "Pos.Filename",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "Pos.Line",
						},
						StartCol: ci.JsonMappingInfo{
							Key: "Pos.Column",
						},
						Message: ci.JsonMappingInfo{
							Key: "Text",
						},
					},
				},
			})

		return rc, err

	case "deps":
		f, err := os.CreateTemp("/tmp", "sbom-*.json")
		if err != nil {
			return 1, fmt.Errorf("error creating temp file: %w", err)
		}
		defer os.Remove(f.Name())

		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "syft",
				CliArgs: []string{
					"scan",
					"--config",
					"config/syft/.syft.yaml",
					"--source-name",
					gitRepoBasePath,
					"--output",
					"spdx-json=" + f.Name(),
					".",
				},
			})
		if err != nil || rc != 0 {
			return rc, err
		}

		rc, _, _, err = lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "grype",
				CliArgs: []string{
					"--config",
					"config/grype/.grype.yaml",
					"--output",
					"json",
					f.Name(),
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						BaseArrayKey: "matches",
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "grype",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "vulnerability.id",
						},
						Level: ci.JsonMappingInfo{
							Key:          "vulnerability.severity",
							DefaultValue: "error",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "artifact.name",
						},
						Message: ci.JsonMappingInfo{
							Key: "vulnerability.description",
							Suffix: &ci.JsonMappingInfo{
								OverrideValue: " - ",
								Suffix: &ci.JsonMappingInfo{
									Key: "vulnerability.dataSource",
									Suffix: &ci.JsonMappingInfo{
										OverrideValue: " - Found version: ",
										Suffix: &ci.JsonMappingInfo{
											Key: "artifact.version",
											Suffix: &ci.JsonMappingInfo{
												OverrideValue: " - Constraint: ",
												Suffix: &ci.JsonMappingInfo{
													Key: "matchDetails.found.versionConstraint",
													Suffix: &ci.JsonMappingInfo{
														OverrideValue: " - Suggested version: ",
														Suffix: &ci.JsonMappingInfo{
															Key: "matchDetails.fix.suggestedVersion",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			})

		return rc, err

	case "markdown":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "markdownlint",
				CliArgs: []string{
					"--config",
					"config/markdownlint/.markdownlint.yaml",
					"--json",
				},
				Ext: ".md",
				Paths: []string{
					filesfind.FilesFindingRootPath,
				},
				JsonInfo: ci.JsonInfos{
					ReadFromStderr: true,
					Mappings: ci.JsonToFindingsMappings{
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "markdownlint",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "ruleNames",
						},
						Level: ci.JsonMappingInfo{
							OverrideValue: "error",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "fileName",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "lineNumber",
						},
						Message: ci.JsonMappingInfo{
							Key: "ruleDescription",
							Suffix: &ci.JsonMappingInfo{
								OverrideValue: " - ",
								Suffix: &ci.JsonMappingInfo{
									Key: "errorDetail",
									Suffix: &ci.JsonMappingInfo{
										OverrideValue: " - ",
										Suffix: &ci.JsonMappingInfo{
											Key: "ruleInformation",
										},
									},
								},
							},
						},
					},
				},
			})

		return rc, err

	case "shell":
		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "shellcheck",
				CliArgs: []string{
					"--format",
					"json",
				},
				Ext: ".sh",
				Paths: []string{
					filesfind.FilesFindingRootPath,
				},
				JsonInfo: ci.JsonInfos{
					Mappings: ci.JsonToFindingsMappings{
						ToolName: ci.JsonMappingInfo{
							OverrideValue: "shellcheck",
						},
						RuleID: ci.JsonMappingInfo{
							Key: "code",
						},
						Level: ci.JsonMappingInfo{
							Key: "level",
						},
						FilePath: ci.JsonMappingInfo{
							Key: "file",
						},
						StartLine: ci.JsonMappingInfo{
							Key: "line",
						},
						EndLine: ci.JsonMappingInfo{
							Key: "endLine",
						},
						StartCol: ci.JsonMappingInfo{
							Key: "column",
						},
						EndCol: ci.JsonMappingInfo{
							Key: "endColumn",
						},
						Message: ci.JsonMappingInfo{
							Key: "message",
						},
					},
				},
			})

		return rc, err

	case "release":
		err := git.TagSemver()
		if err != nil {
			return 1, fmt.Errorf("error tagging semver: %w", err)
		}

		rc, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "goreleaser",
				CliArgs: []string{
					"release",
					"--config",
					"config/goreleaser/.goreleaser.yaml",
					"--clean",
				},
			})

		return rc, err

	case "pr-check-title":
		finding, err := pr.CheckPRTitle(os.Args)
		if err != nil {
			return 1, fmt.Errorf("error checking PR title: %w", err)
		}

		if finding != (ci.Finding{}) {
			err := ci.PrintFindings([]ci.Finding{finding}, lint.GetOutputFormat())
			if err != nil {
				return 1, fmt.Errorf("error printing findings: %w", err)
			}

			return 1, fmt.Errorf("PR title check failed: %s", finding.Message)
		}

		slog.Debug("PR title check passed")

		return 0, nil

	case "repo-template-check-stale":
		finding, err := repotpl.CheckRepoTemplateUpdate(os.Args)
		if err != nil {
			return 1, fmt.Errorf("error checking stale repository template: %w", err)
		}

		if finding != (ci.Finding{}) {
			err := ci.PrintFindings([]ci.Finding{finding}, lint.GetOutputFormat())
			if err != nil {
				return 1, fmt.Errorf("error printing findings: %w", err)
			}

			return 1, fmt.Errorf("stale repository template check failed: %s", finding.Message)
		}

		slog.Debug("stale repository template check passed")

		return 0, nil

	case "branch-check-stale":
		finding, err := branch.CheckStaleBranches(os.Args)
		if err != nil {
			return 1, fmt.Errorf("error checking stale branches: %w", err)
		}

		if finding != (ci.Finding{}) {
			err := ci.PrintFindings([]ci.Finding{finding}, lint.GetOutputFormat())
			if err != nil {
				return 1, fmt.Errorf("error printing findings: %w", err)
			}

			return 1, fmt.Errorf("stale branches check failed: %s", finding.Message)
		}

		slog.Debug("stale branches check passed")

		return 0, nil

	case "ci":
		var waitGroup sync.WaitGroup

		commands := []string{
			"docker",
			"gha",
			"secrets",
			"sast",
			"go-test",
			"go-cover",
			"go-mod-tidy",
			"go-mod-name",
			"lint",
			"deps",
			"markdown",
			"shell",
		}
		waitGroup.Add(len(commands))

		failedCommands := make([]string, 0)

		for _, cmd := range commands {
			go func(command string) {
				defer waitGroup.Done()

				cmdArgs := []string{command}
				if command == "lint" && len(args) > 1 && args[1] == "--fix" {
					cmdArgs = append(cmdArgs, "--fix")
				}

				rc, err := DispatchCommand(config, cmdArgs)
				if err != nil {
					slog.Error(
						"Error executing command",
						slog.String("command", command),
						slog.String("error", err.Error()),
					)

					goRc += rc
				}

				if rc != 0 {
					slog.Error(
						"Command failed",
						slog.String("command", command),
						slog.Int("returnCode", rc),
					)

					failedCommands = append(failedCommands, command)
				} else {
					slog.Debug("Command succeeded", slog.String("command", command))
				}
			}(cmd)
		}

		waitGroup.Wait()

		if goRc != 0 {
			return goRc, fmt.Errorf(
				"one or more commands failed: %s",
				strings.Join(failedCommands, ", "),
			)
		}

		slog.Info("All commands succeeded")

		return 0, nil

	default:
		return 1, fmt.Errorf("unknown command: %s", args[0])
	}
}
