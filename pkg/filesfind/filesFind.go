package filesfind

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	FilesFindingRootPath = "/src"
)

var ErrNoExtension = fmt.Errorf("file extension is required")

type FilesFindingArgs struct {
	Extension   string
	Paths       []string
	IgnorePaths []string
	Recursive   bool
}

func handleArgs(args FilesFindingArgs) (FilesFindingArgs, error) {
	if args.Paths == nil {
		args.Paths = []string{"/src"}
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
