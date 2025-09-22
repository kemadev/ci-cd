// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

//nolint:cyclop // the enormous switch is (hopefully) easily understandable for a human
package dispatch

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/kemadev/ci-cd/internal/branch"
	"github.com/kemadev/ci-cd/internal/config"
	"github.com/kemadev/ci-cd/internal/lint"
	"github.com/kemadev/ci-cd/internal/pr"
	"github.com/kemadev/ci-cd/pkg/ci"
	"github.com/kemadev/ci-cd/pkg/filesfind"
	"github.com/kemadev/go-framework/pkg/git"
)

var (
	ErrUnknownCommand    = fmt.Errorf("unknown command")
	ErrNoCommandProvided = fmt.Errorf("no command provided")
	ErrExitCodeNotZero   = fmt.Errorf("exit code is not zero")
	ErrFindingFound      = fmt.Errorf("finding found")
	ErrCommandFailed     = fmt.Errorf("command failed")
)

const (
	CommandDocker           = "docker"
	CommandGHA              = "gha"
	CommandSecrets          = "secrets"
	CommandSAST             = "sast"
	CommandGoTest           = "go-test"
	CommandGoCover          = "go-cover"
	CommandGoBuild          = "go-build"
	CommandGoModTidy        = "go-mod-tidy"
	CommandGoModName        = "go-mod-name"
	CommandGoLint           = "go-lint"
	CommandDeps             = "deps"
	CommandMarkdown         = "markdown"
	CommandShell            = "shell"
	CommandRelease          = "release"
	CommandPRTitleCheck     = "pr-title-check"
	CommandBranchStaleCheck = "branch-stale-check"
	CommandCI               = "ci"
	CommandDepsBump         = "deps-bump"
	CommandHelp             = "help"
)

