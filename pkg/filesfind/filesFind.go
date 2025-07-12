// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package filesfind

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

var ErrNoExtension = fmt.Errorf("file extension is required")

type FilesFindingArgs struct {
	Extension   string
	Paths       []string
	IgnorePaths []string
	Recursive   bool
}

func GetFilesFindingRootPath() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	return workDir, nil
}

func handleArgs(args FilesFindingArgs) (FilesFindingArgs, error) {
	rootPath, err := GetFilesFindingRootPath()
	if err != nil {
		return FilesFindingArgs{}, fmt.Errorf("error getting root path: %w", err)
	}

	if args.Paths == nil {
		args.Paths = []string{rootPath}
	}

	if args.Extension == "" {
		return FilesFindingArgs{}, ErrNoExtension
	}

	if args.IgnorePaths == nil {
		args.IgnorePaths = []string{}
	}

	return args, nil
}

func FindFilesByExtension(args FilesFindingArgs) ([]string, error) {
	fileArgs, err := handleArgs(args)
	if err != nil {
		return nil, fmt.Errorf("error handling args: %w", err)
	}

	files := []string{}

	for _, path := range fileArgs.Paths {
		d, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("error reading directory %s: %w", path, err)
		}

		for _, entry := range d {
			if entry.IsDir() {
				if !fileArgs.Recursive || slices.Contains(fileArgs.IgnorePaths, entry.Name()) {
					continue
				}

				subDirPath := fmt.Sprintf("%s/%s", path, entry.Name())
				subDirArgs := fileArgs
				subDirArgs.Paths = []string{subDirPath}

				subDirFiles, err := FindFilesByExtension(subDirArgs)
				if err != nil {
					return nil, fmt.Errorf(
						"error finding files in subdirectory %s: %w",
						subDirPath,
						err,
					)
				}

				files = append(files, subDirFiles...)
			} else if strings.HasSuffix(entry.Name(), fileArgs.Extension) {
				filePath := fmt.Sprintf("%s/%s", path, entry.Name())
				files = append(files, filePath)
			}
		}
	}

	return files, nil
}
