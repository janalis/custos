package zipsourceremovedbeforearchiveclose

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
	const open = `<?php function probe($flag, ZipArchive $given) { $z = new ZipArchive(); $z->open("b.zip", ZipArchive::CREATE); if ($z->addFile("s.txt", "n.txt")) { `
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"removed first", `unlink("s.txt"); $z->close();`, 1},
		{"close statement", `$z->close(); unlink("s.txt");`, 0},
		{"negated guard", `if (!$z->close()) { return; } unlink("s.txt");`, 0},
		{"positive guard", `if ($z->close()) { unlink("s.txt"); }`, 0},
		{"failed close branch", `if ($z->close()) { echo 1; } else { unlink("s.txt"); }`, 0},
		{"try finally", `try { $z->close(); } finally { unlink("s.txt"); }`, 0},
		{"after try", `try { $z->close(); } catch (Exception $e) { echo 1; } unlink("s.txt");`, 0},
		{"close in finally", `try { echo 1; } finally { $z->close(); } unlink("s.txt");`, 0},
		{"removal in catch", `try { $z->close(); } catch (Exception $e) { unlink("s.txt"); }`, 1},
		{"removal in catch with finally", `try { $z->close(); } catch (Exception $e) { unlink("s.txt"); } finally { echo 1; }`, 1},
		{"close in catch", `try { echo 1; } catch (Exception $e) { $z->close(); } unlink("s.txt");`, 1},
		{"close in branch", `if ($flag) { $z->close(); } unlink("s.txt");`, 1},
		{"close in unbraced branch", `if ($flag) $z->close(); unlink("s.txt");`, 1},
		{"close in else", `if ($flag) { echo 1; } else { $z->close(); } unlink("s.txt");`, 1},
		{"close in loop", `while ($flag) { $z->close(); } unlink("s.txt");`, 1},
		{"short circuit close", `$flag && $z->close(); unlink("s.txt");`, 1},
		{"short circuit chain", `$z->close() && unlink("s.txt");`, 0},
		{"left operand close", `$done = $z->close() && $flag; unlink("s.txt");`, 0},
		{"eager right operand close", `$done = $flag & $z->close(); unlink("s.txt");`, 0},
		{"ternary arm close", `$flag ? $z->close() : null; unlink("s.txt");`, 1},
		{"ternary condition close", `$done = $z->close() ? 1 : 0; unlink("s.txt");`, 0},
		{"match arm close", `match (true) { $flag => $z->close(), default => null }; unlink("s.txt");`, 1},
		{"close in callback", `$later = function () use ($z) { $z->close(); }; unlink("s.txt");`, 1},
		{"close in arrow callback", `$later = fn() => $z->close(); unlink("s.txt");`, 1},
		{"removal in callback", `$later = fn() => unlink("s.txt"); $z->close(); $later();`, 0},
		{"removal in closure", `$later = function () { unlink("s.txt"); }; $z->close(); $later();`, 0},
		{"other archive closed", `$given->close(); unlink("s.txt");`, 1},
		{"function named close", `close($z); unlink("s.txt");`, 1},
		{"close after removal", `unlink("s.txt"); if ($z->close()) { echo 1; }`, 1},
		{"close after a protected call, removal in finally", `try { risky(); $z->close(); } finally { unlink("s.txt"); }`, 1},
		{"close after a swallowed call", `try { risky(); $z->close(); } catch (Exception $e) { echo 1; } unlink("s.txt");`, 1},
		{"close after a propagating call", `try { risky(); $z->close(); } finally { echo 1; } unlink("s.txt");`, 0},
		{"guarded close first in try", `try { if (!$z->close()) { return; } risky(); } catch (Exception $e) { echo 1; } unlink("s.txt");`, 0},
		{"null coalescing assignment close", `$flag ??= $z->close(); unlink("s.txt");`, 1},
		{"assigned close", `$flag = $z->close(); unlink("s.txt");`, 0},
		{"elseif close", `if ($flag) { return; } elseif ($z->close()) { unlink("s.txt"); }`, 0},
		{"elseif close may be skipped", `if ($flag) { echo 1; } elseif ($z->close()) { echo 2; } unlink("s.txt");`, 1},
		{"close in elseif body", `if ($flag) { echo 1; } elseif ($flag > 1) { $z->close(); } unlink("s.txt");`, 1},
		{"rebound archive", `$z = $given; $z->close(); unlink("s.txt");`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(open + tc.body + ` } }`)
			opt := syntax.Options{Version: phpversion.PHP85}
			f := syntax.ParseBest("review.php", src, opt)
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: got %v, want %d", got, tc.findings)
			}
		})
	}
	// Unknown receivers and top-level code keep the existing behaviour; archives
	// without a flow identity are matched by spelling.
	const holder = `<?php function make(): ZipArchive { return new ZipArchive(); } class Holder { public ZipArchive $zip; public ZipArchive $other; public static ZipArchive $shared; function probe(Holder $o, Holder $p) { `
	for source, want := range map[string]int{
		`<?php function make(): ZipArchive { return new ZipArchive(); } $z = make(); if ($z->addFile("s.txt")) { $z->close(); unlink("s.txt"); }`: 0,
		`<?php function make(): ZipArchive { return new ZipArchive(); } $z = make(); if ($z->addFile("s.txt")) { unlink("s.txt"); $z->close(); }`: 1,
		holder + `if ($this->zip->addFile("s.txt")) { $this->zip->close(); unlink("s.txt"); } } }`:                                                0,
		holder + `if ($this->zip->addFile("s.txt")) { if (!$this->zip->close()) { return; } unlink("s.txt"); } } }`:                               0,
		holder + `if (self::$shared->addFile("s.txt")) { self::$shared->close(); unlink("s.txt"); } } }`:                                          0,
		holder + `if ($o->zip->addFile("s.txt")) { $o->zip->close(); unlink("s.txt"); } } }`:                                                      0,
		holder + `if ($this->zip->addFile("s.txt")) { unlink("s.txt"); $this->zip->close(); } } }`:                                                1,
		holder + `if ($this->zip->addFile("s.txt")) { $this->other->close(); unlink("s.txt"); } } }`:                                              1,
		holder + `if ($this->zip->addFile("s.txt")) { $this->zip = make(); $this->zip->close(); unlink("s.txt"); } } }`:                           1,
		holder + `if ($o->zip->addFile("s.txt")) { $o = $p; $o->zip->close(); unlink("s.txt"); } } }`:                                             1,
		holder + `$z = new ZipArchive(); if ($this->zip->addFile("s.txt")) { $z->close(); unlink("s.txt"); } } }`:                                 1,
		`<?php function p($f) { $z = new ZipArchive(); if ($f) { helper($z); } if ($z->addFile("s.txt")) { $z->close(); unlink("s.txt"); } }`:     0,
		holder + `if (make()->addFile("s.txt")) { make()->close(); unlink("s.txt"); } } }`:                                                        1,
		`<?php if ($z->addFile("s.txt")) { $z->close(); unlink("s.txt"); }`:                                                                       0,
		`<?php $z = new ZipArchive(); if ($z->addFile("s.txt")) { if (!$z->close()) { exit(1); } unlink("s.txt"); }`:                              0,
		`<?php $z = new ZipArchive(); $z->close(); if ($z->addFile("s.txt")) { unlink("s.txt"); }`:                                                0,
		`<?php $z = new ZipArchive(); $f = function () use ($z) { if ($z->addFile("s.txt")) { unlink("s.txt"); } };`:                              1,
		`<?php function p(ZipArchive $z) { if ($z->addFile("s.txt")) { return fn() => unlink("s.txt"); } }`:                                       0,
	} {
		f := syntax.ParseBest("review.php", []byte(source), syntax.Options{Version: phpversion.PHP85})
		if len(f.Errors) != 0 {
			t.Fatalf("invalid probe %s: %v", source, f.Errors)
		}
		if got := e.Analyze(f); len(got) != want {
			t.Fatalf("%s: got %v, want %d", source, got, want)
		}
	}
}
