package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/types"
)

func typeArgsIndex(t *testing.T, src string) *indexRelations {
	t.Helper()
	ix := index.New(nil)
	ix.Add(index.Extract(syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpversion.PHP84})))
	return newIndexRelations(ix, phpversion.PHP84)
}

func TestTypeArgsAs(t *testing.T) {
	r := typeArgsIndex(t, `<?php
/** @template T */
class Box {}
/** @extends Box<string> */
class Text extends Box {}
class Deep extends Text {}
/**
 * @template U
 * @extends Box<array<int, U>>
 */
class Lists extends Box {}
/** @template V */
interface Seen {}
/** @implements Seen<int> */
class Both extends Text implements Seen {}
class Plain {}
`)
	for _, c := range []struct {
		class    string
		args     []types.Type
		ancestor string
		want     string
	}{
		{`\Text`, nil, `\Box`, "string"},
		{`\Deep`, nil, `\Box`, "string"},
		{`\Lists`, []types.Type{types.Int}, `\Box`, "array<int,int>"},
		{`\Lists`, nil, `\Box`, "array<int,mixed>"},
		{`\Both`, nil, `\Seen`, "int"},
		{`\Both`, nil, `\Box`, "string"},
		{`\Plain`, nil, `\Box`, ""},
		{`\Missing`, nil, `\Box`, ""},
	} {
		got := r.TypeArgsAs(c.class, c.args, c.ancestor)
		text := ""
		if len(got) > 0 {
			text = got[0].DocString()
		}
		if text != c.want {
			t.Errorf("%s as %s = %q, want %q", c.class, c.ancestor, text, c.want)
		}
	}
}

// Each level repeats its template: the arguments would double at every step.
func TestTypeArgsAsIsBounded(t *testing.T) {
	var src strings.Builder
	src.WriteString("<?php\n/** @template T */\nclass C0 {}\n")
	const levels = maxTypeArgClasses - 2
	for i := 1; i <= levels; i++ {
		fmt.Fprintf(&src, "/**\n * @template T\n * @extends C%d<array{a: T, b: T}>\n */\nclass C%d extends C%d {}\n", i-1, i, i-1)
	}
	r := typeArgsIndex(t, src.String())
	for _, arg := range r.TypeArgsAs(fmt.Sprintf(`\C%d`, levels), []types.Type{types.Int}, `\C0`) {
		if n := len(arg.DocString()); n > 4*types.MaxDocTypeLen {
			t.Errorf("argument of %d bytes", n)
		}
	}
}

func TestTypeArgsAsStopsWhenTheBudgetRunsOut(t *testing.T) {
	r := typeArgsIndex(t, `<?php
/** @template T */
class Far {}
/**
 * @template T
 * @extends Far<T>
 */
class Start extends Far {}
`)
	classes := make([]types.Type, 0, 3000)
	for i := range cap(classes) {
		classes = append(classes, types.Of(fmt.Sprintf(`\\Some\\Rather\\Long\\Namespace\\Class%d`, i)))
	}
	huge := types.Union(classes...)
	if n := len(huge.DocString()); n <= maxTypeArgWork {
		t.Fatalf("argument of only %d bytes", n)
	}
	if got := r.TypeArgsAs(`\Start`, []types.Type{huge}, `\Far`); got != nil {
		t.Errorf("arguments past the budget: %d", len(got))
	}
	if got := r.TypeArgsAs(`\Start`, []types.Type{types.Int}, `\Far`); len(got) != 1 {
		t.Errorf("within the budget: %d arguments", len(got))
	}
}

func BenchmarkTypeArgsAs(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php\n/** @template T */\nclass C0 {}\n")
	for i := 1; i <= 8; i++ {
		fmt.Fprintf(&src, "/**\n * @template T\n * @extends C%d<T>\n */\nclass C%d extends C%d {}\n", i-1, i, i-1)
	}
	ix := index.New(nil)
	ix.Add(index.Extract(syntax.Parse("t.php", []byte(src.String()), syntax.Options{Version: phpversion.PHP84})))
	args := []types.Type{types.Int}
	b.Run("cached", func(b *testing.B) {
		r := newIndexRelations(ix, phpversion.PHP84)
		for b.Loop() {
			r.TypeArgsAs(`\C8`, args, `\C0`)
		}
	})
	b.Run("walk", func(b *testing.B) {
		for b.Loop() {
			newIndexRelations(ix, phpversion.PHP84).TypeArgsAs(`\C8`, args, `\C0`)
		}
	})
}

// An argument of unknown type binds nothing: its template reads as mixed.
func TestTypeArgsAsUnknownArgument(t *testing.T) {
	r := typeArgsIndex(t, `<?php
/** @template T */
class Box {}
/**
 * @template U
 * @extends Box<U>
 */
class Sub extends Box {}
`)
	if got := r.TypeArgsAs(`\Sub`, []types.Type{types.Unknown}, `\Box`); len(got) != 1 || !got[0].Has("mixed") {
		t.Errorf("TypeArgsAs = %v", got)
	}
}

func TestTypeArgsDeclaredStaticCalls(t *testing.T) {
	probeNative(t, `class F { /** @return array<int, int> */ public static function make() {} }
$c = 'F';
probe($c::make(), F::{'make'}(), F::make());`, func(ctx *analysis.Context, call *syntax.FuncCall) {
		for i, want := range []bool{false, false, true} {
			if got := typeArgsDeclared(ctx, CallArgument(call.Args, i, ""), 0); got != want {
				t.Errorf("argument %d: %v, want %v", i, got, want)
			}
		}
	})
}
