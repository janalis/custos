package astquery

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

// ProbeArg returns the first argument of the last call to probe() in f.
func ProbeArg(f *syntax.File) syntax.Expr {
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
		f := Parse(t, src)
		call := ProbeArg(f).Parent().Parent().Parent()
		if got := IsLogicalOperand(call); got != want {
			t.Errorf("%s: got %v", src, got)
		}
	}
}

func TestSmallExprHelpers(t *testing.T) {
	f := Parse(t, `<?php probe([-1, 2.5, -$x, 'a\'b', "c\td", <<<EOT
h
EOT, `+"`ls`"+`, "x$y"]);`)
	arr := ProbeArg(f).(*syntax.Array)
	el := func(i int) syntax.Expr { return arr.Items[i].Value }
	for i, want := range []bool{true, true, false, false, false, false, false, false} {
		if got := IsNumberLiteral(el(i)); got != want {
			t.Errorf("IsNumberLiteral(%s) = %v", Text(f, el(i)), got)
		}
	}
	for i, want := range []bool{false, false, false, true, true, true, false, true} {
		if got := IsStringLiteral(el(i)); got != want {
			t.Errorf("IsStringLiteral(%s) = %v", Text(f, el(i)), got)
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
	f := Parse(t, `<?php X::$p = $q;`)
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
