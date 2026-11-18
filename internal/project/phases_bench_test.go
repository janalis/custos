package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/infer"
)

// benchProject writes a synthetic project: n source files whose classes
// call one another, functions of other files and builtins, and a vendor
// directory of n/2 library classes. It returns the root and the sources.
func benchProject(b *testing.B, n int) (root string, files []string) {
	b.Helper()
	root = b.TempDir()
	write := func(rel, src string) string {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			b.Fatal(err)
		}
		// Older than racyWindow, so caches keep them.
		old := time.Now().Add(-time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			b.Fatal(err)
		}
		return path
	}
	write("composer.json", `{"name":"bench/app"}`)
	for i := range n / 2 {
		write(fmt.Sprintf("vendor/lib/src/Lib%d.php", i), fmt.Sprintf(`<?php
namespace Lib;
class Lib%d {
    public function get(string $key, ?array $default = null): ?array { return $default; }
    public static function make(): static { return new static(); }
}
`, i))
	}
	for i := range n {
		var body strings.Builder
		fmt.Fprintf(&body, `<?php
namespace App;

use Lib\Lib%d;

function helper%d(string $s): string { return trim(strtolower($s)); }

final class Service%d
{
    private array $items = [];

    public function __construct(private Lib%d $lib) {}

    public function run(array $input): array
    {
        $out = [];
        foreach ($input as $k => $v) {
            if (is_string($v) && strlen($v) > 0) {
                $out[$k] = helper%d($v) . helper%d(str_replace('a', 'b', $v));
            } elseif (is_array($v)) {
                $out = array_merge($out, array_map('strval', $v));
            }
        }
        $cfg = $this->lib->get('k%d') ?? [];
        if (count($cfg) > 0 && in_array('x', $cfg)) {
            $this->items[] = sprintf('%%s-%%d', implode(',', $cfg), count($out));
        }
        return $out;
    }
}
`, i%max(n/2, 1), i, i, i%max(n/2, 1), i, (i+1)%n, i)
		files = append(files, write(fmt.Sprintf("src/Service%d.php", i), body.String()))
	}
	return root, files
}

func benchOpen(b *testing.B, root, cache string) *Project {
	b.Helper()
	p, err := Open(catalogue.All(), Options{
		Root: root, Paths: []string{filepath.Join(root, "src")}, Cache: cache,
		Analysis: analysis.Config{PHP: phpversion.PHP84},
		Parse:    syntax.Options{Version: phpversion.PHP84},
	})
	if err != nil {
		b.Fatal(err)
	}
	return p
}

// BenchmarkAnalyzeReport runs every phase on a synthetic project, without
// a cache, and with the cache an earlier run left (declarations of vendor
// files, flow summaries).
func BenchmarkAnalyzeReport(b *testing.B) {
	root, _ := benchProject(b, 200)
	b.Run("cold", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchOpen(b, root, "").AnalyzeReport()
		}
	})
	b.Run("warm", func(b *testing.B) {
		cache := b.TempDir()
		benchOpen(b, root, cache).AnalyzeReport()
		b.ReportAllocs()
		for b.Loop() {
			benchOpen(b, root, cache).AnalyzeReport()
		}
	})
}

// BenchmarkIndexPhase indexes the synthetic project's sources.
func BenchmarkIndexPhase(b *testing.B) {
	_, files := benchProject(b, 200)
	opt := syntax.Options{Version: phpversion.PHP84}
	b.ReportAllocs()
	for b.Loop() {
		BuildIndexUnread(context.Background(), files, opt, infer.Options{}, nil)
	}
}

// BenchmarkSymbolCache opens a cache of the project's vendor declarations
// and saves it again after one miss.
func BenchmarkSymbolCache(b *testing.B) {
	root, _ := benchProject(b, 400)
	vendor, err := Discover([]string{filepath.Join(root, "vendor")}, nil)
	if err != nil {
		b.Fatal(err)
	}
	dir := b.TempDir()
	opt := syntax.Options{Version: phpversion.PHP84}
	BuildIndexUnread(context.Background(), vendor, opt, infer.Options{}, OpenSymbolCache(dir, root, opt, infer.Options{}))
	b.Run("open", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			OpenSymbolCache(dir, root, opt, infer.Options{})
		}
	})
	b.Run("save", func(b *testing.B) {
		c := OpenSymbolCache(dir, root, opt, infer.Options{})
		BuildIndexUnread(context.Background(), vendor, opt, infer.Options{}, c) // every entry used
		b.ReportAllocs()
		for b.Loop() {
			c.missed = true
			if err := c.Save(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkDispatchCalls analyzes a file made of function calls with every
// rule enabled: most rules subscribe to calls.
func BenchmarkDispatchCalls(b *testing.B) {
	engine, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php\nfunction f(array $a, string $s) {\n" + strings.Repeat("  $x = strlen($s) + count($a); $y = str_replace('a', 'b', $s); $z = array_map('trim', $a); in_array($s, $a);\n", 50) + "}\n")
	opt := syntax.Options{Version: phpversion.PHP85}
	AnalyzeBuffer(engine, "calls.php", src, opt) // warm embedded symbols
	b.ReportAllocs()
	for b.Loop() {
		AnalyzeBuffer(engine, "calls.php", src, opt)
	}
}
