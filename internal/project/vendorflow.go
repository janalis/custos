package project

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/platform/safeio"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/index"
)

// WithVendorFlow returns options whose snapshots answer the contracts of
// dependency callables by summarizing, on demand, the files under the vendor
// directory of root that declare them.
func WithVendorFlow(root string, ix *index.Index, parse syntax.Options, options flow.Options) flow.Options {
	// Both sides are absolute (see absolute): a dependency is any file below.
	// The working directory is read once: filepath.Abs asks the system for
	// it on every relative path, and flow asks for each callable it meets.
	dir := absolute(filepath.Join(root, vendorDir(root)))
	cwd, _ := os.Getwd()
	within := func(path string) bool {
		return dir != "" && strings.HasPrefix(absoluteFrom(cwd, path), dir+string(filepath.Separator))
	}
	read := func(path string) ([]byte, error) { return safeio.ReadFile(path, flow.MaxVendorBytes) }
	options.Vendor = flow.NewVendor(ix, parse, options.Assertions, within, read)
	return options
}

// confinedVendorFlow is WithVendorFlow for a captured root: the index names
// its files relative to root (with forward slashes), and dependency files
// under the vendor directory vendor are read inside root, whatever the
// working directory.
func confinedVendorFlow(root *os.Root, vendor string, ix *index.Index, parse syntax.Options, options flow.Options) flow.Options {
	within := func(name string) bool { return vendor == "." || strings.HasPrefix(path.Clean(name), vendor+"/") }
	read := func(name string) ([]byte, error) { return ReadConfined(root, name, flow.MaxVendorBytes) }
	options.Vendor = flow.NewVendor(ix, parse, options.Assertions, within, read)
	return options
}
