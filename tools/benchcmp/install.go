//go:build unix

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

// install puts each package used by tools at its pinned version in its own
// Composer project, and each corpus at its tag with its dependencies
// (analyzers resolve vendor symbols). Up-to-date entries are kept.
func (b *bench) install(tools []Tool, corpora []Corpus, log io.Writer) error {
	var pkgs []string
	for _, t := range tools {
		if t.Package != "" && !slices.Contains(pkgs, t.Package) {
			pkgs = append(pkgs, t.Package)
		}
	}
	for _, pkg := range pkgs {
		version := b.cfg.Packages[pkg]
		dir := b.packageDir(pkg)
		if err := b.fresh(dir, version, log, func() error {
			// No package script or plugin runs: tools are only executed
			// when measured.
			return sh(log, dir, "composer", "require", "--no-interaction", "--no-progress",
				"--no-scripts", "--no-plugins", pkg+":"+version)
		}); err != nil {
			return fmt.Errorf("%s: %w", pkg, err)
		}
	}
	for _, c := range corpora {
		dir := b.corpusDir(c)
		if err := b.fresh(dir, c.Tag, log, func() error {
			if err := sh(log, "", "git", "-c", "advice.detachedHead=false", "clone", "--quiet", "--depth", "1", "--branch", c.Tag, c.Repo, dir); err != nil {
				return err
			}
			if _, err := os.Stat(filepath.Join(dir, "composer.json")); err != nil {
				return nil
			}
			return sh(log, dir, "composer", "install", "--no-interaction", "--no-progress",
				"--no-dev", "--no-scripts", "--no-plugins", "--ignore-platform-reqs")
		}); err != nil {
			return fmt.Errorf("%s: %w", c.Name, err)
		}
	}
	return nil
}

// fresh rebuilds dir with build unless its stamp records version.
func (b *bench) fresh(dir, version string, log io.Writer, build func() error) error {
	stamp := dir + ".version"
	if got, err := os.ReadFile(stamp); err == nil && string(got) == version {
		return nil
	}
	fmt.Fprintf(log, "installing %s %s\n", filepath.Base(dir), version)
	if err := errors.Join(os.RemoveAll(dir), os.MkdirAll(dir, 0o755)); err != nil {
		return err
	}
	if err := build(); err != nil {
		return err
	}
	return os.WriteFile(stamp, []byte(version), 0o644)
}

// sh runs a command in dir, its output going to log.
func sh(log io.Writer, dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}
