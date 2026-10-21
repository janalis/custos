package streamopenfailureunchecked

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// warningProbe records joinedWarningUnknown for the call passed to probe().
type warningProbe struct{ got *[]bool }

func (warningProbe) ID() string               { return "StreamOpenFailureUnchecked" }
func (warningProbe) Semantic()                {}
func (warningProbe) Flow()                    {}
func (warningProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (p warningProbe) Check(ctx *analysis.Context, n syntax.Node) {
	if name, ok := n.(*syntax.FuncCall).Name.(*syntax.Name); ok && name.Value == "probe" {
		*p.got = append(*p.got, joinedWarningUnknown(ctx, n.(*syntax.FuncCall).Args.Args[0].(*syntax.Arg).Value.(*syntax.FuncCall)))
	}
}

func warningUnknown(t *testing.T, src string) bool {
	t.Helper()
	var got []bool
	e, err := analysis.NewEngine([]analysis.Rule{warningProbe{&got}}, analysis.Config{Only: []string{"StreamOpenFailureUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("warning.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %s: %v", src, f.Errors)
	}
	e.Analyze(f)
	if len(got) != 1 {
		t.Fatalf("probed %d calls", len(got))
	}
	return got[0]
}

func TestJoinedWarningUnknown(t *testing.T) {
	for _, tc := range []struct {
		name, before string
		want         bool
	}{
		{"none", "", false},
		{"throwing arrow", "set_error_handler(fn()=>throw new E());", true},
		{"throwing closure", "set_error_handler(function(){throw new E();});", true},
		{"return false", "set_error_handler(function(){return false;});", false},
		{"return true", "set_error_handler(fn()=>true);", true},
		{"typed callback", "set_error_handler(fn(array $a)=>false);", true},
		{"extra default", "set_error_handler(fn($a,$b,$c,$d,$e=new E())=>false);", true},
		{"unpacked callback", "set_error_handler(...$o);", true},
		{"named handler", "set_error_handler('handler');", true},
		{"larger handler", "set_error_handler(function(){work();return false;});", true},
		{"restored", "set_error_handler(fn()=>throw new E());restore_error_handler();", false},
		{"conditional restore", "set_error_handler(fn()=>throw new E());if($c){restore_error_handler();}", true},
		{"conditional registration", "if($c){set_error_handler(fn()=>throw new E());}", true},
		{"project call", "set_error_handler(fn()=>false);work();", true},
		{"method", "set_error_handler(fn()=>false);$o->work();", true},
		{"property", "set_error_handler(fn()=>false);$p=$o->p;", true},
		{"array", "set_error_handler(fn()=>false);$p=$o[0];", true},
		{"unset", "set_error_handler(fn()=>false);unset($o);", true},
		{"destructor assignment", "set_error_handler(fn()=>false);$o=null;", true},
		{"property assignment", "set_error_handler(fn()=>false);$o->p=1;", true},
		{"conversion", "set_error_handler(fn()=>false);$p='x'.$o;", true},
		{"cast", "set_error_handler(fn()=>false);$p=(string)$o;", true},
		{"scalar call", "set_error_handler(fn()=>false);str_replace('x','y','z');", false},
		{"object call", "set_error_handler(fn()=>false);str_replace('x','y',$o);", true},
		{"fresh local", "set_error_handler(fn()=>false);$p='a';$p='b';", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "function f($o,$c){" + tc.before + "probe(fopen('local','r'));}"
			if got := warningUnknown(t, source); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
	for _, source := range []string{
		`set_error_handler(fn()=>throw new E());probe(fopen('x','r'));`,
		`function f($c){while($c){probe(fopen('x','r'));set_error_handler(fn()=>throw new E());}}`,
		`function f(){again: probe(fopen('x','r'));set_error_handler(fn()=>throw new E());goto again;}`,
		"function f(){" + strings.Repeat("$p='x';", 2050) + "probe(fopen('x','r'));}",
	} {
		if got := warningUnknown(t, source); !got {
			t.Fatalf("unexpected boundary %v", got)
		}
	}
}
