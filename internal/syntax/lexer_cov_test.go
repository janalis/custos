package syntax

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/phpver"
)

// lexSnippets are edge cases of the scanner. Each one is lexed with PHP 8.5
// semantics and, when php is installed, compared token by token (kind and
// range) with token_get_all.
var lexSnippets = []string{
	// open / close tags and their newline
	"<?php\r\n$a;", "<?php\r$a;", "<?php\t$a;", "<?php $a ?>\r\nx", "<?php $a ?>\rx", "<?php $a ?>\nx",
	// comments
	"<?php $a /= 2;", "<?php // c ?> x", "<?php # c\r\n$a;", "<?php /* open", "<?php /** doc */ $a;",
	// binary string prefixes
	"<?php b'x'; B\"y\"; b\"$z\";", "<?php b<<<EOT\nx\nEOT;\n", "<?php b<<<1;", "<?php $b<<1;",
	// contextual keywords
	"<?php readonly(1);", "<?php readonly ( 1 );", "<?php readonly class A {}", "<?php readonly /* c */ (1);",
	"<?php enum /* c */ A {}", "<?php enum // c\n A {}", "<?php enum # c\n A {}", "<?php enum /* open",
	"<?php enum ", "<?php enum #[A] class B {}",
	// asymmetric visibility
	"<?php class A { private( set ) int $a; }", "<?php class A { public(\tset\t) int $a; }",
	"<?php class A { private(get) int $a; }", "<?php class A { private( set", "<?php class A { private( set x",
	"<?php class A { private( se", "<?php class A { private(",
	// numbers
	"<?php 0x7FFFFFFFFFFFFFFFF;", "<?php 0x7FFF_FFFF_FFFF_FFFF;", "<?php 0b11111111111111111111111111111111111111111111111111111111111111111;",
	"<?php 0o7777777777777777777777;", "<?php 0O777777777777777;", "<?php 07777777777777777777777;", "<?php 0777777777777777777;",
	"<?php 99999999999999999999;", "<?php 1.; 1..2; .5e3; 1e+5; 1e-5; 1e; 1_000;", "<?php 0xZ; 0b2; 0o8;",
	// strings
	"<?php \"$a\\", `<?php "${$x}";`, `<?php "${ a}";`, `<?php "${a}";`, `<?php "${a[0]}";`,
	`<?php "$a?->b";`, `<?php "$a->b->c";`, `<?php "$a[-1] $a[b] $a[$c]";`, "<?php `ls $a`;",
	`<?php "{$a}";`, `<?php "\{$a}";`, `<?php "a\"b";`,
	// heredoc / nowdoc
	"<?php <<< EOT\nx\nEOT;\n", "<?php <<<\t\"EOT\"\nx\nEOT;\n", "<?php <<<'EOT'\nx $a\nEOT;\n",
	"<?php <<<EOT\r\nx\r\nEOT;\r\n", "<?php <<<EOT\rx\rEOT;\r", "<?php <<<EOT x\nEOT;\n",
	"<?php <<<EOT\nEOTX\nEOT;\n", "<?php <<<EOT\n  x\n  EOT;\n", "<?php <<<EOT\nx $a y\nEOT;\n", "<?php <<<EOT\nx\n",
	"<?php <<<EOT\nx\\", "<?php <<<EOT", "<?php <<<EOT\n\\$a {$b} ${c}\nEOT;\n", "<?php <<<\nx",
	// member names after comments
	"<?php $o->/*c*/list; $o->#c\nlist; $o?->//c\nclass; $o->#[x]\nlist; $o->/**/ 1;",
	// operators and odd bytes
	"<?php $a |> f(...);", "<?php \x01 $a;", "<?php $a <=> $b ** 2 ?? 3; $a ??= 1; $a?->b; $a::c; \\A\\b; \\ ;",
	"<?php $$a; $ ;", "<?php __halt_compiler(); raw", "<?php __halt_compiler() ?> raw", "<?php __halt_compiler",
	"<?php (int) (integer)(bool)(boolean)(float)(double)(real)(string)(binary)(array)(object)(unset)(void) (abcdefgh) ( int",
	"<?php yield\nfrom $a; yield fromage; yield from", "<?php namespace\\a; a\\b;",
}

// lexErrorSnippets are invalid inputs on which PHP's recovery tokens differ
// from ours (both report the error): only coverage, contiguity and the
// error are checked.
var lexErrorSnippets = []string{
	`<?php "abc`, `<?php 'abc`, `<?php "$a[!]";`, "<?php <<<\"EOT\nx\n", "<?php <<<'EOT\nx\n", "<?php /* open",
	"<?php <<<EOT\nx\n",
}

