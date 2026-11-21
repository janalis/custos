package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/fixing"
	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// oldTargets are the PHP versions every own fixture is re-run at.
var oldTargets = []phpversion.Version{phpversion.PHP53, phpversion.PHP56, phpversion.PHP70}

// TestOwnFixesRespectTargetVersion runs every rule on its own fixtures at
// old targets and applies each offered fix alone, without the fix loop's
// version guard: no fix may make the file need a newer PHP version than
// both the target and the original file.
func TestOwnFixesRespectTargetVersion(t *testing.T) {
	a := engine.(*adapter)
	rules, err := meta.All()
	if err != nil {
		t.Fatal(err)
	}
	applied := 0
	for _, r := range rules {
		rule := r.ID
		if !r.HasFix || !wanted([]string{rule}) {
			continue
		}
		dir := filepath.Join(ownRoot, strings.ToLower(rule), "testdata")
		files, _ := filepath.Glob(filepath.Join(dir, "*.php"))
		for _, f := range files {
			if strings.HasSuffix(f, ".fixed.php") {
				continue
			}
			req, marked, _ := ownCase(t, rule, dir, f, ".json")
			clean, _, err := ParseMarkup(marked)
			if err != nil {
				t.Fatal(err)
			}
			req.Source = clean
			tree := syntax.Parse(f, clean, parseOpts())
			had, _ := syntax.RequiredVersion(tree)
			for _, v := range oldTargets {
				req.PHP = v.String()
				e, err := a.engineFor(req)
				if err != nil {
					t.Fatal(err)
				}
				limit := max(v, had)
				for _, fd := range e.Analyze(tree) {
					for _, fx := range fd.Fixes {
						edits, ok := fixing.Edits(fx)
						if !ok {
							continue
						}
						out, _ := fixing.Apply(clean, edits)
						applied++
						if need, at := syntax.RequiredVersion(syntax.Parse(f, out, parseOpts())); need > limit {
							sp := at.Span()
							t.Errorf("%s/%s at PHP %s: fix %q of %s needs PHP %s for %q", rule, filepath.Base(f), v, fx.Title, fd.Rule, need, out[sp.Start:sp.End])
						}
					}
				}
			}
		}
	}
	if applied == 0 && wanted([]string{"ReturnTypeCanBeDeclared"}) {
		t.Fatal("no fix applied")
	}
	t.Logf("%d fixes checked", applied)
}

// TestOwnFixturesFormatAtTargetVersion formats every own fixture with every
// preset (risky rewrites included) below each formatter gate (5.4, 7.1, 7.3,
// 8.0): the formatter must never refuse its own output for needing a newer
// PHP version, which would mean a rewrite is not gated on the target.
func TestOwnFixturesFormatAtTargetVersion(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join(ownRoot, "*", "testdata", "*.php"))
	var names []string
	var sources [][]byte
	for _, f := range files {
		if strings.HasSuffix(f, ".fixed.php") {
			continue
		}
		marked, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		clean, _, err := ParseMarkup(marked)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, f)
		sources = append(sources, clean)
	}
	for _, name := range format.Presets() {
		r, err := format.Resolve(name, true)
		if err != nil {
			t.Fatal(err)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for i, src := range sources {
				for _, v := range []phpversion.Version{phpversion.PHP53, phpversion.PHP70} {
					res := format.Source("fixture.php", src, format.Options{PHP: v, Settings: &r.Settings})
					if strings.Contains(res.Skipped, "needs PHP") {
						t.Errorf("%s at PHP %s: %s", names[i], v, res.Skipped)
					}
				}
			}
		})
	}
}
