// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package lint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"github.com/kemadev/ci-cd/internal/app/config"
	"github.com/kemadev/ci-cd/pkg/ci"
	"github.com/kemadev/ci-cd/pkg/filesfind"
)

type LinterArgs struct {
	Bin      string
	Ext      string
	Paths    []string
	CliArgs  []string
	Workdir  string
	JSONInfo ci.JSONInfos
	// Return non-zero exit code if at least one finding is found
	FailOnAtLeastOneFinding bool
}

const (
	// NOTE read buffer size is limited, any output line (split function) larger than this will cause deadlock.
	MaxBufferSize = 32 * 1024 * 1024 // 32MB
)

var ErrNoLinterBinary = fmt.Errorf("linter binary is required")

func processPipe(
	config *config.Config,
	pipe io.Reader,
	buf *bytes.Buffer,
	output *os.File,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	reader := io.TeeReader(pipe, buf)
	scanner := bufio.NewScanner(reader)
	// Some linters can output a lot of data, in a one-line json format, however deadlock can occur if the output line is larger than the buffer size
	lb := make([]byte, 0, MaxBufferSize)
	scanner.Buffer(lb, len(lb))

	for scanner.Scan() {
		line := scanner.Text()
		if config.DebugEnabled {
			_, err := fmt.Fprintln(output, line)
			if err != nil {
				slog.Error(
					"error writing to output",
					slog.String("line", line),
					slog.Any("err", err),
				)
			}
		}
	}
}

func startCmd(
	config *config.Config,
	lintArgs LinterArgs,
	args []string,
) (*sync.WaitGroup, *exec.Cmd, *bytes.Buffer, *bytes.Buffer, error) {
	//nolint:gosec // we purposefully pass user controlled arguments, this script does not run outside of CI
	cmd := exec.Command(lintArgs.Bin, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error creating stderr pipe: %w", err)
	}

	if lintArgs.Workdir != "" {
		cmd.Dir = lintArgs.Workdir
	}

	var stdoutBuf, stderrBuf bytes.Buffer

	var waitGroup sync.WaitGroup

	//nolint:mnd // stdout and stderr obviously make 2 output files
	outputs := make([]*os.File, 2)
	outputs[0] = os.Stdout
	outputs[1] = os.Stderr
	numPipes := len(outputs)

	waitGroup.Add(numPipes)

	go processPipe(config, stdoutPipe, &stdoutBuf, outputs[0], &waitGroup)
	go processPipe(config, stderrPipe, &stderrBuf, outputs[1], &waitGroup)

	err = cmd.Start()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error starting command: %w", err)
	}

	return &waitGroup, cmd, &stdoutBuf, &stderrBuf, nil
}

func GetOutputFormat() string {
	format := "human"
	if os.Getenv("GITHUB_ACTIONS") != "" {
		format = "github"
	}

	return format
}

func RunLinter(config *config.Config, lintArgs LinterArgs) (int, string, string, error) {
	if lintArgs.Bin == "" {
		return 1, "", "", ErrNoLinterBinary
	}

	files := []string{}

	if lintArgs.Paths != nil {
		filesList, err := filesfind.FindFilesByExtension(filesfind.FilesFindingArgs{
			Extension:   lintArgs.Ext,
			Paths:       lintArgs.Paths,
			Recursive:   true,
			IgnorePaths: []string{},
		})
		if err != nil {
			return 1, "", "", fmt.Errorf("error finding files: %w", err)
		}

		files = filesList

		if len(files) == 0 {
			slog.Info("no file found")

			return 0, "", "", nil
		}

		for _, file := range files {
			slog.Debug("found file", slog.String("file", file))
		}
	}

	args := lintArgs.CliArgs
	args = append(args, files...)

	format := GetOutputFormat()

	slog.Debug("running linter",
		slog.String("binary", lintArgs.Bin),
		slog.String("args", fmt.Sprintf("%v", args)),
		slog.String("format", format),
		slog.Bool("failOnAtLeastOneFinding", lintArgs.FailOnAtLeastOneFinding),
	)

	waitGroup, cmd, stdoutBuf, stderrBuf, err := startCmd(config, lintArgs, args)
	if err != nil {
		return 1, "", "", fmt.Errorf("error preparing command: %w", err)
	}

	waitGroup.Wait()

	rc, err := handleLinterOutcome(cmd, stdoutBuf, stderrBuf, format, lintArgs)
	if err != nil {
		return rc, "", "", fmt.Errorf("error handling linter outcome: %w", err)
	}

	return rc, stdoutBuf.String(), stderrBuf.String(), nil
}

func handleLinterOutcome(
	cmd *exec.Cmd,
	stdoutBuf *bytes.Buffer,
	stderrBuf *bytes.Buffer,
	format string,
	args LinterArgs,
) (int, error) {
	var findings []ci.Finding

	err := cmd.Wait()
	if err != nil {
		slog.Error("command execution failed", slog.String("error", err.Error()))
	} else {
		slog.Info("command executed successfully")
	}

	retCode := cmd.ProcessState.ExitCode()

	switch args.JSONInfo.Type {
	case "none":
		slog.Debug(
			"No finding parsing requested, skipping",
			slog.String("type", args.JSONInfo.Type),
		)
	case "plain":
		if len(stdoutBuf.String()) == 0 {
			return 0, nil
		}

		find := ci.Finding{
			ToolName:  args.JSONInfo.Mappings.ToolName.OverrideValue,
			RuleID:    args.JSONInfo.Mappings.RuleID.OverrideValue,
			Level:     args.JSONInfo.Mappings.Level.OverrideValue,
			FilePath:  args.JSONInfo.Mappings.FilePath.OverrideValue,
			Message:   args.JSONInfo.Mappings.Message.OverrideValue,
			StartLine: 0,
			EndLine:   0,
			StartCol:  0,
			EndCol:    0,
		}
		findings = append(findings, find)
	default:
		var str string

		if args.JSONInfo.ReadFromStderr {
			str = stderrBuf.String()
		} else {
			str = stdoutBuf.String()
		}

		fa, err := ci.FindingsFromJSON(str, args.JSONInfo)
		if err != nil {
			return 1, fmt.Errorf("error parsing findings: %w", err)
		}

		findings = append(findings, fa...)
	}

	if args.FailOnAtLeastOneFinding && len(findings) > 0 {
		slog.Error(
			"findings found",
			slog.Bool("FailOnAtLeastOneFinding", args.FailOnAtLeastOneFinding),
		)

		retCode = 1
	}

	err = ci.PrintFindings(findings, format)
	if err != nil {
		return 1, fmt.Errorf("error printing findings: %w", err)
	}

	return retCode, nil
}
