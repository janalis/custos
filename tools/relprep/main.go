// Command relprep prepares a release: it resolves the new version (explicit
// X.Y.Z or a patch/minor/major bump of the latest vX.Y.Z tag), turns the
// CHANGELOG's [Unreleased] section into the version's section, writes that
// section as release notes and pins the version in the Composer launcher.
// It prints `version=` and `tag=` lines for $GITHUB_OUTPUT.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("relprep", flag.ContinueOnError)
	fs.SetOutput(stderr)
	bump := fs.String("bump", "", "new version X.Y.Z, or patch | minor | major")
	date := fs.String("date", time.Now().UTC().Format(time.DateOnly), "release date (YYYY-MM-DD)")
	notes := fs.String("notes", "", "write the release notes to this file")
	changelog := fs.String("changelog", "CHANGELOG.md", "changelog to update")
	launcher := fs.String("launcher", "composer/custos", "Composer launcher whose VERSION is pinned")
	dir := fs.String("git", ".", "git repository holding the release tags")
	if err := fs.Parse(args); err != nil || *bump == "" || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: relprep -bump X.Y.Z|patch|minor|major [-date YYYY-MM-DD] [-notes file] [-changelog file] [-launcher file] [-git dir]")
		return 2
	}
	version, err := prepare(*bump, *date, *notes, *changelog, *launcher, *dir)
	if err != nil {
		fmt.Fprintln(stderr, "relprep:", err)
		return 1
	}
	fmt.Fprintf(stdout, "version=%s\ntag=v%s\n", version, version)
	return 0
}

func prepare(bump, date, notes, changelog, launcher, dir string) (string, error) {
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return "", fmt.Errorf("bad date %q", date)
	}
	tags, err := gitTags(dir)
	if err != nil {
		return "", err
	}
	version, err := nextVersion(bump, tags)
	if err != nil {
		return "", err
	}
	cl, err := os.ReadFile(changelog)
	if err != nil {
		return "", err
	}
	newCL, body, err := cutChangelog(string(cl), version, date)
	if err != nil {
		return "", err
	}
	lc, err := os.ReadFile(launcher)
	if err != nil {
		return "", err
	}
	newLC, err := pinLauncher(string(lc), version)
	if err != nil {
		return "", err
	}
	if notes != "" {
		if err := os.WriteFile(notes, []byte(body+"\n"), 0o644); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(changelog, []byte(newCL), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(launcher, []byte(newLC), 0o755); err != nil {
		return "", err
	}
	return version, nil
}

// semver is a release version X.Y.Z (prereleases are not cut by relprep).
type semver [3]int

var semverRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

func parseSemver(s string) (semver, bool) {
	m := semverRe.FindStringSubmatch(s)
	if m == nil {
		return semver{}, false
	}
	var v semver
	for i := range v {
		n, err := strconv.Atoi(m[i+1])
		if err != nil || (len(m[i+1]) > 1 && m[i+1][0] == '0') {
			return semver{}, false
		}
		v[i] = n
	}
	return v, true
}

func (v semver) String() string { return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2]) }

func (v semver) less(w semver) bool {
	for i := range v {
		if v[i] != w[i] {
			return v[i] < w[i]
		}
	}
	return false
}

func gitTags(dir string) ([]string, error) {
	out, err := exec.Command("git", "-C", dir, "tag", "--list", "v*").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("git tag: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	return strings.Fields(string(out)), nil
}

// nextVersion resolves bump against the highest vX.Y.Z tag (0.0.0 when there
// is none); an explicit version must be greater than it.
func nextVersion(bump string, tags []string) (string, error) {
	var latest semver
	for _, t := range tags {
		if v, ok := parseSemver(strings.TrimPrefix(t, "v")); ok && latest.less(v) {
			latest = v
		}
	}
	next := latest
	switch bump {
	case "major":
		next = semver{latest[0] + 1, 0, 0}
	case "minor":
		next = semver{latest[0], latest[1] + 1, 0}
	case "patch":
		next = semver{latest[0], latest[1], latest[2] + 1}
	default:
		v, ok := parseSemver(strings.TrimPrefix(bump, "v"))
		if !ok {
			return "", fmt.Errorf("bad version %q (want X.Y.Z, patch, minor or major)", bump)
		}
		if !latest.less(v) {
			return "", fmt.Errorf("version %s is not greater than the latest tag v%s", v, latest)
		}
		next = v
	}
	return next.String(), nil
}

const unreleased = "## [Unreleased]"

// cutChangelog renames the [Unreleased] section to the version's section and
// opens a fresh, empty [Unreleased] above it; it returns the new changelog and
// the released section's body.
func cutChangelog(cl, version, date string) (string, string, error) {
	start := strings.Index(cl, "\n"+unreleased+"\n")
	if start < 0 {
		return "", "", errors.New("changelog has no \"" + unreleased + "\" heading")
	}
	start++ // the heading line
	bodyStart := start + len(unreleased) + 1
	end := len(cl)
	if i := strings.Index(cl[bodyStart:], "\n## "); i >= 0 {
		end = bodyStart + i + 1
	}
	body := strings.TrimSpace(cl[bodyStart:end])
	if body == "" {
		return "", "", errors.New("the [Unreleased] changelog section is empty")
	}
	heading := fmt.Sprintf("## [%s] - %s", version, date)
	return cl[:start] + unreleased + "\n\n" + heading + "\n" + cl[bodyStart:], body, nil
}

var launcherVersionRe = regexp.MustCompile(`const VERSION = '[^']*';`)

func pinLauncher(src, version string) (string, error) {
	if n := len(launcherVersionRe.FindAllStringIndex(src, -1)); n != 1 {
		return "", fmt.Errorf("launcher: want one `const VERSION = '…';`, found %d", n)
	}
	return launcherVersionRe.ReplaceAllLiteralString(src, "const VERSION = '"+version+"';"), nil
}
