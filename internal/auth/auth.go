// Copyright 2025 kemadev
// SPDX-License-Identifier: MPL-2.0

package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/kemadev/ci-cd/pkg/auth"
)

var ErrNetrcNotSet = errors.New("netrc file environment variable is not set")

func CreateNetrcFromEnv() error {
	netrcContent, present := os.LookupEnv(auth.NetrcEnvVarKey)
	if !present {
		return fmt.Errorf("error finding netrc environment variable: %w", ErrNetrcNotSet)
	}

	f, err := os.Create(os.Getenv("HOME") + "/.netrc")
	if err != nil {
		return fmt.Errorf("error creating netrc file: %w", err)
	}

	_, err = f.WriteString(netrcContent)
	if err != nil {
		return fmt.Errorf("error writing netrc file: %w", err)
	}

	return nil
}
