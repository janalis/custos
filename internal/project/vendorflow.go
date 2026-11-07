package project

import (
	"path/filepath"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// WithVendorFlow returns options whose snapshots answer the contracts of
// dependency callables by summarizing, on demand, the files under the vendor
// directory of root that declare them.
func WithVendorFlow(root string, ix *index.Index, parse syntax.Options, options flow.Options) flow.Options {
	// Both sides are absolute (see absolute): a dependency is any file below.
	dir := absolute(filepath.Join(root, vendorDir(root)))
	within := func(path string) bool {
		return dir != "" && strings.HasPrefix(absolute(path), dir+string(filepath.Separator))
	}
	read := func(path string) ([]byte, error) { return safeio.ReadFile(path, flow.MaxVendorBytes) }
	options.Vendor = flow.NewVendor(ix, parse, options.Assertions, within, read)
	return options
}
