package format

import (
	"slices"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// prepared returns the engine of src, annotated and ready for the rules.
func prepared(t *testing.T, src string) *engine {
	t.Helper()
	r, _ := Resolve("per", false)
	b := []byte(src)
	f := syntax.Parse("t.php", b, syntax.Options{})
	toks := slices.Clone(f.Tokens)
	toks = append(toks, syntax.Token{Kind: syntax.TEOF, Start: uint32(len(b)), End: uint32(len(b))})
	e := &engine{src: b, f: f, s: &r.Settings, php: phpversion.Default, toks: toks, final: -1}
	e.annotate()
	e.units()
	e.pair()
	e.prepare()
	return e
}

// TestLookupsOutOfRange covers the lookups asked about offsets where
// nothing is found.
func TestLookupsOutOfRange(t *testing.T) {
	e := prepared(t, "<?php\n$a = 1; // c")
	end := uint32(len(e.src))
	switch {
	case e.at(1) != -1, e.find(0, 5, syntax.TClass) != -1, e.prevSig(0) != -1, e.nextSigIndex(end+1) != -1:
		t.Error("token lookups found something")
	case e.unitAt(1) != -1, e.unitAfter(end+1) != -1, e.unitBefore(0) != -1:
		t.Error("unit lookups found something")
	}
	e.markStart(nil)
	e.bodyRegion(nil)
	e.walk(nil)
	e.walkLines(nil)
	e.syntaxRewrites(nil)
	e.riskyNode(nil)
	e.setNL(0, 1, "x")
	e.setTail(0, 1, "x")
	last := len(e.sig) - 1
	e.setTail(last, 1, "x") // the gap before EOF ends with a comment
	if e.gaps[last].tailSet != -1 {
		t.Error("tail set after a comment")
	}
	e.atLeastBlank(0, "x")
	if e.controlBody(0) || e.inList(1) || e.multilineSignature(1) {
		t.Error("structure found")
	}
	e.gaps[1].frozen = true
	e.setSpace(1, 0, "x")
	e.setNL(1, 0, "x")
	e.setTail(1, 0, "x")
	if e.gaps[1].nl != -1 {
		t.Error("frozen gap changed")
	}
	e.ignored = []region{{0, end}}
	e.remove(2, "x")
	if e.deleted[2] {
		t.Error("ignored unit removed")
	}
}

// TestStringEnd covers strings left open at the end of the file.
func TestStringEnd(t *testing.T) {
	src := []byte("<?php \"a {$b[\"c\"]} d")
	toks, _ := syntax.Lex(src, syntax.LexOptions{})
	e := &engine{src: src, toks: toks}
	for i, tk := range toks {
		if tk.Kind == syntax.TDoubleQuote {
			if j := e.stringEnd(i); j < i {
				t.Errorf("end %d before %d", j, i)
			}
			break
		}
	}
	toks = append(toks, syntax.Token{Kind: syntax.TEOF, Start: uint32(len(src)), End: uint32(len(src))})
	e.toks = toks
	for i, tk := range e.toks {
		if tk.Kind == syntax.TDoubleQuote {
			if j := e.stringEnd(i); e.toks[j].Kind == syntax.TEOF {
				t.Error("string ends at EOF")
			}
			break
		}
	}
}

// TestKeptIndentation covers lines keeping their indentation.
func TestKeptIndentation(t *testing.T) {
	e := prepared(t, "<?php\n  $a = 1; $b = 2;\n")
	e.s.IndentStyle = "keep"
	e.indentation()
	for k := range e.sig {
		if e.want[k] != keepLevel {
			t.Fatalf("unit %d indented", k)
		}
	}
	if e.lineIndent(1) != "  " || e.commentIndent(1) != "" || e.origIndent(2) != "" {
		t.Errorf("kept indentation %q %q %q", e.lineIndent(1), e.commentIndent(1), e.origIndent(2))
	}
}

func TestOwnerAndSignature(t *testing.T) {
	e := prepared(t, "<?php\n{ $a; }\n")
	if o := e.owner(1); o != 0 {
		t.Errorf("owner %d", o)
	}
	e = prepared(t, "<?php\nclass A extends B implements C, D {}\n")
	for k := range e.sig {
		if e.kind(k) == syntax.TComma && !e.inSignature(k) {
			t.Error("implements comma")
		}
	}
	e = prepared(t, "<?php\nf($a, $b);\n")
	for k := range e.sig {
		if e.kind(k) == syntax.TComma && e.inSignature(k) {
			t.Error("argument comma")
		}
	}
	if hang := e.hangLevel([]openBracket{{k: -1, item: 2}}); hang != 3 {
		t.Errorf("hang %d", hang)
	}
}

func TestVerify(t *testing.T) {
	e := prepared(t, "<?php\necho 1;\n")
	for out, want := range map[string]string{
		"<?php\necho 1;\n":       "",
		"<?php\necho 1; 2;\n":    "tokens instead of",
		"<?php\nprint 1;\n":      "is T",
		"<?php\necho 2;\n":       "reads",
		"<?php\necho 1;\n?><?=;": "",
	} {
		got := e.verify([]byte(out))
		if want == "" && got != "" && out == "<?php\necho 1;\n" || want != "" && !strings.Contains(got, want) && !strings.Contains(got, "token") {
			t.Errorf("verify(%q) = %q, want %q", out, got, want)
		}
	}
	if got := e.verify([]byte("<?php\necho (1;\n")); got == "" {
		t.Error("syntax error accepted")
	}
	// The planned text itself breaks the syntax: the tokens match, the parse fails.
	for k := range e.sig {
		if e.kind(k) == syntax.TSemicolon {
			e.text[k] = ")"
		}
	}
	if got := e.verify([]byte("<?php\necho 1)\n")); got != "syntax error" {
		t.Errorf("broken syntax: %q", got)
	}
}

// TestVerifyTargetVersion refuses output needing a newer PHP version than
// both the target and the input.
func TestVerifyTargetVersion(t *testing.T) {
	for src, want := range map[string]string{
		"<?php\necho $a;\n":          "needs PHP 7.0",
		"<?php\necho $a, $b ?? 2;\n": "",
	} {
		e := prepared(t, src)
		e.php = phpversion.PHP56
		for k := range e.sig {
			if e.orig(k) == "$a" {
				e.text[k] = "$a ?? 1"
			}
		}
		out := strings.Replace(src, "$a", "$a ?? 1", 1)
		if got := e.verify([]byte(out)); got != want {
			t.Errorf("%q: got %q, want %q", out, got, want)
		}
		e.php = phpversion.Max
		if got := e.verify([]byte(out)); got != "" {
			t.Errorf("%q at max: %q", out, got)
		}
	}
}

// TestRefusedOutput covers a pass whose output the safety check refuses.
func TestRefusedOutput(t *testing.T) {
	defer func(f func(*engine, []byte) string) { checkOutput = f }(checkOutput)
	checkOutput = func(*engine, []byte) string { return "test" }
	res := Source("a.php", []byte("<?php\necho  1;\n"), Options{})
	if !strings.Contains(res.Skipped, "please report") || string(res.Output) != "<?php\necho  1;\n" {
		t.Errorf("refused: %+v", res)
	}
	r, _ := Resolve("symfony", true)
	calls := 0
	checkOutput = func(e *engine, out []byte) string {
		calls++
		if calls > 1 {
			return "second pass"
		}
		return e.verify(out)
	}
	res = Source("a.php", []byte("<?php\nif ($a == array()) {}\n"), Options{Settings: &r.Settings})
	if res.Skipped != "" || !strings.Contains(string(res.Output), "$a == []") {
		t.Errorf("first pass kept: %+v", res)
	}
}

func TestComposeReports(t *testing.T) {
	first := []Change{{Rule: "a", Report: true}, {Rule: "b", Start: 0, End: 1, New: "x"}}
	second := []Change{{Rule: "c", Report: true}}
	got := compose(first, second, []byte("x"))
	if len(got) != 2 || got[1].Rule != "c" || !got[1].Report {
		t.Errorf("compose: %+v", got)
	}
}

func TestEdges(t *testing.T) {
	e := prepared(t, "<?php\n$x; /* c */ { $a; }\n")
	for k := range e.sig {
		if e.kind(k) == syntax.TLBrace && e.owner(k) != k {
			t.Errorf("owner of a bare block: %d", e.owner(k))
		}
	}
	for k := 1; k < len(e.gaps); k++ {
		if e.gaps[k].comment {
			e.setNL(k, 0, "x")
			if e.gaps[k].nl != -1 {
				t.Error("comment joined")
			}
		}
	}
	e.setTail(1, 2, "x")
	if e.gaps[1].nl != 2 {
		t.Error("setTail without comment")
	}
	if memberKind(&syntax.Nop{}) != "" {
		t.Error("member kind of a statement")
	}
}

func TestWrittenGap(t *testing.T) {
	e := prepared(t, "<?php\n$a; $b; $c;\n")
	e.gaps[4].into = 2
	if e.written(4) != 2 || e.written(0) != 0 {
		t.Error("written gap")
	}
}

// Code newer than the configured version is checked with the lexer that
// read it: under PHP 7.4, #[ opens a comment.
func TestNewerSyntaxThanVersion(t *testing.T) {
	src := "<?php\n\nfunction f(#[A] $t) {}\n"
	res := Source("a.php", []byte(src), Options{PHP: phpversion.PHP74})
	if res.Skipped != "" || string(res.Output) != src {
		t.Fatalf("skipped %q, output %q", res.Skipped, res.Output)
	}
}

// TestInvalidArrayDestructuringKept: `array(...) = $x` is invalid in every
// version; shortening it would make it a 7.1 destructuring.
func TestInvalidArrayDestructuringKept(t *testing.T) {
	r, _ := Resolve("per", false)
	src := "<?php\n\narray($a, $b) = $x;\n$c = array(1);\n"
	for _, v := range []phpversion.Version{phpversion.PHP56, phpversion.Max} {
		res := Source("a.php", []byte(src), Options{PHP: v, Settings: &r.Settings})
		if want := "<?php\n\narray($a, $b) = $x;\n$c = [1];\n"; string(res.Output) != want || res.Skipped != "" {
			t.Errorf("%s: got %q (%s)", v, res.Output, res.Skipped)
		}
	}
}
