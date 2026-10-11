package zipentrynamecollision

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestReviewSafety(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	const open = `<?php function make(): ZipArchive { return new ZipArchive(); } class Holder { public ZipArchive $zip; public ZipArchive $other; public static ZipArchive $shared; public int $count = 0; function log($v) {} function probe(Holder $o, Holder $p, array $list) { $z = new ZipArchive(); $z->open("a.zip", ZipArchive::CREATE); `
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"collision", `if ($z->addFromString("n", "one")) { $z->addFromString("n", "two"); }`, 1},
		{"alias of the same archive", `if ($z->addFromString("n", "one")) { $y = $z; $y->addFromString("n", "two"); }`, 1},
		{"reopened before the guard", `$z->close(); $z->open("a.zip"); if ($z->addFromString("n", "one")) { $z->addFromString("n", "two"); }`, 1},
		{"conditional removal", `if ($z->addFromString("n", "one")) { if (rand()) { $z->deleteIndex(0); } $z->addFromString("n", "two"); }`, 1},
		{"renamed by name", `if ($z->addFromString("n", "one")) { $z->renameName("n", "old"); $z->addFromString("n", "two"); }`, 0},
		{"renamed by index", `if ($z->addFromString("n", "one")) { $z->renameIndex(0, "old"); $z->addFromString("n", "two"); }`, 0},
		{"deleted by name", `if ($z->addFromString("n", "one")) { $z->deleteName("n"); $z->addFromString("n", "two"); }`, 0},
		{"deleted by index", `if ($z->addFromString("n", "one")) { $z->deleteIndex(0); $z->addFromString("n", "two"); }`, 0},
		{"reverted archive", `if ($z->addFromString("n", "one")) { $z->unchangeAll(); $z->addFromString("n", "two"); }`, 0},
		{"reverted entry", `if ($z->addFromString("n", "one")) { $z->unchangeName("n"); $z->addFromString("n", "two"); }`, 0},
		{"closed and reopened", `if ($z->addFromString("n", "one")) { $z->close(); $z->open("b.zip"); $z->addFromString("n", "two"); }`, 0},
		{"receiver rebound", `if ($z->addFromString("n", "one")) { $first = $z; $z = new ZipArchive(); $z->open("b.zip", ZipArchive::CREATE); $z->addFromString("n", "two"); }`, 0},
		{"archive escaped to a call", `if ($z->addFromString("n", "one")) { helper($z); $z->addFromString("n", "two"); }`, 0},
		{"other archive", `$y = new ZipArchive(); if ($y->addFromString("n", "one")) { $z->addFromString("n", "two"); }`, 0},
		{"other entry name", `if ($z->addFromString("m", "one")) { $z->addFromString("n", "two"); }`, 0},
		{"call result archive", `if (make()->addFromString("n", "one")) { make()->addFromString("n", "two"); }`, 0},
		{"insertion in a callback", `if ($z->addFromString("n", "one")) { $later = function () use ($z) { $z->addFromString("n", "two"); }; }`, 0},
		{"factory archive", `$f = make(); if ($f->addFromString("n", "one")) { echo strlen("x"); $f->addFromString("n", "two"); }`, 1},
		{"property archive", `if ($this->zip->addFromString("n", "one")) { $this->count = 1; $this->zip->addFromString("n", strtoupper("two")); }`, 1},
		{"nested property archive", `if ($o->zip->addFromString("n", "one")) { $o->zip->addFromString("n", "two"); }`, 1},
		{"static property archive", `if (self::$shared->addFromString("n", "one")) { if (rand()) { echo 1; } self::$shared->addFromString("n", "two"); }`, 1},
		{"property archive wrapped in a call", `if ($this->zip->addFromString("n", "one")) { $this->log($this->zip->addFromString("n", "two")); }`, 1},
		{"other property", `if ($this->zip->addFromString("n", "one")) { $this->other->addFromString("n", "two"); }`, 0},
		{"factory and known archive", `$f = make(); if ($f->addFromString("n", "one")) { $z->addFromString("n", "two"); }`, 0},
		{"property entry deleted", `if ($this->zip->addFromString("n", "one")) { $this->zip->deleteName("n"); $this->zip->addFromString("n", "two"); }`, 0},
		{"property reassigned", `if ($this->zip->addFromString("n", "one")) { $this->zip = $z; $this->zip->addFromString("n", "two"); }`, 0},
		{"property owner rebound", `if ($o->zip->addFromString("n", "one")) { $o = $p; $o->zip->addFromString("n", "two"); }`, 0},
		{"property reset by a function", `if ($this->zip->addFromString("n", "one")) { reset_archive($this); $this->zip->addFromString("n", "two"); }`, 0},
		{"property reset by a static call", `if ($this->zip->addFromString("n", "one")) { Registry::reset(); $this->zip->addFromString("n", "two"); }`, 0},
		{"property with a new object between", `if ($this->zip->addFromString("n", "one")) { $x = new Resetter($this); $this->zip->addFromString("n", "two"); }`, 0},
		{"property with a callback between", `if ($this->zip->addFromString("n", "one")) { $x = fn() => 1; $this->zip->addFromString("n", "two"); }`, 0},
		{"array element archive", `if ($list[0]->zip->addFromString("n", "one")) { $list[0]->zip->addFromString("n", "two"); }`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.ParseBest("review.php", []byte(open+tc.body+" } }"), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
}