//nolint:funlen // the enormous switch is (hopefully) easily understandable for a human
func Run(config *config.Config, args []string) (int, error) {
	gitSvc := git.NewGitService()

	gitRepoBasePath, err := gitSvc.GetGitBasePath()
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

	filesFindRootPath, err := filesfind.GetFilesFindingRootPath()
	if err != nil {
		return 1, fmt.Errorf("error getting files finding root path: %w", err)
	}

	goRc := 0
	goErr := error(nil)

	switch args[0] {
	case CommandDocker:
		slog.Info("running " + CommandDocker)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "hadolint",
				Ext: "Dockerfile",
				Paths: []string{
					filesFindRootPath,
				},
				CliArgs: []string{
					"--format",
					"json",
				},
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "hadolint",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "code",
						},
						Level: ci.JSONMappingInfo{
							Key: "level",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "file",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "line",
						},
						Message: ci.JSONMappingInfo{
							Key: "message",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandDocker+": %w", err)
		}

		return retCode, nil

	case CommandGHA:
		slog.Info("running " + CommandGHA)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "actionlint",
				Ext: ".yaml",
				Paths: []string{
					filesFindRootPath + "/.github/workflows",
					filesFindRootPath + "/.github/actions",
				},
				CliArgs: []string{
					"-format",
					"{{json .}}",
				},
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "gha-actionlint",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "kind",
						},
						Level: ci.JSONMappingInfo{
							OverrideValue: "warning",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "filepath",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "line",
						},
						StartCol: ci.JSONMappingInfo{
							Key: "col",
						},
						Message: ci.JSONMappingInfo{
							Key: "message",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandGHA+": %w", err)
		}

		return retCode, nil

	case CommandSecrets:
		slog.Info("running " + CommandSecrets)

		retCode, _, _, err := lint.RunLinter(
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
					"/var/config/gitleaks/.gitleaksignore",
					"--report-path",
					"-",
				},
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "secrets-gitleaks",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "RuleID",
						},
						Level: ci.JSONMappingInfo{
							OverrideValue: "error",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "File",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "StartLine",
						},
						EndLine: ci.JSONMappingInfo{
							Key: "EndLine",
						},
						StartCol: ci.JSONMappingInfo{
							Key: "StartColumn",
						},
						EndCol: ci.JSONMappingInfo{
							Key: "EndColumn",
						},
						Message: ci.JSONMappingInfo{
							Key: "Description",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandSecrets+": %w", err)
		}

		return retCode, nil

	case CommandSAST:
		slog.Info("running " + CommandSAST)

		retCode, _, _, err := lint.RunLinter(
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
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						BaseArrayKey: "results",
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "sast-semgrep",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "check_id",
						},
						Level: ci.JSONMappingInfo{
							Key: "extra.severity",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "path",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "start.line",
						},
						EndLine: ci.JSONMappingInfo{
							Key: "end.line",
						},
						StartCol: ci.JSONMappingInfo{
							Key: "start.col",
						},
						EndCol: ci.JSONMappingInfo{
							Key: "end.col",
						},
						Message: ci.JSONMappingInfo{
							Key: "extra.message",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandSAST+": %w", err)
		}

		return retCode, nil

	case CommandGoTest:
		slog.Info("running " + CommandGoTest)

		for _, mod := range goModList {
			if strings.HasPrefix(mod, filesFindRootPath+"/deploy/") {
				slog.Info("skipping "+CommandGoTest, slog.String("mod", mod))

				continue
			}

			slog.Info("running "+CommandGoTest, slog.String("mod", mod))
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
					JSONInfo: ci.JSONInfos{
						Type: "stream",
						Mappings: ci.JSONToFindingsMappings{
							ToolName: ci.JSONMappingInfo{
								OverrideValue: "go-test",
							},
							RuleID: ci.JSONMappingInfo{
								OverrideValue: "no-failing-test",
							},
							Level: ci.JSONMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JSONMappingInfo{
								Key: "Package",
								// Get path relative to git repo base path
								ValueTransformerRegex: gitRepoBasePath + "/(.*)",
								Suffix: &ci.JSONMappingInfo{
									// Add a /
									OverrideValue: "/",
									Suffix: &ci.JSONMappingInfo{
										Key: "Output",
										// Name of test file producing the finding
										ValueTransformerRegex: `\s*(\w+_test.go):`,
									},
								},
							},
							StartLine: ci.JSONMappingInfo{
								Key:                   "Output",
								ValueTransformerRegex: `\s*\w_test.go:(\d+):`,
							},
							Message: ci.JSONMappingInfo{
								Key:                   "Output",
								GlobalSelectorRegex:   `\s*(\w_test.go:\d+):`,
								ValueTransformerRegex: `\s*\w_test.go:\d+:\s*(.*)`,
							},
						},
					},
				})

			if retCode != 0 {
				goRc = 1
			}

			if err != nil {
				return 1, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, nil

	case CommandGoCover:
		slog.Info("running " + CommandGoCover)

		for _, mod := range goModList {
			if strings.HasPrefix(mod, filesFindRootPath+"/deploy/") {
				slog.Info("skipping "+CommandGoCover, slog.String("mod", mod))

				continue
			}

			slog.Info("running "+CommandGoCover, slog.String("mod", mod))
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
					JSONInfo: ci.JSONInfos{
						Type: "stream",
						Mappings: ci.JSONToFindingsMappings{
							ToolName: ci.JSONMappingInfo{
								OverrideValue: "go-cover",
							},
							RuleID: ci.JSONMappingInfo{
								OverrideValue: "no-cover-below-70",
							},
							Level: ci.JSONMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JSONMappingInfo{
								Key: "Package",
								// Get path relative to git repo base path
								ValueTransformerRegex: gitRepoBasePath + "/(.*)",
							},
							Message: ci.JSONMappingInfo{
								Key: "Output",
								// Failed test or coverage lesser than 70%
								GlobalSelectorRegex:   `coverage:\s*([0-6](\d)?(\.\d)?)\% of statements`,
								ValueTransformerRegex: `coverage:\s*([0-6](\d)?(\.\d)?\%) of statements`,
								Suffix: &ci.JSONMappingInfo{
									OverrideValue: " package coverage is below 70%",
								},
							},
						},
					},
				})

			if retCode != 0 {
				goRc = 1
			}

			if err != nil {
				return 1, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case CommandGoBuild:
		slog.Info("running " + CommandGoBuild)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "goreleaser",
				CliArgs: []string{
					"build",
					"--config",
					"/var/config/goreleaser/.goreleaser.yaml",
					"--clean",
					"--snapshot",
				},
				JSONInfo: ci.JSONInfos{
					Type: "none",
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandGoBuild+": %w", err)
		}

		return retCode, nil

	case CommandGoModTidy:
		slog.Info("running " + CommandGoModTidy)

		for _, mod := range goModList {
			slog.Info("running "+CommandGoModTidy, slog.String("mod", mod))
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
					JSONInfo: ci.JSONInfos{
						Type: "plain",
						Mappings: ci.JSONToFindingsMappings{
							ToolName: ci.JSONMappingInfo{
								OverrideValue: "go-mod-tidy",
							},
							RuleID: ci.JSONMappingInfo{
								OverrideValue: "no-unused-dependency",
							},
							Level: ci.JSONMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JSONMappingInfo{
								OverrideValue: strings.TrimPrefix(
									strings.Join(
										strings.Split(mod, filesFindRootPath)[1:],
										"",
									),
									"/",
								),
							},
							Message: ci.JSONMappingInfo{
								OverrideValue: "Unused dependencies found in " + mod,
							},
						},
					},
				})

			if retCode != 0 {
				goRc = 1
			}

			if err != nil {
				return 1, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, nil

	case CommandGoModName:
		slog.Info("running " + CommandGoModName)

		for _, mod := range goModList {
			slog.Info("running "+CommandGoModName, slog.String("mod", mod))
			expectedGoModName := gitRepoBasePath + strings.Split(strings.Join(strings.Split(mod, filesFindRootPath)[1:], ""), "/go.mod")[0]
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
					JSONInfo: ci.JSONInfos{
						Type: "object",
						Mappings: ci.JSONToFindingsMappings{
							ToolName: ci.JSONMappingInfo{
								OverrideValue: "go-mod-name",
							},
							RuleID: ci.JSONMappingInfo{
								OverrideValue: "mod-name-must-match-repo-structure",
							},
							Level: ci.JSONMappingInfo{
								OverrideValue: "error",
							},
							FilePath: ci.JSONMappingInfo{
								OverrideValue: strings.TrimPrefix(
									strings.Join(
										strings.Split(mod, filesFindRootPath)[1:],
										"",
									),
									"/",
								),
							},
							Message: ci.JSONMappingInfo{
								Key: "Module.Path",
								GlobalSelectorRegex: strings.ReplaceAll(
									expectedGoModName,
									".",
									`\.`,
								) + "$",
								InvertGlobalSelector: true,
								Suffix: &ci.JSONMappingInfo{
									OverrideValue: " does not match the repository structure, module name should be ",
									Suffix: &ci.JSONMappingInfo{
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
				return 1, fmt.Errorf("error running go test in %s: %w", mod, err)
			}
		}

		return goRc, goErr

	case CommandGoLint:
		var fixEnabled bool
		if len(args) > 1 && args[1] == "--fix" {
			fixEnabled = true
		}

		slog.Info("running "+CommandGoLint, slog.Bool("fixEnabled", fixEnabled))

		lintArgs := []string{
			"run",
			"--config",
			"/var/config/golangci-lint/.golangci.yaml",
			"--show-stats=false",
			"--output.json.path",
			"stdout",
		}
		if fixEnabled {
			lintArgs = append(lintArgs, "--fix")
		}

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin:     "golangci-lint",
				CliArgs: lintArgs,
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						BaseArrayKey: "Issues",
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "golangci-lint",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "FromLinter",
						},
						Level: ci.JSONMappingInfo{
							Key: "Severity",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "Pos.Filename",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "Pos.Line",
						},
						StartCol: ci.JSONMappingInfo{
							Key: "Pos.Column",
						},
						Message: ci.JSONMappingInfo{
							Key: "Text",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandGoLint+": %w", err)
		}

		return retCode, nil

	case CommandDeps:
		sbomFile, err := os.CreateTemp("/tmp", "sbom-*.json")
		if err != nil {
			return 1, fmt.Errorf("error creating temp file: %w", err)
		}

		defer os.Remove(sbomFile.Name())

		slog.Info(
			"running "+CommandDeps,
			slog.String("step", "syft"),
			slog.String("outputFile", sbomFile.Name()),
		)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "syft",
				CliArgs: []string{
					"scan",
					"--config",
					"/var/config/syft/.syft.yaml",
					"--source-name",
					gitRepoBasePath,
					"--output",
					"spdx-json=" + sbomFile.Name(),
					"--enrich",
					"go",
					".",
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandDeps+": %w", err)
		}

		if retCode != 0 {
			return retCode, fmt.Errorf(CommandDeps+": %w", ErrExitCodeNotZero)
		}

		slog.Info(
			"running "+CommandDeps,
			slog.String("step", "grype"),
			slog.String("outputFile", sbomFile.Name()),
		)

		retCode, _, _, err = lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "grype",
				CliArgs: []string{
					"--config",
					"/var/config/grype/.grype.yaml",
					"--output",
					"json",
					sbomFile.Name(),
				},
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						BaseArrayKey: "matches",
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "grype",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "vulnerability.id",
						},
						Level: ci.JSONMappingInfo{
							Key:          "vulnerability.severity",
							DefaultValue: "error",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "artifact.name",
						},
						Message: ci.JSONMappingInfo{
							Key: "vulnerability.description",
							Suffix: &ci.JSONMappingInfo{
								OverrideValue: " - ",
								Suffix: &ci.JSONMappingInfo{
									Key: "vulnerability.dataSource",
									Suffix: &ci.JSONMappingInfo{
										OverrideValue: " - Found version: ",
										Suffix: &ci.JSONMappingInfo{
											Key: "artifact.version",
											Suffix: &ci.JSONMappingInfo{
												OverrideValue: " - Constraint: ",
												Suffix: &ci.JSONMappingInfo{
													Key: "matchDetails.found.versionConstraint",
													Suffix: &ci.JSONMappingInfo{
														OverrideValue: " - Suggested version: ",
														Suffix: &ci.JSONMappingInfo{
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
		if err != nil {
			return 1, fmt.Errorf(CommandDeps+": %w", err)
		}

		return retCode, nil

	case CommandMarkdown:
		slog.Info("running " + CommandMarkdown)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "markdownlint",
				CliArgs: []string{
					"--config",
					"/var/config/markdownlint/.markdownlint.yaml",
					"--json",
				},
				Ext: ".md",
				Paths: []string{
					filesFindRootPath,
				},
				JSONInfo: ci.JSONInfos{
					ReadFromStderr: true,
					Mappings: ci.JSONToFindingsMappings{
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "markdownlint",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "ruleNames",
						},
						Level: ci.JSONMappingInfo{
							OverrideValue: "error",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "fileName",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "lineNumber",
						},
						Message: ci.JSONMappingInfo{
							Key: "ruleDescription",
							Suffix: &ci.JSONMappingInfo{
								OverrideValue: " - ",
								Suffix: &ci.JSONMappingInfo{
									Key: "errorDetail",
									Suffix: &ci.JSONMappingInfo{
										OverrideValue: " - ",
										Suffix: &ci.JSONMappingInfo{
											Key: "ruleInformation",
										},
									},
								},
							},
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandMarkdown+": %w", err)
		}

		return retCode, nil

	case CommandShell:
		slog.Info("running " + CommandShell)

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "shellcheck",
				CliArgs: []string{
					"--format",
					"json",
				},
				Ext: ".sh",
				Paths: []string{
					filesFindRootPath,
				},
				JSONInfo: ci.JSONInfos{
					Mappings: ci.JSONToFindingsMappings{
						ToolName: ci.JSONMappingInfo{
							OverrideValue: "shellcheck",
						},
						RuleID: ci.JSONMappingInfo{
							Key: "code",
						},
						Level: ci.JSONMappingInfo{
							Key: "level",
						},
						FilePath: ci.JSONMappingInfo{
							Key: "file",
						},
						StartLine: ci.JSONMappingInfo{
							Key: "line",
						},
						EndLine: ci.JSONMappingInfo{
							Key: "endLine",
						},
						StartCol: ci.JSONMappingInfo{
							Key: "column",
						},
						EndCol: ci.JSONMappingInfo{
							Key: "endColumn",
						},
						Message: ci.JSONMappingInfo{
							Key: "message",
						},
					},
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandShell+": %w", err)
		}

		return retCode, nil

	case CommandRelease:
		slog.Info("running " + CommandRelease)
		slog.Info("running "+CommandRelease, slog.String("step", "tag-semver"))

		skip, err := gitSvc.TagSemver()
		if err != nil {
			return 1, fmt.Errorf("error tagging semver: %w", err)
		}

		if skip {
			slog.Info("skipping release step, no new semver tag created")

			return 0, nil
		}

		slog.Info("running "+CommandRelease, slog.String("step", "goreleaser"))

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin: "goreleaser",
				CliArgs: []string{
					"release",
					"--config",
					"/var/config/goreleaser/.goreleaser.yaml",
					"--clean",
				},
			})
		if retCode != 0 {
			return retCode, fmt.Errorf(
				"error running goreleaser: exit code %d: %w",
				retCode,
				ErrExitCodeNotZero,
			)
		}

		if err != nil {
			return 1, fmt.Errorf(CommandRelease+": %w", err)
		}

		return retCode, nil

	case CommandPRTitleCheck:
		slog.Info("running " + CommandPRTitleCheck)

		finding, err := pr.CheckPRTitle(os.Args[2])
		if err != nil {
			return 1, fmt.Errorf("error checking PR title: %w", err)
		}

		if finding != (ci.Finding{}) {
			err := ci.PrintFindings([]ci.Finding{finding}, lint.GetOutputFormat())
			if err != nil {
				return 1, fmt.Errorf("error printing findings: %w", err)
			}

			return 1, fmt.Errorf("pr title check failed: %s: %w", finding.Message, ErrFindingFound)
		}

		slog.Info("pr title check passed")

		return 0, nil

	case CommandBranchStaleCheck:
		slog.Info("running " + CommandBranchStaleCheck)

		finding, err := branch.CheckStaleBranches(gitSvc)
		if err != nil {
			return 1, fmt.Errorf("error checking stale branches: %w", err)
		}

		if finding != (ci.Finding{}) {
			err := ci.PrintFindings([]ci.Finding{finding}, lint.GetOutputFormat())
			if err != nil {
				return 1, fmt.Errorf("error printing findings: %w", err)
			}

			return 1, fmt.Errorf(
				"stale branches check failed: %s: %w",
				finding.Message,
				ErrFindingFound,
			)
		}

		slog.Info("stale branches check passed")

		return 0, nil

	case CommandCI:
		slog.Info("running " + CommandCI)

		var waitGroup sync.WaitGroup

		commands := []string{
			"docker",
			"gha",
			"secrets",
			"sast",
			"go-test",
			"go-cover",
			"go-build",
			"go-mod-tidy",
			"go-mod-name",
			"go-lint",
			"deps",
			"markdown",
			"shell",
		}
		waitGroup.Add(len(commands))

		var (
			failedCommands   []string
			failedCommandsMu sync.Mutex
		)

		for _, cmd := range commands {
			slog.Info("running command", slog.String("command", cmd))

			go func(command string) {
				defer waitGroup.Done()

				cmdArgs := []string{command}
				if command == CommandGoLint && len(args) > 1 && args[1] == "--fix" {
					cmdArgs = append(cmdArgs, "--fix")
				}

				retCode, err := Run(config, cmdArgs)
				if err != nil {
					slog.Error(
						"Error executing command",
						slog.String("command", command),
						slog.String("error", err.Error()),
					)
				}

				if retCode != 0 {
					slog.Error(
						"Command failed",
						slog.String("command", command),
						slog.Int("returnCode", retCode),
					)

					failedCommandsMu.Lock()

					failedCommands = append(failedCommands, command)

					failedCommandsMu.Unlock()
				} else {
					slog.Debug("Command succeeded", slog.String("command", command))
				}
			}(cmd)
		}

		waitGroup.Wait()

		if len(failedCommands) > 0 {
			return 1, fmt.Errorf(
				"one or more commands failed: %s: %w",
				strings.Join(failedCommands, ", "),
				ErrCommandFailed,
			)
		}

		slog.Info("All commands succeeded")

		return 0, nil

	case CommandDepsBump:
		slog.Info("running " + CommandDepsBump)

		if config.DebugEnabled {
			os.Setenv("LOG_LEVEL", "debug")
		}

		retCode, _, _, err := lint.RunLinter(
			config,
			lint.LinterArgs{
				Bin:     "renovate",
				CliArgs: []string{},
				JSONInfo: ci.JSONInfos{
					Type: "none",
				},
			})
		if err != nil {
			return 1, fmt.Errorf(CommandDepsBump+": %w", err)
		}

		return retCode, nil

	case "help":
		slog.Info("Available commands:")
		slog.Info("  " + CommandDocker + " - Run Dockerfile linter")
		slog.Info("  " + CommandGHA + " - Run GitHub Actions linter")
		slog.Info("  " + CommandSecrets + " - Run secrets detection")
		slog.Info("  " + CommandSAST + " - Run Static Application Security Testing (SAST)")
		slog.Info("  " + CommandGoTest + " - Run Go tests")
		slog.Info("  " + CommandGoCover + " - Run Go test coverage")
		slog.Info("  " + CommandGoModTidy + " - Run Go mod tidyness check")
		slog.Info("  " + CommandGoModName + " - Check Go module name check")
		slog.Info("  " + CommandGoLint + " - Run Go linter")
		slog.Info("  " + CommandDeps + " - Run dependency analysis")
		slog.Info("  " + CommandMarkdown + " - Run Markdown linter")
		slog.Info("  " + CommandShell + " - Run Shell script linter")
		slog.Info("  " + CommandRelease + " - Run release process")
		slog.Info("  " + CommandPRTitleCheck + " - Check PR title format")
		slog.Info("  " + CommandBranchStaleCheck + " - Check for stale branches")
		slog.Info("  " + CommandCI + " - Run all CI commands (mimics GitHub Pull Request CI)")
		slog.Info("  " + CommandHelp + " - Show this help message")

		return 0, nil

	default:
		return 1, fmt.Errorf("command %s: %w", args[0], ErrUnknownCommand)
	}
}
