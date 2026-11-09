package selfupdate

import (
	"os"
	"path/filepath"
	"strings"
)

// Method is how the running binary was installed.
type Method int

const (
	// Standalone is a release binary placed on the PATH by hand.
	Standalone Method = iota
	// Source is a development build.
	Source
	// Homebrew is the janalis/tap/custos cask.
	Homebrew
	// Composer is the binary run by the janalis/custos Composer launcher.
	Composer
)

// Install describes the running binary.
type Install struct {
	Method Method
	// Path is the binary itself, symlinks resolved.
	Path string
	// Root is the Composer project requiring custos, when known.
	Root string
}

// executable and getenv are os.Executable and os.Getenv; tests replace them.
var (
	executable = os.Executable
	getenv     = os.Getenv
)

// Detect tells how the binary of the given version was installed. The
// Composer launcher announces itself with CUSTOS_INSTALL=composer and
// passes the project root in CUSTOS_COMPOSER_ROOT; Homebrew and launcher
// downloads are otherwise recognized by where the binary lives.
func Detect(version string) (Install, error) {
	if !Valid(version) {
		return Install{Method: Source}, nil
	}
	exe, err := executable()
	if err != nil {
		return Install{}, err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	in := Install{Method: Standalone, Path: exe}
	slashed := filepath.ToSlash(exe)
	switch {
	case getenv("CUSTOS_INSTALL") == "composer":
		in.Method, in.Root = Composer, getenv("CUSTOS_COMPOSER_ROOT")
	case strings.Contains(slashed, "/Caskroom/custos/"):
		in.Method = Homebrew
	case strings.Contains(slashed, "/janalis/custos/composer/.bin/custos-"):
		in.Method = Composer
	}
	return in, nil
}
