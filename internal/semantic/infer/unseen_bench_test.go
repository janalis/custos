package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// BenchmarkLiteralLocalReads types every expression of a long function
// whose locals start from literals and are read again and again, passed to
// calls of unknown and known callees: each read checks whether a write its
// definitions do not show may have replaced the literal's value.
func BenchmarkLiteralLocalReads(b *testing.B) {
	const locals, stmts = 40, 600
	var sb strings.Builder
	sb.WriteString("<?php\nfunction helper(string $s): string { return $s; }\n")
	sb.WriteString("function run($obj, array $rows) {\n")
	for i := range locals {
		fmt.Fprintf(&sb, "    $v%d = 'value%d'; $n%d = %d;\n", i, i, i, i)
	}
	sb.WriteString("    $ref = &$v0;\n")
	for i := range stmts {
		a, c := i%locals, (i*7+3)%locals
		switch i % 4 {
		case 0:
			fmt.Fprintf(&sb, "    $obj->m%d($v%d, $n%d);\n", i%9, a, c)
		case 1:
			fmt.Fprintf(&sb, "    helper($v%d . $v%d);\n", a, c)
		case 2:
			fmt.Fprintf(&sb, "    if ($rows) { $obj->call(strlen($v%d) + $n%d); }\n", a, c)
		default:
			fmt.Fprintf(&sb, "    $out[] = [$v%d, $n%d, $ref];\n", a, c)
		}
	}
	sb.WriteString("    return $out;\n}\n")
	src := []byte(sb.String())
	opt := syntax.Options{Version: phpversion.PHP84}
	f := syntax.Parse("t.php", src, opt)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	nm := names.New(f)
	b.ReportAllocs()
	for b.Loop() {
		env := infer.NewEnv(f, nm, ix, phpversion.PHP84, infer.Options{})
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if x, ok := n.(syntax.Expr); ok {
				_ = env.TypeOf(x)
			}
			return true
		})
	}
}