func TestLexErrorSnippets(t *testing.T) {
	for _, s := range lexErrorSnippets {
		toks, errs := Lex([]byte(s), LexOptions{Version: phpver.PHP85})
		if len(errs) == 0 {
			t.Errorf("%q: no lexical error", s)
		}
		if len(toks) == 0 || int(toks[len(toks)-1].End) != len(s) {
			t.Errorf("%q: tokens do not cover the input", s)
		}
	}
}

func TestLexSnippetsMatchPHP(t *testing.T) {
	dir := t.TempDir()
	var files []string
	got := map[string][]Token{}
	for i, s := range lexSnippets {
		toks, _ := Lex([]byte(s), LexOptions{Version: phpver.PHP85})
		var end uint32
		for _, tk := range toks {
			if tk.Start != end {
				t.Errorf("%q: tokens not contiguous at %d", s, tk.Start)
			}
			end = tk.End
		}
		if int(end) != len(s) {
			t.Errorf("%q: tokens cover %d of %d bytes", s, end, len(s))
		}
		p := filepath.Join(dir, strings.Repeat("0", 3-len(itoa(i)))+itoa(i)+".php")
		if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
		files = append(files, p)
		got[p] = toks
	}
	if _, err := exec.LookPath("php"); err != nil {
		t.Skip("php not installed: only contiguity checked")
	}
	oracle := runOracle(t, files)
	for i, p := range files {
		want, ok := oracle[p]
		if !ok {
			t.Errorf("%q: no oracle output", lexSnippets[i])
			continue
		}
		if msg := compareTokens([]byte(lexSnippets[i]), got[p], want); msg != "" {
			t.Errorf("%q: %s", lexSnippets[i], msg)
		}
	}
}

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

func TestLexVersionEdges(t *testing.T) {
	cases := []struct {
		src  string
		ver  phpver.Version
		want []TokenKind
	}{
		{"<?php $a |> $b", phpver.PHP84, []TokenKind{TOpenTag, TVariable, TBar, TGreater, TVariable}},
		{"<?php readonly (1)", phpver.PHP81, []TokenKind{TOpenTag, TString, TLParen, TLNumber, TRParen}},
		{"<?php readonly /* c */ (1)", phpver.PHP81, []TokenKind{TOpenTag, TReadonly, TLParen, TLNumber, TRParen}},
		{"<?php readonly", phpver.PHP81, []TokenKind{TOpenTag, TReadonly}},
		{"<?php readonly(1)", phpver.PHP82, []TokenKind{TOpenTag, TReadonly, TLParen, TLNumber, TRParen}},
		{"<?php PRIVATE(SET) private( set )", phpver.PHP84, []TokenKind{TOpenTag, TPrivateSet, TPrivate, TLParen, TString, TRParen}},
		{"<?php (void)", phpver.PHP84, []TokenKind{TOpenTag, TLParen, TString, TRParen}},
		{"<?php <<<A\n  A;\n", phpver.PHP72, []TokenKind{TOpenTag, TStartHeredoc, TEncapsedAndWhitespace}},
		{"<?php <<<A\nA;\n", phpver.PHP72, []TokenKind{TOpenTag, TStartHeredoc, TEndHeredoc, TSemicolon}},
		{"<?php <<<A\nA x\nA\n", phpver.PHP72, []TokenKind{TOpenTag, TStartHeredoc, TEncapsedAndWhitespace, TEndHeredoc}},
		{"<?php \"$a?->b\"", phpver.PHP74, []TokenKind{TOpenTag, TDoubleQuote, TVariable, TEncapsedAndWhitespace, TDoubleQuote}},
	}
	for _, c := range cases {
		if got := kinds(c.src, c.ver); !eqKinds(got, c.want) {
			t.Errorf("%q (%s): got %v want %v", c.src, c.ver, got, c.want)
		}
	}
}

func TestLexDefaultVersionAndShortTag(t *testing.T) {
	toks, _ := Lex([]byte("<? $a ?>x<?= 1"), LexOptions{ShortOpenTag: true})
	var ks []TokenKind
	for _, tk := range toks {
		ks = append(ks, tk.Kind)
	}
	want := []TokenKind{TOpenTag, TWhitespace, TVariable, TWhitespace, TCloseTag, TInlineHTML, TOpenTagWithEcho, TWhitespace, TLNumber}
	if !eqKinds(ks, want) {
		t.Fatalf("got %v want %v", ks, want)
	}
}

// TestLexNeverStalls drives the scanner from a state no input reaches: the
// loop guard must still consume one byte per token.
func TestLexNeverStalls(t *testing.T) {
	l := &lexer{src: []byte("ab"), ver: phpver.PHP84, states: []lexState{lexState(255)}, halt: -1}
	toks, _ := l.run()
	if len(toks) != 2 || toks[0].Kind != TBadCharacter || toks[1].End != 2 {
		t.Fatalf("got %v", toks)
	}
}
