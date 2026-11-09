package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"custos/internal/platform/selfupdate"
)

// Seams for tests: self-update reaches the network, the package managers
// and the running binary.
var (
	detectInstall = selfupdate.Detect
	latestRelease = selfupdate.Latest
	replaceBinary = selfupdate.Replace
	lookPath      = exec.LookPath
	runCommand    = func(dir string, stdout, stderr io.Writer, name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = dir, os.Stdin, stdout, stderr
		return cmd.Run()
	}
)

// Package-manager commands that update an install they own.
var (
	brewUpgrade    = []string{"brew", "upgrade", "janalis/tap/custos"}
	composerUpdate = []string{"composer", "update", "janalis/custos"}
)

func cmdSelfUpdate(version string, args []string, stdout, stderr io.Writer) (int, error) {
	fs := newFlagSet("self-update", stderr)
	check := fs.Bool("check", false, "only report whether a newer release exists (exit 1 when one does)")
	target := fs.String("version", "", "install this release instead of the latest (release binaries only)")
	if err := fs.Parse(args); err != nil {
		return 2, err
	}
	if fs.NArg() > 0 {
		return 2, errors.New("self-update takes no arguments")
	}
	if *target != "" && !selfupdate.Valid(*target) {
		return 2, fmt.Errorf("--version must be X.Y.Z, got %q", *target)
	}
	in, err := detectInstall(version)
	if err != nil {
		return 2, err
	}
	if in.Method == selfupdate.Source {
		return 2, errors.New("this binary was built from source: pull and rebuild it (make build)")
	}
	if *target != "" && in.Method != selfupdate.Standalone {
		return 2, errors.New("--version only applies to release binaries; ask the package manager for a version")
	}
	ctx := context.Background()
	latest := *target
	if latest == "" {
		if latest, err = latestRelease(ctx); err != nil {
			return 2, err
		}
	}
	if *target == "" && !selfupdate.Newer(version, latest) {
		fmt.Fprintf(stdout, "custos %s is up to date\n", version)
		return 0, nil
	}
	if *check {
		fmt.Fprintf(stdout, "custos %s is available (you have %s)\n", latest, version)
		return 1, nil
	}
	switch in.Method {
	case selfupdate.Homebrew:
		return runManager("", stdout, stderr, brewUpgrade)
	case selfupdate.Composer:
		if in.Root == "" {
			return 2, fmt.Errorf("cannot find the Composer project requiring custos; run %q in it", strings.Join(composerUpdate, " "))
		}
		code, err := runManager(in.Root, stdout, stderr, composerUpdate)
		if err == nil {
			fmt.Fprintf(stderr, "custos: if composer.json keeps an older version, run: composer require --dev janalis/custos:^%s\n", latest)
		}
		return code, err
	}
	if err := replaceBinary(ctx, in.Path, latest); err != nil {
		return 2, err
	}
	fmt.Fprintf(stdout, "custos %s → %s\n", version, latest)
	return 0, nil
}

// runManager runs a package-manager command in dir with the output passed
// through.
func runManager(dir string, stdout, stderr io.Writer, argv []string) (int, error) {
	line := strings.Join(argv, " ")
	if _, err := lookPath(argv[0]); err != nil {
		return 2, fmt.Errorf("%s is not on the PATH; run: %s", argv[0], line)
	}
	fmt.Fprintf(stderr, "custos: running %s\n", line)
	if err := runCommand(dir, stdout, stderr, argv[0], argv[1:]...); err != nil {
		return 2, fmt.Errorf("%s: %w", line, err)
	}
	return 0, nil
}
