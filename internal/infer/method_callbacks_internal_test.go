package infer

import (
	"testing"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/types"
)

func TestMethodCallbackRecoveryAndNative(t *testing.T) {
	f := syntax.Parse("t.php", []byte(`<?php final class C { /** @return int */ public static function text(): string { return ''; } } call_user_func([C::class, 'text']);`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := NewEnv(f, names.New(f), ix, phpver.PHP85)
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok {
			if got := env.Native().TypeOf(call); !got.Equal(types.String) {
				t.Errorf("native callback: %s", got.String())
			}
		}
		return true
	})
	for _, expr := range []syntax.Expr{
		&syntax.Array{Items: []*syntax.ArrayItem{nil, {Value: &syntax.Literal{LitKind: syntax.LitString, Raw: "'x'"}}}},
		&syntax.Array{Items: []*syntax.ArrayItem{{}, {Value: &syntax.Literal{LitKind: syntax.LitString, Raw: "'x'"}}}},
		&syntax.Array{Items: []*syntax.ArrayItem{{Value: &syntax.Literal{LitKind: syntax.LitString, Raw: "\"$x\""}}, {Value: &syntax.Literal{LitKind: syntax.LitString, Raw: "'x'"}}}},
		&syntax.Array{Items: []*syntax.ArrayItem{{Value: &syntax.ClassConstFetch{Class: &syntax.Name{Value: "parent"}, Name: &syntax.Identifier{Value: "class"}}}, {Value: &syntax.Literal{LitKind: syntax.LitString, Raw: "'x'"}}}},
	} {
		if got, ok := env.literalMethodCallback(expr); !ok || !got.IsUnknown() {
			t.Errorf("recovery callback: %s, %v", got.String(), ok)
		}
	}
}
