package phpunit

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestMetadataArgumentRecovery(t *testing.T) {
	if metadataArgument(nil, "methodName") != nil {
		t.Fatal("nil argument list")
	}
	value := &syntax.Literal{LitKind: syntax.LitString, Raw: `'rows'`}
	for _, args := range [][]syntax.Expr{
		{&syntax.Identifier{Value: "recovery"}},
		{&syntax.Arg{Name: &syntax.Identifier{Value: "methodName"}, Value: value}, &syntax.Arg{Value: value}},
	} {
		if metadataArgument(&syntax.ArgList{Args: args}, "methodName") != nil {
			t.Fatal("invalid bindings")
		}
	}
}

func TestAttributesBeforePHP8(t *testing.T) {
	if len(AttributeMetadata(&analysis.Context{PHP: phpversion.PHP74}, nil)) != 0 {
		t.Fatal("old PHP")
	}
}

func BenchmarkAttributeMetadata(b *testing.B) {
	f := syntax.Parse("test.php", []byte(`<?php use PHPUnit\Framework\Attributes\DataProvider as Rows; class Tests { #[Rows(methodName: 'rows')] public function testRows() {} }`), syntax.Options{Version: phpversion.Max})
	method := f.Stmts[1].(*syntax.ClassLike).Members[0].(*syntax.Method)
	ctx := &analysis.Context{File: f, Src: f.Src, PHP: phpversion.Max}
	ctx.Names()
	b.ReportAllocs()
	for b.Loop() {
		AttributeMetadata(ctx, method.Attrs)
	}
}
