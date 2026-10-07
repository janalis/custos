package util

import (
	"sort"
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/syntax"
)

// probeArg returns the first argument of the last call to probe() in f.
func probeArg(f *syntax.File) syntax.Expr {
	var arg syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok && CallLastName(c) == "probe" {
			args, _ := CallArgValues(c)
			arg = args[0]
		}
		return true
	})
	return arg
}

func TestPossibleValuesReaching(t *testing.T) {
	for src, want := range map[string]string{
		`<?php function f($p = 'a') { probe($p); }`:                         "'a'",
		`<?php function f($p = 'a') { $p = 'b'; probe($p); }`:               "'b'",
		`<?php function f() { $x = 'a'; $x = 'b'; probe($x); }`:             "'b'",
		`<?php function f($c) { $x = $c ? 'a' : 'b'; probe($x ?? 'z'); }`:   "'a' 'b' 'z'",
		`<?php function f() { $x = 'a'; $x .= 'b'; probe($x); }`:            "?",
		`<?php class C { const K = 'k'; function f() { probe(self::K); } }`: "'k'",
	} {
		f := parse(t, src)
		vals, known := PossibleValuesReaching(f, probeArg(f))
		got := "?"
		if known {
			var out []string
			for _, v := range vals {
				out = append(out, text(f, v))
			}
			sort.Strings(out)
			got = strings.Join(out, " ")
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", src, got, want)
		}
	}
}

func TestIsLogicalOperand(t *testing.T) {
	for src, want := range map[string]bool{
		`<?php if (probe(1)) {}`:                true,
		`<?php while ((probe(1))) {}`:           true,
		`<?php $a = !probe(1);`:                 true,
		`<?php $a = $b && probe(1);`:            true,
		`<?php $a = probe(1) ? 1 : 2;`:          true,
		`<?php $a = probe(1) ?: 2;`:             false,
		`<?php $a = probe(1);`:                  false,
		`<?php if ($x == probe(1)) {}`:          false,
		`<?php do {} while (probe(1));`:         true,
		`<?php $a = $b and probe(1);`:           true,
		`<?php $a = probe(1) . 'x';`:            false,
		`<?php if ($a) {} elseif (probe(1)) {}`: true,
	} {
		f := parse(t, src)
		call := probeArg(f).Parent().Parent().Parent()
		if got := IsLogicalOperand(call); got != want {
			t.Errorf("%s: got %v", src, got)
		}
	}
}

func TestArgBindsByRef(t *testing.T) {
	params := []index.Param{{Name: "$a"}, {Name: "$b", ByRef: true}}
	variadic := []index.Param{{Name: "$a"}, {Name: "$rest", ByRef: true, Variadic: true}}
	for _, c := range []struct {
		src     string
		params  []index.Param
		callRef bool
		want    bool
	}{
		{`<?php f(1);`, params, false, false},
		{`<?php f(1, $x);`, params, false, true},
		{`<?php f(b: $x);`, params, false, true},
		{`<?php f(a: $x);`, params, false, false},
		{`<?php f(1, 2, 3);`, variadic, false, true},
		{`<?php f(&$x);`, params, false, false},
		{`<?php f(&$x);`, params, true, true},
	} {
		f := parse(t, c.src)
		var list *syntax.ArgList
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok {
				list = call.Args
			}
			return true
		})
		if got := ArgBindsByRef(list, c.params, c.callRef); got != c.want {
			t.Errorf("%s (callTimeRef=%v): got %v", c.src, c.callRef, got)
		}
	}
	if ArgBindsByRef(nil, params, true) {
		t.Error("nil list")
	}
}

func TestSmallExprHelpers(t *testing.T) {
	f := parse(t, `<?php probe([-1, 2.5, -$x, 'a\'b', "c\td", <<<EOT
h
EOT, `+"`ls`"+`, "x$y"]);`)
	arr := probeArg(f).(*syntax.Array)
	el := func(i int) syntax.Expr { return arr.Items[i].Value }
	for i, want := range []bool{true, true, false, false, false, false, false, false} {
		if got := IsNumberLiteral(el(i)); got != want {
			t.Errorf("IsNumberLiteral(%s) = %v", text(f, el(i)), got)
		}
	}
	for i, want := range []bool{false, false, false, true, true, true, false, true} {
		if got := IsStringLiteral(el(i)); got != want {
			t.Errorf("IsStringLiteral(%s) = %v", text(f, el(i)), got)
		}
	}
	if v, ok := QuotedStringValue(el(3)); !ok || v != "a'b" {
		t.Errorf("QuotedStringValue single = %q %v", v, ok)
	}
	if v, ok := QuotedStringValue(el(4)); !ok || v != "c\td" {
		t.Errorf("QuotedStringValue double = %q %v", v, ok)
	}
	if _, ok := QuotedStringValue(el(0)); ok {
		t.Error("QuotedStringValue on a number")
	}
	if !MentionsVariable(arr, "x") || MentionsVariable(arr, "z") || MentionsVariable(nil, "x") {
		t.Error("MentionsVariable")
	}
	if AsExpr(arr) == nil || AsExpr(f.Stmts[0]) != nil {
		t.Error("AsExpr")
	}
}

func TestIndentBefore(t *testing.T) {
	src := []byte("a\n\t  b = c;\n")
	if got := IndentBefore(src, 5); got != "\t  " {
		t.Errorf("IndentBefore at b = %q", got)
	}
	if got := IndentBefore(src, 9); got != "" {
		t.Errorf("IndentBefore at c = %q", got)
	}
	if got := IndentBefore(src, 0); got != "" {
		t.Errorf("IndentBefore at 0 = %q", got)
	}
}

func TestIsStaticPropName(t *testing.T) {
	f := parse(t, `<?php X::$p = $q;`)
	var names []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && IsStaticPropName(v) {
			names = append(names, v.Name)
		}
		return true
	})
	if strings.Join(names, ",") != "p" {
		t.Errorf("got %v", names)
	}
}

func TestDocHasAnnotation(t *testing.T) {
	f := parse(t, `<?php class C {
	/** @var int */
	private $a;
	/** @ORM\Column */
	private $b;
	private $c;
}`)
	var got []bool
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if p, ok := n.(*syntax.Property); ok {
			got = append(got, DocHasAnnotation(f, p))
		}
		return true
	})
	if len(got) != 3 || got[0] || !got[1] || got[2] {
		t.Errorf("got %v", got)
	}
}
