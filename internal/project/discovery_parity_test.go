package project

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/rules/nestednotoperators"
	"github.com/janalis/custos/internal/project/config"
)

// The command line walks the disk and the MCP server a captured root: both
// must index the same files, so a project reports the same findings
// through either. Every directory and file the policy prunes or keeps for
// declarations only is represented.
func TestDiscoveryParity(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{
		"src/A.php", "src/Gen/B.php", "src/.cache/C.php", "src/api.stub.php", "src/old.inc", "src/node_modules/N.php",
		"src/var/cache/V.php", "src/tests/T.php",
		"vendor/acme/lib/src/D.php", "vendor/acme/lib/src/test/E.php", "vendor/acme/lib/src/Tests/F.php",
		"vendor/acme/lib/tests/G.php", "vendor/acme/lib/node_modules/H.php", "vendor/acme/lib/.git/I.php",
		"vendor/acme/lib/vendor/J.php", "vendor/janalis/custos/bin/K.php", "vendor/acme/lib/src/x.stub.php",
	} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(f)), "<?php")
	}
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	scope := []string{root}
	files, err := Discover(scope, cfg.Exclude)
	if err != nil {
		t.Fatal(err)
	}
	indexed, _ := IndexSourcesSkipping(root, files, scope, cfg.Exclude)
	disk := map[string]bool{}
	for _, f := range indexed {
		rel, _ := filepath.Rel(root, f)
		disk[filepath.ToSlash(rel)] = true
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	s, err := CaptureSnapshot(context.Background(), r, []analysis.Rule{nestednotoperators.New()}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	captured := map[string]bool{}
	for _, f := range append(append([]string{}, s.OwnFiles...), s.Vendor...) {
		if strings.HasSuffix(f, ".php") {
			captured[f] = true
		}
	}
	keys := func(m map[string]bool) []string {
		return slices.Sorted(func(yield func(string) bool) {
			for k := range m {
				if !yield(k) {
					return
				}
			}
		})
	}
	if !slices.Equal(keys(disk), keys(captured)) {
		t.Fatalf("disk indexes %v\nsnapshot %v", keys(disk), keys(captured))
	}
}
