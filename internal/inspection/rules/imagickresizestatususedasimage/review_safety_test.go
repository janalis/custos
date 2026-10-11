package imagickresizestatususedasimage

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
	const resize = "$r = $i->resizeImage(1, 1, 1, 1); "
	for _, tc := range []struct {
		name, body string
		findings   int
	}{
		{"plain", resize + "$r->getImageBlob();", 1},
		{"parenthesized", resize + "($r)->getImageBlob();", 1},
		{"chained", "$i->resizeImage(1, 1, 1, 1)->getImageBlob();", 1},
		{"foreach value", resize + "foreach ($l as $r) { $r->count(); }", 0},
		{"foreach key", resize + "foreach ($l as $r => $x) {} $r->count();", 0},
		{"foreach destructured", resize + "foreach ($l as [$r, $x]) {} $r->count();", 0},
		{"foreach unrelated", resize + "foreach ($l as $k => $x) {} $r->count();", 1},
		{"short destructuring", resize + "[$r, $o] = pair(); $r->count();", 0},
		{"list destructuring", resize + "list($o, $r) = pair(); $r->count();", 0},
		{"keyed destructuring", resize + "['a' => ['b' => $r]] = pair(); $r->count();", 0},
		{"unrelated destructuring", resize + "[$o, $p] = pair(); $r->count();", 1},
		{"reference alias", resize + "$a = &$r; $a = $i; $r->getNumberImages();", 0},
		{"alias before assignment", "$a = &$r; " + resize + "$a = $i; $r->getNumberImages();", 0},
		{"unrelated alias", resize + "$a = &$o; $a = $i; $r->getNumberImages();", 1},
		{"global", "global $r; " + resize + "g(); $r->count();", 0},
		{"static", "static $r; " + resize + "$r->count();", 0},
		{"unrelated static", "static $o; " + resize + "$r->count();", 1},
		{"closure reference", resize + "$f = function () use (&$r) { $r = new ArrayObject(); }; $r->count();", 0},
		{"closure value", resize + "$f = function () use ($r) { [$r] = pair(); }; $r->count();", 1},
		{"arrow", resize + "$f = fn () => [$r] = pair(); $r->count();", 1},
		{"catch", resize + "try { g(); } catch (Exception $r) {} $r->getMessage();", 0},
		{"catch unbound", resize + "try { g(); } catch (Exception) {} $r->getMessage();", 1},
		{"rebind before assignment", "foreach ($l as $r) {} [$r] = pair(); " + resize + "$r->getImageBlob();", 1},
		{"rebind after use", resize + "$r->getImageBlob(); [$r] = pair();", 1},
		{"closure after use", resize + "$r->getImageBlob(); $f = function () use (&$r) {};", 1},
		{"rebind later in loop", resize + "while (g()) { $r->count(); [$r] = pair(); }", 0},
		{"rebind later in outer loop", resize + "do { for (;;) { $r->count(); } [$r] = pair(); } while (g());", 0},
		{"rebind after loop", resize + "foreach ($l as $x) { $r->getImageBlob(); } [$r] = pair();", 1},
		{"by-reference builtin argument", resize + "preg_match('/a/', 'a', $r); $r->count();", 0},
		{"by-value builtin argument", resize + "var_export($r, true); $r->getImageBlob();", 1},
		{"copied", resize + "$b = $r; $b->getImageBlob();", 1},
		{"copied after rebind", resize + "[$r] = pair(); $b = ($r); $b->count();", 0},
		{"copy source rebound later", resize + "$b = $r; [$r] = pair(); $b->getImageBlob();", 1},
	} {
		for _, wrap := range [][2]string{{"<?php function f(Imagick $i, array $l) { ", " }"}, {"<?php $i = new Imagick(); ", ""}} {
			t.Run(tc.name, func(t *testing.T) {
				src := []byte(wrap[0] + tc.body + wrap[1])
				f := syntax.ParseBest("review.php", src, syntax.Options{Version: phpversion.PHP85})
				if len(f.Errors) != 0 {
					t.Fatalf("invalid probe: %v", f.Errors)
				}
				if got := e.Analyze(f); len(got) != tc.findings {
					t.Fatalf("%s\nfindings: got %v, want %d", src, got, tc.findings)
				}
			})
		}
	}
}
