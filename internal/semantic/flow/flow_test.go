package flow

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

func environment(t testing.TB, src string, snapshot *Snapshot) (*Env, *syntax.File) {
	t.Helper()
	f := syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse: %v", f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	return New(f, infer.NewEnv(f, names.New(f), ix, phpversion.Max), snapshot), f
}

func callsNamed(f *syntax.File, name string) []*syntax.FuncCall {
	var calls []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if nm, ok := c.Name.(*syntax.Name); ok && nm.Value == name {
				calls = append(calls, c)
			}
		}
		return true
	})
	return calls
}

func argument(c *syntax.FuncCall) syntax.Expr { return c.Args.Args[0].(*syntax.Arg).Value }

func TestResolveAndGuards(t *testing.T) {
	for _, tc := range []struct {
		src            string
		resolve, taint bool
		context        Context
	}{
		{`$x='literal'; probe($x);`, true, false, HTML},
		{`$x=$_GET['x']; probe($x);`, true, true, HTML},
		{`$x=$_GET['x']; if($flag){$x='safe';} probe($x);`, false, true, HTML},
		{`$x=$_GET['x']; $x=htmlspecialchars($x); probe($x);`, true, false, HTML},
		{`$x=$_GET['x']; $x=unknown($x); probe($x);`, true, false, HTML},
		{`$x=$_GET['x']; if (!in_array($x,['/home','/help'],true)){exit;} probe($x);`, true, false, URL},
		{`$x=$_GET['x']; if (!ctype_digit($x)){return;} probe($x);`, true, false, SQL},
		{`$x=$_GET['x']; if(str_contains($x,"\r") || str_contains($x,"\n")){exit;} probe($x);`, true, false, Header},
		{`$x=$_GET['x']; if($x !== 'safe'){return;} probe($x);`, true, false, HTML},
		{`$x=$_GET['x']; while($flag){$x='safe';} probe($x);`, false, true, HTML},
		{`$x=$_GET['x']; while($flag){break;} probe($x);`, true, true, HTML},
		{`$x=$_GET['x']; do{continue;}while($flag); probe($x);`, true, true, HTML},
		{`$x=$_GET['x']; for($i=0;$i<2;$i++){ $x='safe'; } probe($x);`, false, true, HTML},
		{`$x=$_GET['x']; foreach($items as $x){} probe($x);`, false, false, HTML},
		{`$x=$_GET['x']; unset($x);probe($x);`, false, false, HTML},
		{`$x=$_GET['x']; eval($code);probe($x);`, false, false, HTML},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			x := argument(callsNamed(f, "probe")[0])
			_, known := e.Resolve(x)
			if known != tc.resolve {
				t.Errorf("resolve=%v", known)
			}
			if actual := e.Tainted(x, tc.context); actual != tc.taint {
				t.Errorf("taint=%v value=%+v", actual, e.Value(x))
			}
		})
	}
}

func TestStateAndFinally(t *testing.T) {
	for _, tc := range []struct {
		src    string
		closed bool
	}{
		{`$h=fopen('a','r'); fclose($h); fread($h,1);`, true},
		{`$h=fopen('a','r');$alias=$h;fclose($alias);fread($h,1);`, true},
		{`$h=fopen('a','r');if($flag){fclose($h);}fread($h,1);`, false},
		{`$h=fopen('a','r');fclose($h);$h=fopen('b','r');fread($h,1);`, false},
		{`$h=fopen('a','r');fclose($h);escape($h);fread($h,1);`, false},
		{`$h=fopen('a','r');try{work();}finally{fclose($h);}fread($h,1);`, true},
		{`function f($h){fclose($h);fread($h,1);}`, true},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			c := callsNamed(f, "fread")[0]
			_, ok := e.StateBefore(c, argument(c), "fclose")
			if ok != tc.closed {
				t.Errorf("closed=%v", ok)
			}
		})
	}
}

func TestScopesAndWrappers(t *testing.T) {
	e, f := environment(t, `function passthrough($x){return $x;} function encode($x){return htmlspecialchars($x);} function sink($x){system($x);} function generator(){yield system($_GET['x']);}`, nil)
	file := Extract(f, e.types.Index, phpversion.Max, nil)
	snapshot := NewSnapshot(file)
	for _, tc := range []struct {
		src   string
		taint bool
	}{
		{`$x=passthrough($_GET['x']);probe($x);`, true},
		{`$x=encode($_GET['x']);probe($x);`, false},
		{`$x=passthrough(x: $_GET['x']);probe($x);`, true},
	} {
		local, g := environment(t, tc.src, snapshot)
		local.types.Index = index.New(e.types.Index)
		local.types.Index.Add(index.Extract(g))
		if actual := local.Tainted(argument(callsNamed(g, "probe")[0]), HTML); actual != tc.taint {
			t.Fatalf("%s taint=%v", tc.src, actual)
		}
	}
	local, g := environment(t, `sink($_GET['x']);$g=generator();probe($g);`, snapshot)
	local.types.Index = index.New(e.types.Index)
	sinks := local.Sinks(callsNamed(g, "sink")[0])
	if len(sinks) != 1 || len(sinks[0].Argument.Sources) != 1 {
		t.Fatalf("sinks=%+v", sinks)
	}
	if v := local.Value(argument(callsNamed(g, "probe")[0])); v.Identity == 0 || !v.Complete {
		t.Fatalf("generator=%+v", v)
	}
	if len(file.Summaries["generator"].Sinks) != 0 {
		t.Fatal("deferred generator sinks escaped")
	}
	if _, ok := snapshot.WithoutFile(f.Path).lookup("encode"); ok {
		t.Fatal("stale summary")
	}
	if _, ok := NewSnapshot(file, &File{Path: "other.php", Summaries: file.Summaries}).lookup("encode"); ok {
		t.Fatal("ambiguous summary")
	}
	if len(e.Calls(nil)) != 0 || len(e.Calls(f.Stmts[0])) != 0 {
		t.Fatal("nested scope calls escaped")
	}
}

func TestValueBounds(t *testing.T) {
	a := Value{Complete: true}
	for i := 0; i < MaxSources+1; i++ {
		a = merge(a, Value{Complete: true, Sources: []Source{{Kind: fmt.Sprint(i), Parameter: -1}}})
	}
	if a.Complete || len(a.Sources) != MaxSources {
		t.Fatal("source budget")
	}
	e, f := environment(t, `$x=1;switch($x){case 1:$x=2;}probe($x);`, nil)
	if _, ok := e.Resolve(argument(callsNamed(f, "probe")[0])); ok {
		t.Fatal("unsupported scope established proof")
	}
}

func BenchmarkFlow(b *testing.B) {
	var src strings.Builder
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&src, "$v%d=$_GET['x'];", i)
	}
	src.WriteString("probe($v199);")
	e, f := environment(b, src.String(), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := New(f, e.types, nil)
		env.Value(x)
	}
}
