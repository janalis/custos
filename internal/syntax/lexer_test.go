package syntax

import (
	"strings"
	"testing"

	"custos/internal/phpver"
)

func kinds(src string, v phpver.Version) []TokenKind {
	toks, _ := Lex([]byte(src), LexOptions{Version: v})
	var out []TokenKind
	for _, t := range toks {
		if !t.Kind.IsTrivia() {
			out = append(out, t.Kind)
		}
	}
	return out
}

func eqKinds(a, b []TokenKind) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLexVersionGating(t *testing.T) {
	cases := []struct {
		name string
		src  string
		ver  phpver.Version
		want []TokenKind
	}{
		{"fn keyword 7.4", "<?php fn", phpver.PHP74, []TokenKind{TOpenTag, TFn}},
		{"fn identifier 7.3", "<?php fn", phpver.PHP73, []TokenKind{TOpenTag, TString}},
		{"match identifier 7.4", "<?php match", phpver.PHP74, []TokenKind{TOpenTag, TString}},
		{"attribute 8.0", "<?php #[A]", phpver.PHP80, []TokenKind{TOpenTag, TAttribute, TString, TRBracket}},
		{"hash comment 7.4", "<?php #[A]", phpver.PHP74, []TokenKind{TOpenTag}},
		{"enum class", "<?php enum Foo", phpver.PHP81, []TokenKind{TOpenTag, TEnum, TString}},
		{"enum as name", "<?php enum(1)", phpver.PHP81, []TokenKind{TOpenTag, TString, TLParen, TLNumber, TRParen}},
		{"enum extends", "<?php class enum extends", phpver.PHP81, []TokenKind{TOpenTag, TClass, TString, TExtends}},
		{"yield from", "<?php yield  from $a", phpver.PHP70, []TokenKind{TOpenTag, TYieldFrom, TVariable}},
		{"yield from 5.6", "<?php yield from", phpver.PHP56, []TokenKind{TOpenTag, TYield, TString}},
		{"private(set)", "<?php private(set) int", phpver.PHP84, []TokenKind{TOpenTag, TPrivateSet, TString}},
		{"private(set) 8.3", "<?php private(set)", phpver.PHP83, []TokenKind{TOpenTag, TPrivate, TLParen, TString, TRParen}},
		{"pipe 8.5", "<?php $a |> f(...)", phpver.PHP85, []TokenKind{TOpenTag, TVariable, TPipe, TString, TLParen, TEllipsis, TRParen}},
		{"void cast 8.5", "<?php (void) f()", phpver.PHP85, []TokenKind{TOpenTag, TVoidCast, TString, TLParen, TRParen}},
		{"property after arrow", "<?php $a->class", phpver.PHP84, []TokenKind{TOpenTag, TVariable, TObjectOperator, TString}},
		{"keyword after ::", "<?php A::class", phpver.PHP84, []TokenKind{TOpenTag, TString, TPaamayimNekudotayim, TClass}},
		{"names", `<?php \A\B C\D namespace\E`, phpver.PHP84, []TokenKind{TOpenTag, TNameFullyQualified, TNameQualified, TNameRelative}},
		{"cast spaces", "<?php ( int )$a", phpver.PHP84, []TokenKind{TOpenTag, TIntCast, TVariable}},
		{"not a cast", "<?php (int $a)", phpver.PHP84, []TokenKind{TOpenTag, TLParen, TString, TVariable, TRParen}},
		{"interpolation", `<?php "a $b[0] {$c->d} ${e}"`, phpver.PHP84, []TokenKind{
			TOpenTag, TDoubleQuote, TEncapsedAndWhitespace, TVariable, TLBracket, TNumString, TRBracket,
			TEncapsedAndWhitespace, TCurlyOpen, TVariable, TObjectOperator, TString, TRBrace,
			TEncapsedAndWhitespace, TDollarOpenCurlyBraces, TStringVarname, TRBrace, TDoubleQuote,
		}},
		{"plain double quoted", `<?php "a {b} \$c"`, phpver.PHP84, []TokenKind{TOpenTag, TConstantEncapsedString}},
		{"heredoc flexible", "<?php <<<EOT\n  a $b\n  EOT;\n", phpver.PHP73, []TokenKind{
			TOpenTag, TStartHeredoc, TEncapsedAndWhitespace, TVariable, TEncapsedAndWhitespace, TEndHeredoc, TSemicolon,
		}},
		{"heredoc legacy", "<?php <<<EOT\n  EOT\nEOT;\n", phpver.PHP72, []TokenKind{
			TOpenTag, TStartHeredoc, TEncapsedAndWhitespace, TEndHeredoc, TSemicolon,
		}},
		{"nowdoc", "<?php <<<'X'\n$a\nX;\n", phpver.PHP84, []TokenKind{TOpenTag, TStartHeredoc, TEncapsedAndWhitespace, TEndHeredoc, TSemicolon}},
		{"halt compiler", "<?php __halt_compiler(); <?php junk", phpver.PHP84, []TokenKind{
			TOpenTag, THaltCompiler, TLParen, TRParen, TSemicolon, THaltCompilerData,
		}},
		{"close tag", "<?php echo 1 ?>\nhtml", phpver.PHP84, []TokenKind{TOpenTag, TEcho, TLNumber, TCloseTag, TInlineHTML}},
		{"numbers", "<?php 0x1F 0b11 0o17 1_000 1.5e3 .5 9223372036854775808", phpver.PHP84, []TokenKind{
			TOpenTag, TLNumber, TLNumber, TLNumber, TLNumber, TDNumber, TDNumber, TDNumber,
		}},
	}
	for _, c := range cases {
		if got := kinds(c.src, c.ver); !eqKinds(got, c.want) {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestLexShortOpenTag(t *testing.T) {
	toks, _ := Lex([]byte("a<? echo 1;"), LexOptions{Version: phpver.PHP84, ShortOpenTag: true})
	if toks[0].Kind != TInlineHTML || toks[1].Kind != TOpenTag || toks[1].End-toks[1].Start != 2 {
		t.Fatalf("got %v", toks)
	}
	toks, _ = Lex([]byte("a<? echo 1;"), LexOptions{Version: phpver.PHP84})
	if len(toks) != 1 || toks[0].Kind != TInlineHTML {
		t.Fatalf("short tags off: got %v", toks)
	}
}

// checkContiguous verifies the token stream covers src exactly.
func checkContiguous(t *testing.T, src []byte, toks []Token) {
	t.Helper()
	pos := uint32(0)
	for i, tk := range toks {
		if tk.Start != pos || tk.End < tk.Start {
			t.Fatalf("token %d (%s) [%d:%d] not contiguous at %d", i, tk.Kind, tk.Start, tk.End, pos)
		}
		pos = tk.End
	}
	if int(pos) != len(src) {
		t.Fatalf("tokens end at %d, src len %d", pos, len(src))
	}
}

func FuzzLex(f *testing.F) {
	for _, s := range []string{
		"<?php echo 1;", `<?php "a $b[1] {$c} ${d}"`, "<?php <<<A\nx $y\nA;\n", "x<?= $a ?>y",
		"<?php /** doc */ #[A] fn($x) => $x?->y", "<?php __halt_compiler();data",
		"<?php `ls $dir` ?>", "<?php (int)(float)$x <=> $y ??= 1",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		for _, v := range []phpver.Version{phpver.PHP56, phpver.PHP74, phpver.PHP85} {
			src := []byte(s)
			toks, _ := Lex(src, LexOptions{Version: v, ShortOpenTag: strings.Contains(s, "short")})
			checkContiguous(t, src, toks)
		}
	})
}
