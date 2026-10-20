package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// sharedDefectProbe evaluates one shared query at every `probe(...)` call.
type sharedDefectProbe struct {
	query   string
	results *[]string
}

func (sharedDefectProbe) ID() string               { return "AbstractClassInstantiation" }
func (sharedDefectProbe) Semantic()                {}
func (sharedDefectProbe) Flow()                    {}
func (sharedDefectProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (p sharedDefectProbe) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if name, ok := c.Name.(*syntax.Name); !ok || name.Value != "probe" {
		return
	}
	a, b := CallArgument(c.Args, 0, ""), CallArgument(c.Args, 1, "")
	result := ""
	switch p.query {
	case "origin":
		if v := NativeLocalValue(ctx, a); v != nil {
			result = ctx.Text(v)
		}
	case "origin-of-scandir":
		scanned := func(e syntax.Expr) bool {
			call, ok := e.(*syntax.FuncCall)
			return ok && NativeBuiltin(ctx, call, "scandir")
		}
		if v := NativeLocalResultOf(ctx, a, "false", scanned); v != nil {
			result = ctx.Text(v)
		}
	case "only":
		if v := NativeLocalOnly(ctx, a); v != nil {
			result = ctx.Text(v)
		}
	case "origin-false", "origin-null":
		if v := NativeLocalResult(ctx, a, strings.TrimPrefix(p.query, "origin-")); v != nil {
			result = ctx.Text(v)
		}
	case "same":
		result = fmt.Sprint(NativeSameValue(ctx, a, b))
	case "stable":
		result = fmt.Sprint(expansionDSame(ctx, a, b))
	case "false":
		result = fmt.Sprint(NativeFailureExcluded(ctx, a, types.Of("resource", "false"), "false"))
	case "local-false":
		result = fmt.Sprint(NativeFailureExcluded(ctx, a, types.Of("resource", "false"), "false"))
	case "static":
		result = fmt.Sprint(NativeStaticLocal(ctx, a))
	case "writes-known":
		scope := syntax.EnclosingVariableScope(a)
		result = fmt.Sprint(nativeContractScope(ctx, scope).localWrites(ctx, scope).known)
	case "warning-unknown":
		result = fmt.Sprint(nativeJoinedWarningUnknown(ctx, a.(*syntax.FuncCall)))
	case "string-nonnull":
		result = fmt.Sprint(NativeNarrowedExcludes(ctx, a, types.Of("string", "null"), "null"))
	case "length":
		result = fmt.Sprint(NativeLengthProven(ctx, a, 4))
	case "class":
		if created, ok := a.(*syntax.New); ok {
			if class := NativeNewClass(ctx, created); class != nil {
				result = class.FQN
			}
		}
	case "int":
		v, known := NativeInt(ctx, a)
		result = fmt.Sprintf("%d/%t", v, known)
	case "truthy":
		result = fmt.Sprint(ExpansionCTruthy(c))
	case "dominated":
		// Every earlier call named mark() that dominates the probe.
		count := 0
		for _, event := range nativeContractEvents(ctx, syntax.EnclosingVariableScope(c)) {
			if mark, ok := event.(*syntax.FuncCall); ok && ctx.Text(mark.Name) == "mark" && NativeDominates(mark, c) {
				count++
			}
		}
		result = fmt.Sprint(count)
	}
	*p.results = append(*p.results, result)
}

func sharedDefectResults(t testing.TB, query, src, php string) string {
	t.Helper()
	var got []string
	r := sharedDefectProbe{query, &got}
	cfg := analysis.Config{Only: []string{r.ID()}}
	if php != "" {
		cfg.PHP = phpversion.MustParse(php)
	}
	e, err := analysis.NewEngine([]analysis.Rule{r}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("shared.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %s: %v", src, f.Errors)
	}
	e.Analyze(f)
	return strings.Join(got, ",")
}

// A variable several assignments reach resolves to the first call result no
// guard on its own path covers: the one a failure check has to worry about.
func TestSharedLocalValueAcrossPaths(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"loop, first result tested", `function f($a, $c){$u=parse_url($a); if($u === false){return;} while($c--){probe($u); $u=scandir($a);}}`, "scandir($a)"},
		{"loop, first result untested", `function f($a, $c){$u=parse_url($a); while($c--){probe($u); $u=scandir($a);}}`, "parse_url($a)"},
		{"loop, tested inside", `function f($a, $c){$u=parse_url($a); while($c--){if(!$u){return;} probe($u); $u=scandir($a);}}`, "parse_url($a)"},
		{"loop, literal first", `function f($a, $c){$u=['k'=>1]; while($c--){probe($u); $u=scandir($a);}}`, "scandir($a)"},
		{"loop, literal later", `function f($a, $c){$u=parse_url($a); if($u === false){return;} while($c--){probe($u); $u=false;}}`, "parse_url($a)"},
		{"loop, unassigned first", `function f($a, $c){while($c--){probe($u); $u=scandir($a);}}`, "scandir($a)"},
		{"one path, first result tested", `function f($a, $c){$u=parse_url($a); if($u === false){return;} if($c){$u=scandir($a);} probe($u);}`, "scandir($a)"},
		{"one path, method result", `function f($a, $c){$u=parse_url($a); if($u === false){return;} if($c){$u=$a->load();} probe($u);}`, "$a->load()"},
		{"one path, static result", `function f($a, $c){$u=parse_url($a); if($u === false){return;} if($c){$u=Repo::load();} probe($u);}`, "Repo::load()"},
		{"one path, both tested", `function f($a, $c){$u=parse_url($a); if($u === false){return;} if($c){$u=scandir($a);} if($u === false){return;} probe($u);}`, ""},
		{"null excluded for a result that fails with false", `function f($a, $c){$u=parse_url($a); if($c){$u=scandir($a);} if(!isset($u)){return;} probe($u);}`, "parse_url($a)"},
		{"null excluded for a result that fails with null", `function f($a, $c){$u=preg_replace('/a/', 'b', $a); if($c){$u=preg_replace('/c/', 'd', $a);} if($u === null){return;} probe($u);}`, ""},
		{"false excluded for a result that fails with null", `function f($a, $c){$u=preg_replace('/a/', 'b', $a); if($c){$u=preg_replace('/c/', 'd', $a);} if($u === false){return;} probe($u);}`, "preg_replace('/a/', 'b', $a)"},
		{"false excluded for a result of unknown type", `function f($a, $c){$u=load($a); if($c){$u=other($a);} if($u === false){return;} probe($u);}`, ""},
		{"one path, literals", `function f($a, $c){$u=1; if($c){$u=2;} probe($u);}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, "origin", tc.src, ""); got != tc.want {
				t.Fatalf("origin=%q, want %q", got, tc.want)
			}
		})
	}
}

// An inspection that states the failure value it is about is not answered
// from the declared return type of whatever call produced the value.
func TestSharedLocalResultUsesTheStatedFailure(t *testing.T) {
	const twoResults = `function f($a, $c){$u=parse_url($a); if($u === false){return;} if($c){$u=`
	for _, tc := range []struct{ name, query, src, want string }{
		{"method result, null excluded, declared type", "origin", twoResults + `$a->load();} if($u === null){return;} probe($u);}`, ""},
		{"method result, null excluded, fails with false", "origin-false", twoResults + `$a->load();} if($u === null){return;} probe($u);}`, "$a->load()"},
		{"method result, false excluded, fails with false", "origin-false", twoResults + `$a->load();} if($u === false){return;} probe($u);}`, ""},
		{"method result, false excluded, fails with null", "origin-null", twoResults + `$a->load();} if($u === false){return;} probe($u);}`, "parse_url($a)"},
		{"function result, null excluded, declared type", "origin", twoResults + `scandir($a);} if($u === null){return;} probe($u);}`, "scandir($a)"},
		{"function result, null excluded, fails with null", "origin-null", twoResults + `scandir($a);} if($u === null){return;} probe($u);}`, ""},
		{"function result, false excluded, fails with null", "origin-null", twoResults + `scandir($a);} if($u === false){return;} probe($u);}`, "parse_url($a)"},
		{"single result", "origin-false", `function f($a){$u=parse_url($a); probe($u);}`, "parse_url($a)"},
		// An inspection that names the calls it is about is not answered with
		// the unguarded result of another one.
		{"other result first", "origin-false", `function f($a, $c){$u=parse_url($a); if($c){$u=scandir($a);} probe($u);}`, "parse_url($a)"},
		{"other result first, passed over", "origin-of-scandir", `function f($a, $c){$u=parse_url($a); if($c){$u=scandir($a);} probe($u);}`, "scandir($a)"},
		{"other result first, own result guarded", "origin-of-scandir", `function f($a, $c){$u=parse_url($a); if($c){$u=scandir($a); if(!$u){return;}} probe($u);}`, ""},
		{"only other results", "origin-of-scandir", `function f($a, $c){$u=parse_url($a); if($c){$u=glob($a);} probe($u);}`, ""},
		{"known literal is unrelated", "origin-of-scandir", `function f(){$u='known'; probe($u);}`, ""},
		{"known matching result", "origin-of-scandir", `function f($a){$u=scandir($a); probe($u);}`, "scandir($a)"},
		{"literal before later matching result", "origin-of-scandir", `function f($a,$c){$u='known'; if($c){$u=scandir($a);} probe($u);}`, "scandir($a)"},
		{"single other result", "origin-of-scandir", `function f($a){$u=parse_url($a); probe($u);}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, tc.query, tc.src, ""); got != tc.want {
				t.Fatalf("origin=%q, want %q", got, tc.want)
			}
		})
	}
}

// A local resolves to its plain assignment only while no other write of the
// variable lies between that assignment and the read.
func TestSharedLocalValueSeesEveryWrite(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"plain", `$n=5; probe($n);`, "5"},
		{"temporary array argument", `function f(){$n=5; work(['sink'=>$n]); probe($n);}`, "5"},
		{"released foreach reference", `function f($a){foreach($a as &$n){} unset($n); $n=5; work(); probe($n);}`, "5"},
		{"conditional binding released", `function f($a,$c){if($c){foreach($a as &$n){} unset($n);} $n=5; work(); probe($n);}`, "5"},
		{"older alias survives conditional release", `function f($a,$c){$r=&$n;if($c){foreach($a as &$n){}unset($n);} $n=5;work();probe($n);}`, ""},

		{"binding escapes before release", `function f($a,$c){while($c){foreach($a as &$n){} break; unset($n);} $n=5; work(); probe($n);}`, ""},
		{"exception skips release", `function f($a){try{foreach($a as &$n){} work();unset($n);}catch(E){} $n=5;work();probe($n);}`, ""},

		{"conditional release", `function f($a,$c){foreach($a as &$n){} if($c){unset($n);} $n=5; work(); probe($n);}`, ""},
		{"nested release", `function f($a,$c){foreach($a as &$n){} if($c){unset($n); $n=5; work(); probe($n);}}`, "5"},
		{"unbraced conditional release", `function f($a,$c){foreach($a as &$n){} if($c) unset($n); $n=5; work(); probe($n);}`, ""},
		{"reference bound again", `function f($a){foreach($a as &$n){} unset($n); $r=&$n; $n=5; work(); probe($n);}`, ""},
		{"label may skip release", `function f($a,$c){foreach($a as &$n){} if($c) goto next; unset($n); next: $n=5; work(); probe($n);}`, ""},
		{"silenced call", `function f($a){$u=@parse_url($a); probe($u);}`, "parse_url($a)"},
		{"silenced call in parentheses", `function f($a){$u=(@(parse_url($a))); probe($u);}`, "parse_url($a)"},
		{"copy", `$m=5; $n=$m; probe($n);`, "5"},
		{"post increment", `function f(){$n=5; work(); $n++; probe($n);}`, ""},
		{"pre decrement", `function f(){$n=5; work(); --$n; probe($n);}`, ""},
		{"increment before assignment", `function f(){$n=1; $n++; $n=5; work(); probe($n);}`, "5"},
		{"increment after read", `function f(){$n=5; work(); probe($n); $n++;}`, "5"},
		{"destructuring", `function f($r){$n=5; work(); [$n, $m]=$r; probe($n);}`, ""},
		{"nested destructuring", `function f($r){$n=5; work(); [[$n], $m]=$r; probe($n);}`, ""},
		{"skipped destructuring slot", `function f($r){$n=5; work(); [, $m]=$r; probe($n);}`, "5"},
		{"list destructuring", `function f($r){$n=5; work(); list('k'=>$n)=$r; probe($n);}`, ""},
		{"list loop target", `function f($r){$n=5; work(); foreach($r as list($n)){} probe($n);}`, ""},
		{"reference list loop source", `function f(){$n=[[5]]; work(); foreach($n as list(&$v)){$v=1;} probe($n);}`, ""},
		{"loop value", `function f($r){$n=5; work(); foreach($r as $n){} probe($n);}`, ""},
		{"loop key", `function f($r){$n=5; work(); foreach($r as $n=>$v){} probe($n);}`, ""},
		{"reference loop source", `function f(){$n=[5]; work(); foreach($n as &$v){$v=1;} probe($n);}`, ""},
		{"reference loop source in target", `function f(){$n=[[5]]; work(); foreach($n as [&$v]){$v=1;} probe($n);}`, ""},
		{"nested reference loop source", `function f(){$n=[[[5]]]; work(); foreach($n as [[&$v]]){$v=1;} probe($n);}`, ""},
		{"value loop source", `function f(){$n=[5]; work(); foreach($n as $v){} probe($n);}`, "[5]"},
		{"catch binding", `function f(){$n=5; try{work();}catch(E $n){} probe($n);}`, ""},
		{"catch without variable", `function f(){$n=5; try{work();}catch(E){} probe($n);}`, "5"},
		{"global declaration", `function f(){$n=5; work(); global $n; probe($n);}`, ""},
		{"static declaration", `function f(){$n=5; work(); static $n; probe($n);}`, ""},
		{"unset", `function f(){$n=5; work(); unset($n); probe($n);}`, ""},
		{"reference alias", `function f(){$n=5; work(); $r=&$n; $r=1; probe($n);}`, ""},
		{"written through an earlier alias", `function f($a){$r=&$n; $n=5; $r=1; probe($n);}`, ""},
		{"earlier alias, nothing in between", `function f($a){$r=&$n; $n=5; probe($n);}`, "5"},
		{"earlier capture then call", `function f($a){$g=function() use (&$n){$n=1;}; $n=5; $g(); probe($n);}`, ""},
		{"earlier capture then builtin call", `function f($a){$g=function() use (&$n){$n=1;}; $n=5; strlen($a); probe($n);}`, "5"},
		{"earlier capture then method call", `function f($o){$g=function() use (&$n){$n=1;}; $n=5; $o->run(); probe($n);}`, ""},
		{"earlier capture then static call", `function f($o){$g=function() use (&$n){$n=1;}; $n=5; Runner::run(); probe($n);}`, ""},
		{"earlier capture then construction", `function f($o){$g=function() use (&$n){$n=1;}; $n=5; new Runner(); probe($n);}`, ""},
		{"global then call", `function f($o){global $n; $n=5; work(); probe($n);}`, ""},
		{"reference item then call", `function f($o){$n=5; $w=[&$n]; work(); probe($n);}`, ""},
		{"reference loop then write of the other name", `function f($o){$n=[5]; foreach($n as &$v){} $n=[6]; $v=1; probe($n);}`, ""},
		{"globals entry of the variable", `$n=5; work(); $GLOBALS['n']=6; probe($n);`, ""},
		{"globals entry of another variable", `$n=5; work(); $GLOBALS['m']=6; probe($n);`, "5"},
		{"dynamic globals entry", `$n=5; work(); $GLOBALS[$k]=6; probe($n);`, ""},
		{"local named like the table", `function f(){$GLOBALS=[]; $n=5; work(); probe($n);}`, "5"},
		{"earlier capture then callback builtin", `function f($a){$g=function() use (&$n){$n=1;}; $n=5; array_map($g, $a); probe($n);}`, ""},
		{"string offset write", `function f(){$n='abc'; work(); $n[0]='x'; probe($n);}`, ""},
		{"interpolated string offset write", `function f($a){$n="a$a"; work(); $n[0]='x'; probe($n);}`, ""},
		{"reference alias after read", `function f(){$n=5; work(); probe($n); $r=&$n;}`, "5"},
		{"reference item", `function f(){$n=5; work(); $w=[&$n]; probe($n);}`, ""},
		{"reference capture", `function f(){$n=5; work(); $g=function() use (&$n){$n=1;}; probe($n);}`, ""},
		{"value capture", `function f(){$n=5; work(); $g=function() use ($n){}; probe($n);}`, "5"},
		{"arrow function", `function f(){$n=5; work(); $g=fn() => $n++; probe($n);}`, "5"},
		{"builtin reference argument", `function f($s){$n=5; work(); sscanf($s, '%d', $n); probe($n);}`, ""},
		{"reference argument of the assigned call", `function f($s){$n=preg_match('/a/', $s, $n); probe($n);}`, "preg_match('/a/', $s, $n)"},
		{"builtin named reference argument", `function f($s){$n=5; work(); preg_match('/a/', $s, matches: $n); probe($n);}`, ""},
		{"builtin value argument", `function f($s){$n=5; work(); strlen($n); probe($n);}`, "5"},
		{"builtin named value argument", `function f($s){$n=5; work(); strlen(string: $n); probe($n);}`, "5"},
		{"builtin unknown named argument", `function f($s){$n=5; work(); strlen(text: $n); probe($n);}`, ""},
		{"builtin surplus argument", `function f($s){$n=5; work(); strlen($s, $n); probe($n);}`, "5"},
		{"builtin variadic value argument", `function f($s){$n=5; work(); max($s, 1, $n); probe($n);}`, "5"},
		{"builtin variadic reference argument", `function f($s){$n=5; work(); sscanf($s, '%d%d', $m, $n); probe($n);}`, ""},
		{"builtin unpacked argument", `function f($s){$n=[5]; work(); max(...$n); probe($n);}`, ""},
		{"call-time reference", `function f($s){$n=5; work(); strlen(&$n); probe($n);}`, ""},
		{"placeholder call", `function f($s){$n=5; work(); $g=strlen(...); probe($n);}`, "5"},
		{"callee without arguments", `function f($s){$n=5; work(); new stdClass; probe($n);}`, "5"},
		{"unresolved method argument", `function f($o){$n=5; work(); $o->fill($n); probe($n);}`, ""},
		{"reference method argument", `class W { function fill(&$x){} } function f(W $o){$n=5; work(); $o->fill($n); probe($n);}`, ""},
		{"value method argument", `class W { function fill($x){} } function f(W $o){$n=5; work(); $o->fill($n); probe($n);}`, "5"},
		{"unpacked method argument", `class W { function fill($x){} } function f(W $o){$n=[5]; work(); $o->fill(...$n); probe($n);}`, ""},
		{"builtin method value argument", `function f(DateTime $o){$n=5; work(); $o->format($n); probe($n);}`, "5"},
		{"unresolved static argument", `function f(){$n=5; work(); Unknown::fill($n); probe($n);}`, ""},
		{"reference static argument", `class W { static function fill(&$x){} } function f(){$n=5; work(); W::fill($n); probe($n);}`, ""},
		{"builtin static value argument", `function f(){$n=5; work(); DateTime::createFromFormat($n, 'x'); probe($n);}`, "5"},
		{"unresolved constructor argument", `function f(){$n=5; work(); new Unknown($n); probe($n);}`, ""},
		{"dynamic constructor argument", `function f($c){$n=5; work(); new $c($n); probe($n);}`, ""},
		{"reference constructor argument", `class W { function __construct(&$x){} } function f(){$n=5; work(); new W($n); probe($n);}`, ""},
		{"unresolved function argument", `function f(){$n=5; work(); if(fill($n)){} probe($n);}`, ""},
		{"builtin constructor value argument", `function f(){$n=5; work(); new DateTime($n); probe($n);}`, "5"},
		{"argument read by its own call", `class W { function __construct(&$x){} } function f(){$n=5; work(); probe($n, new W($n));}`, "5"},
		{"extract", `function f($r){$n=5; work(); extract($r); probe($n);}`, ""},
		{"parse_str into scope", `function f($r){$n=5; work(); parse_str($r); probe($n);}`, ""},
		{"parse_str into array", `function f($r){$n=5; work(); parse_str($r, $out); probe($n);}`, "5"},
		{"include", `function f(){$n=5; work(); include 'x.php'; probe($n);}`, ""},
		{"eval", `function f($c){$n=5; work(); eval($c); probe($n);}`, ""},
		{"variable variable write", `function f($k){$n=5; work(); $$k=1; probe($n);}`, ""},
		{"variable variable read", `function f($k){$n=5; work(); $m=$$k; probe($n);}`, "5"},
		{"element write of array literal", `function f(){$n=[5]; work(); $n[0]=1; probe($n);}`, ""},
		{"element step of array literal", `function f(){$n=[5]; work(); $n[0]++; probe($n);}`, ""},
		{"element write of object", `function f(){$n=new SplFixedArray(2); work(); $n[0]=1; probe($n);}`, "new SplFixedArray(2)"},
		{"property write", `function f(){$n=new stdClass(); work(); $n->p=1; probe($n);}`, "new stdClass()"},
		{"used as index", `function f(){$n=5; work(); $a[$n]=1; probe($n);}`, "5"},
		{"file level", `$n=5; work(); $n++; probe($n);`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, "origin", tc.src, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestSharedLocalWriteBudget(t *testing.T) {
	var src strings.Builder
	src.WriteString("function f(){$n=5; work();")
	for range nativeLocalWriteBudget + 1 {
		src.WriteString("$o++;")
	}
	src.WriteString("probe($n, $n);}")
	if got := sharedDefectResults(t, "origin", src.String(), ""); got != "" {
		t.Fatalf("origin beyond budget: %q", got)
	}
	// Calls that can run project code are budgeted like writes.
	src.Reset()
	src.WriteString("function f($o){$n=5; work();")
	for range nativeLocalWriteBudget + 1 {
		src.WriteString("new R;")
	}
	src.WriteString("probe($n);}")
	if got := sharedDefectResults(t, "origin", src.String(), ""); got != "" {
		t.Fatalf("origin beyond call budget: %q", got)
	}
	src.Reset()
	src.WriteString("function f($o){$t=$o->p; if($t !== false){")
	for range nativeLocalWriteBudget + 1 {
		src.WriteString("$u++;")
	}
	// The flow facts answer first where the scope is followed: code the
	// scope evaluates leaves the guard to the walk.
	src.WriteString("probe($t);} eval($o);}")
	if got := sharedDefectResults(t, "false", src.String(), ""); got != "false" {
		t.Fatalf("guard beyond budget: %q", got)
	}
	// A variable without a reaching value is identified by name alone.
	src.Reset()
	for range nativeLocalWriteBudget + 1 {
		src.WriteString("$o++;")
	}
	src.WriteString("fclose($n); fread($n, 1);")
	if sharedDefectStable(t, src.String()) {
		t.Fatal("name identity beyond budget")
	}
	if !sharedDefectStable(t, "fclose($n); fread($n, 1);") {
		t.Fatal("name identity within budget")
	}
}

// Two reads denote the same value only when nothing rebinds the variable
// between them; using it as an index of another write does not.
func TestSharedIdentityAcrossWrites(t *testing.T) {
	for _, tc := range []struct{ query, name, src, want string }{
		{"stable", "untouched parameter", `function f($n){work($m); probe($n, $n);}`, "true"},
		{"stable", "destructured between", `function f($n, $r){mark($n); [$n]=$r; probe($n, $n);} `, "true"},
		{"stable", "destructured", `function f($n, $r){$a=$n; [$n]=$r; probe($a, $n);}`, "false"},
		{"stable", "increment between reads", `function f($n){$x=keep($m) + $n; $n++; probe($n, $n);}`, "true"},
		{"same", "increment", `function f(){$n=5; $a=$n; $n++; probe($a, $n);}`, "false"},
		{"same", "copy", `function f(){$n=new stdClass(); $a=$n; probe($a, $n);}`, "true"},
		{"same", "loop target", `function f($r){$n=new stdClass(); $a=$n; foreach($r as $n){} probe($a, $n);}`, "false"},
		{"same", "builtin output argument", `function f($s){$n=new stdClass(); $a=$n; preg_match('/a/', $s, $n); probe($a, $n);}`, "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, tc.query, tc.src, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
	// expansionDSame compares two different reads of one name.
	for _, tc := range []struct {
		name, between string
		want          bool
	}{
		{"nothing", ``, true},
		{"index of another write", `$a[$n]=1;`, true},
		{"property name of another write", `$o->{$n}=1;`, true},
		{"unset of another element", `unset($a[$n]);`, true},
		{"plain write", `$n=other();`, false},
		{"element write", `$n[0]=1;`, true},
		{"destructuring", `[$n]=$r;`, false},
		{"loop key", `foreach($r as $n=>$v){}`, false},
		{"increment", `$n++;`, false},
		{"unset", `unset($n);`, false},
		{"reference alias", `$q=&$n;`, false},
		{"reference capture", `$g=function() use (&$n){};`, false},
		{"extract", `extract($r);`, false},
		{"variable variable", `$$k=1;`, false},
		{"builtin updating its handle", `sodium_crypto_secretstream_xchacha20poly1305_push($n, 'a');`, true},
		{"unknown function", `fill($n);`, false},
		{"unknown function reading it", `fill($n + 1);`, false},
		{"unknown method", `$o->fill($n);`, false},
		{"unknown method reading it", `$o->fill($n + 1);`, false},
		{"variable variable from a builtin", `preg_match('/a/', 'a', $$k);`, true},
	} {
		t.Run("between "+tc.name, func(t *testing.T) {
			src := `function f($n, $r, $a, $o, $k){fclose($n); ` + tc.between + ` fread($n, 1);}`
			if got := sharedDefectStable(t, src); got != tc.want {
				t.Fatalf("stable=%v", got)
			}
		})
	}
}

// sharedDefectStable applies expansionDSame to the handles of fclose() and fread().
func sharedDefectStable(t testing.TB, src string) bool {
	t.Helper()
	result := false
	r := sharedStableProbe{&result}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{Only: []string{r.ID()}})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("stable.php", []byte("<?php "+src), syntax.Options{})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %s: %v", src, f.Errors)
	}
	e.Analyze(f)
	return result
}

type sharedStableProbe struct{ result *bool }

func (sharedStableProbe) ID() string               { return "AbstractClassInstantiation" }
func (sharedStableProbe) Semantic()                {}
func (sharedStableProbe) Flow()                    {}
func (sharedStableProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (p sharedStableProbe) Check(ctx *analysis.Context, n syntax.Node) {
	second := n.(*syntax.FuncCall)
	if ctx.Text(second.Name) != "fread" {
		return
	}
	for _, event := range nativeContractEvents(ctx, syntax.EnclosingVariableScope(second)) {
		if first, ok := event.(*syntax.FuncCall); ok && ctx.Text(first.Name) == "fclose" {
			*p.result = expansionDSame(ctx, CallArgument(first.Args, 0, ""), CallArgument(second.Args, 0, ""))
		}
	}
}

func TestSharedFailureGuardPositions(t *testing.T) {
	for _, tc := range []struct{ query, name, src, want string }{
		{"false", "unguarded", `function f($p){$x=fopen($p,'r'); probe($x);}`, "false"},
		{"false", "no input", `function f($p){probe();}`, "false"},
		{"false", "strict branch", `function f($p){$x=fopen($p,'r'); if($x !== false){probe($x);}}`, "true"},
		{"false", "truthy exit", `function f($p){$x=fopen($p,'r'); if(!$x){return;} probe($x);}`, "true"},
		{"false", "earlier exit", `function f($p){$x=fopen($p,'r'); if($x === false){return;} echo 1; probe($x);}`, "true"},
		{"false", "ternary arm", `function f($p){$x=fopen($p,'r'); return $x === false ? 0 : probe($x);}`, "true"},
		// Property reads have no flow value: type narrowing proves them, on a
		// receiver whose class declares the property.
		{"false", "property branch", `final class H { public $h; public $g; } function f(H $o){if($o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property constant-first branch", `final class H { public $h; public $g; } function f(H $o){if(false !== $o->h){probe($o->h);}}`, "true"},
		{"false", "property else arm", `final class H { public $h; public $g; } function f(H $o){if($o->h === false){return 0;}else{probe($o->h);}}`, "true"},
		{"false", "property wrong arm", `final class H { public $h; public $g; } function f(H $o){if($o->h !== false){return 0;}else{probe($o->h);}}`, "false"},
		{"false", "property elseif arm", `final class H { public $h; public $g; } function f(H $o, $c){if($c){return 0;}elseif($o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property after elseif", `final class H { public $h; public $g; } function f(H $o, $c){if($c){return 0;}elseif($o->h === false){return 1;}else{probe($o->h);}}`, "true"},
		{"false", "property after first test", `final class H { public $h; public $g; } function f(H $o, $c){if($o->h === false){return 0;}elseif($c){probe($o->h);}}`, "true"},
		{"false", "property later elseif", `final class H { public $h; public $g; } function f(H $o, $c, $d){if($c){return 0;}elseif($o->h === false){return 1;}elseif($d){probe($o->h);}}`, "true"},
		{"false", "property unrelated elseif", `final class H { public $h; public $g; } function f(H $o, $c, $d){if($c){return 0;}elseif($d){probe($o->h);}elseif($o->h === false){return 1;}}`, "false"},
		{"false", "property spaced else if", `final class H { public $h; public $g; } function f(H $o, $c){if($o->h === false){return 0;}else if($c){probe($o->h);}}`, "true"},
		{"false", "property conjunction operand", `final class H { public $h; public $g; } function f(H $o){return $o->h !== false && probe($o->h);}`, "true"},
		{"false", "property word conjunction operand", `final class H { public $h; public $g; } function f(H $o){return ($o->h !== false and probe($o->h));}`, "true"},
		{"false", "property disjunction operand", `final class H { public $h; public $g; } function f(H $o){return $o->h === false || probe($o->h);}`, "true"},
		{"false", "property word disjunction operand", `final class H { public $h; public $g; } function f(H $o){return ($o->h === false or probe($o->h));}`, "true"},
		{"false", "property wrong disjunction", `final class H { public $h; public $g; } function f(H $o){return $o->h !== false || probe($o->h);}`, "false"},
		{"false", "property left operand", `final class H { public $h; public $g; } function f(H $o){return probe($o->h) && $o->h !== false;}`, "false"},
		{"false", "property other operator", `final class H { public $h; public $g; } function f(H $o){return ($o->h !== false) + probe($o->h);}`, "false"},
		{"false", "property ternary arm", `final class H { public $h; public $g; } function f(H $o){return $o->h !== false ? probe($o->h) : 0;}`, "true"},
		{"false", "property ternary else arm", `final class H { public $h; public $g; } function f(H $o){return $o->h === false ? 0 : probe($o->h);}`, "true"},
		{"false", "property wrong ternary arm", `final class H { public $h; public $g; } function f(H $o){return $o->h !== false ? 0 : probe($o->h);}`, "false"},
		{"false", "property short ternary", `final class H { public $h; public $g; } function f(H $o){return $o->h === false ?: probe($o->h);}`, "true"},
		{"false", "property ternary condition", `final class H { public $h; public $g; } function f(H $o){return probe($o->h) ? $o->h !== false : 0;}`, "false"},
		{"false", "property negated guard", `final class H { public $h; public $g; } function f(H $o){if(!($o->h === false)){probe($o->h);}}`, "true"},
		{"false", "property conjunction guard", `final class H { public $h; public $g; } function f(H $o, $c){if($c && $o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property false conjunction", `final class H { public $h; public $g; } function f(H $o, $c){if($c && $o->h === false){return 0;} probe($o->h);}`, "false"},
		{"false", "property disjunction exit", `final class H { public $h; public $g; } function f(H $o, $c){if($c || $o->h === false){return 0;} probe($o->h);}`, "true"},
		{"false", "property true disjunction", `final class H { public $h; public $g; } function f(H $o, $c){if($c || $o->h !== false){probe($o->h);}}`, "false"},
		{"false", "property preceding exit", `final class H { public $h; public $g; } function f(H $o){if($o->h === false){return 0;} probe($o->h);}`, "true"},
		{"false", "property other value", `final class H { public $h; public $g; } function f(H $o){if($o->g === false){return 0;} probe($o->h);}`, "false"},
		{"false", "repeated call", `function f($o){if($o->h() === false){return 0;} probe($o->h());}`, "false"},
		// A parameter has no reaching producer: the guard holds only while
		// nothing writes the variable between the test and the read.
		{"false", "untracked variable branch", `function f($s, $o){$t=$o->p; if($t !== false){probe($t);}}`, "true"},
		{"false", "untracked variable rewritten in the branch", `function f($s, $o){$t=$o->p; if($t !== false){$t=$o->q; probe($t);}}`, "false"},
		{"false", "untracked variable stepped in the branch", `function f($s, $o){$t=$o->p; if($t !== false){$t++; probe($t);}}`, "false"},
		{"false", "untracked variable variable", `function f($s, $o){if($$s !== false){probe($$s);}}`, "false"},
		{"false", "stepping operand", `function f($i){if($i++ !== false){probe($i++);}}`, "false"},
		{"false", "closure boundary", `final class H { public $h; public $g; } function f(H $o){if($o->h === false){return 0;} $g=function() use ($o){probe($o->h);};}`, "false"},
		{"local-false", "property branch", `function f($o){if(\FALSE !== ($o->h)){probe($o->h);}}`, "false"},
		{"local-false", "callable branch", `function f($o){$g=$o->pick(); if(\FALSE !== ($g)){probe($g);}}`, "true"},
		{"length", "no input", `function f($s){probe();}`, "false"},
		{"length", "flow fact", `function f($s){if(strlen($s) !== 4){return;} echo 1; probe($s);}`, "true"},
		{"length", "property branch", `final class H { public $s; } function f(H $o){if(strlen($o->s) === 4){probe($o->s);}}`, "true"},
		{"length", "property constant-first", `final class H { public $s; } function f(H $o){if(4 === strlen($o->s)){probe($o->s);}}`, "true"},
		{"length", "property other length", `final class H { public $s; } function f(H $o){if(strlen($o->s) === 5){probe($o->s);}}`, "false"},
		{"length", "property ternary", `final class H { public $s; } function f(H $o){return strlen($o->s) === 4 ? probe($o->s) : 0;}`, "true"},
		{"length", "property other test", `final class H { public $s; } function f(H $o){if(count($o->s) === 4){probe($o->s);}}`, "false"},
		{"length", "property loose test", `final class H { public $s; } function f(H $o){if(strlen($o->s) == 4){probe($o->s);}}`, "true"},
		{"length", "property elseif arm", `final class H { public $s; } function f(H $o, $c){if($c){return 0;}elseif(strlen($o->s) === 4){probe($o->s);}}`, "true"},
		{"length", "property else arm", `final class H { public $s; } function f(H $o){if(strlen($o->s) !== 4){return 0;}else{probe($o->s);}}`, "true"},
		{"length", "property after failed elseif", `final class H { public $s; } function f(H $o, $c){if($c){return 0;}elseif(strlen($o->s) !== 4){return 1;}else{probe($o->s);}}`, "true"},
		{"length", "property conjunction operand", `final class H { public $s; } function f(H $o){return strlen($o->s) === 4 && probe($o->s);}`, "true"},
		{"length", "property disjunction operand", `final class H { public $s; } function f(H $o){return strlen($o->s) !== 4 || probe($o->s);}`, "true"},
		{"length", "call result", `function f($o){if(strlen($o->s()) === 4){probe($o->s());}}`, "false"},
		{"length", "unknown writes", `function f($s, $a){if(strlen($s) === 4){extract($a); probe($s);}}`, "false"},
	} {
		t.Run(tc.query+" "+tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, tc.query, tc.src, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestSharedSelfInstantiation(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"named class", `abstract class A { function m(){ probe(new self()); } }`, "A"},
		{"namespaced class", `namespace App; abstract class A { function m(){ probe(new SELF); } }`, `App\A`},
		{"enum", `enum E { case X; function m(){ probe(new self()); } }`, "E"},
		{"other class", `class B {} class A { function m(){ probe(new B()); } }`, "B"},
		{"static", `abstract class A { function m(){ probe(new static()); } }`, ""},
		{"trait", `trait T { function m(){ probe(new self()); } }`, ""},
		{"anonymous class", `$o = new class { function m(){ probe(new self()); } };`, ""},
		{"closure", `abstract class A { function m(){ return function(){ probe(new self()); }; } }`, ""},
		{"arrow function", `abstract class A { function m(){ return fn() => probe(new self()); } }`, ""},
		{"function", `function f(){ probe(new self()); }`, ""},
		{"file level", `probe(new self());`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, "class", tc.src, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestSharedIntegerConstants(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// A known integer stepped on the only path to the read.
		{`$n=-1; $n--; probe($n);`, "-2/true"},
		{`$n=-1; ++$n; probe($n);`, "0/true"},
		{`$n=5; $n++; $n++; $n--; probe($n);`, "6/true"},
		{`$n=5; $n+=3; probe($n);`, "8/true"},
		{`$n=5; $n-=7; probe($n);`, "-2/true"},
		{`$n=5; $m=2; $n+=$m; $n++; probe($n);`, "8/true"},
		{`$n=5; $n*=2; probe($n);`, "0/false"},
		{`$n=5; $n+=$u; probe($n);`, "0/false"},
		{`$n=5; $n-=$u; probe($n);`, "0/false"},
		{`$n='a'; $n++; probe($n);`, "0/false"},
		{`$n=5; if($c){$n++;} probe($n);`, "0/false"},
		{`$n=5; while($c){probe($n); $n++;}`, "0/false"},
		{`function f($n){$n++; probe($n);}`, "0/false"},
		{`$n=9223372036854775807; $n++; probe($n);`, "0/false"},
		{`$n=-9223372036854775807; $n-=5; probe($n);`, "0/false"},
		{`$n=5; $n-=PHP_INT_MIN; probe($n);`, "0/false"},
		{`$n=5; probe($n++);`, "0/false"},
		{`$n=5; probe(++$n);`, "0/false"},
		{`$n=0;` + strings.Repeat(`$n++;`, nativeStepBudget) + `probe($n);`, "16/true"},
		{`$n=0;` + strings.Repeat(`$n++;`, nativeStepBudget+1) + `probe($n);`, "0/false"},
		{`probe(ZLIB_ENCODING_RAW);`, "-15/true"},
		{`probe(ZLIB_ENCODING_GZIP);`, "31/true"},
		{`probe(PHP_EOL);`, "0/false"},
	} {
		if got := sharedDefectResults(t, "int", tc.src, ""); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.src, got, tc.want)
		}
	}
}

func TestSharedTruthinessOwners(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`if(probe()){}`, "true"},
		{`if($c){}elseif(probe()){}`, "true"},
		{`if($c){}elseif(!probe()){}`, "true"},
		{`if($x = probe()){}`, "true"},
		{`if($x .= probe()){}`, "false"},
		{`if($x ??= probe()){}`, "false"},
		{`while(probe()){}`, "true"},
		{`do{}while(probe());`, "true"},
		{`$x = probe();`, "false"},
	} {
		if got := sharedDefectResults(t, "truthy", tc.src, ""); got != tc.want {
			t.Errorf("%s: got %q want %q", tc.src, got, tc.want)
		}
	}
}

func TestSharedDominanceSkipsConditionalEvaluation(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"statement", `mark(); probe();`, "1"},
		{"match arm", `$x = match($k){1 => mark(), default => 0}; probe();`, "0"},
		{"match subject", `$x = match(mark()){1 => 1, default => 0}; probe();`, "1"},
		{"same match arm", `$x = match($k){default => [mark(), probe()]};`, "1"},
		{"other match arm", `$x = match($k){1 => mark(), default => probe()};`, "0"},
		{"null-safe argument", `$d?->format(mark()); probe();`, "0"},
		{"null-safe chained argument", `$d?->a->format(mark()); probe();`, "0"},
		{"null-safe property chain argument", `$d?->a['k']->format(mark()); probe();`, "0"},
		{"null-safe call chain argument", `$d?->m()->a->format(mark()); probe();`, "0"},
		{"null-safe dynamic property name", `$y = $d?->{mark()}; probe();`, "0"},
		{"null-safe chained dynamic property name", `$y = $d?->m()->{mark()}; probe();`, "0"},
		{"same null-safe dynamic property name", `$y = $d?->{mark() . probe()};`, "1"},
		{"plain dynamic property name", `$y = $d->{mark()}; probe();`, "1"},
		{"null-safe index", `$y = $d?->a[mark()]; probe();`, "0"},
		{"null-safe parenthesised receiver", `$y = ($d?->a)[mark()]; probe();`, "0"},
		{"null-safe receiver", `mark()?->format(1); probe();`, "1"},
		{"same null-safe arguments", `$d?->format(mark(), probe());`, "1"},
		{"same null-safe index", `$y = $d?->a[mark() + probe()];`, "1"},
		{"plain argument", `$d->format(mark()); probe();`, "1"},
		{"plain index", `$y = $a[mark()]; probe();`, "1"},
		{"static receiver index", `$y = A::$a[mark()]; probe();`, "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, "dominated", tc.src, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func BenchmarkSharedLocalWrites(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php function f($r, $s, $o){$n=new SplFixedArray(4);")
	for i := range 60 {
		fmt.Fprintf(&src, "$v%d=1; $v%d++; [$a%d, $b%d]=$r; foreach($r as $k%d=>$w%d){} preg_match('/a/', $s, $m%d); $o->fill($v%d); $n[%d]=1;", i, i, i, i, i, i, i, i, i%4)
	}
	src.WriteString("probe($n);}")
	text := []byte(src.String())
	var got []string
	r := sharedDefectProbe{"origin", &got}
	e, err := analysis.NewEngine([]analysis.Rule{r}, analysis.Config{Only: []string{r.ID()}})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", text, syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		got = got[:0]
		e.Analyze(f)
	}
	if len(got) != 1 || got[0] == "" {
		b.Fatalf("origin=%v", got)
	}
}

// A call that can enter the function again assigns its static locals: what
// was established of one before the call says nothing of a read after it,
// even though the lexical scan still names the same producer.
func TestSharedStaticLocalAcrossReentry(t *testing.T) {
	const static = `function f($a, $o, $c){static $s; $s=parse_url($a); `
	var crowded strings.Builder
	crowded.WriteString("function f($s){")
	for range nativeLocalWriteBudget + 1 {
		crowded.WriteString("$u++;")
	}
	crowded.WriteString("probe($s);}")
	for _, tc := range []struct{ query, name, src, want string }{
		{"static", "declared", static + `probe($s);}`, "true"},
		{"static", "other local", static + `probe($a);}`, "false"},
		{"static", "not a variable", static + `probe($a->p);}`, "false"},
		{"static", "computed name beside a static", static + `probe($$a);}`, "true"},
		{"static", "computed name, no static", `function f($a){probe($$a);}`, "false"},
		{"static", "scope beyond the budget", crowded.String(), "true"},
		{"static", "declared in a nested function", `function f($s){$g=function(){static $s;}; probe($s);}`, "false"},

		{"same", "copies around nothing", static + `$t=$s; $u=$s; probe($t, $u);}`, "true"},
		{"same", "copies around a call", static + `$t=$s; $o->run(); $u=$s; probe($t, $u);}`, "false"},
		{"same", "copies around a call on one path", static + `$t=$s; if($c){$o->run();} $u=$s; probe($t, $u);}`, "false"},
		{"same", "copies after a call on one path", static + `if($c){$o->run();} $t=$s; $u=$s; probe($t, $u);}`, "false"},
		{"same", "copies around a builtin", static + `$t=$s; strlen('a'); $u=$s; probe($t, $u);}`, "true"},
		{"same", "copies around a conversion to text", static + `$t=$s; strlen($a); $u=$s; probe($t, $u);}`, "false"},
		{"same", "plain local around a call", `function f($a, $o){$s=parse_url($a); $t=$s; $o->run(); $u=$s; probe($t, $u);}`, "true"},
		{"same", "object around a call it is handed to", `function f($o){$s=new stdClass(); $t=$s; $o->run($s); probe($t, $s);}`, "false"},
		{"same", "object handed to a call on one path", `function f($o, $c){$s=new stdClass(); $t=$s; if($c){$o->run($t);} probe($t, $s);}`, "false"},
		{"same", "static object around a call", `function f($o){static $s; $s=new stdClass(); $t=$s; $o->run(); probe($t, $s);}`, "false"},

		{"local-false", "tested, then read", static + `if($s !== false){probe($s);}}`, "true"},
		{"local-false", "tested, call, then read", static + `if($s !== false){$o->run(); probe($s);}}`, "false"},
		{"local-false", "tested, call in a loop with the read", static + `if($s !== false){foreach($c as $e){probe($s); $o->run();}}}`, "false"},
		{"local-false", "plain local tested, call, then read", `function f($a, $o){$s=parse_url($a); if($s !== false){$o->run(); probe($s);}}`, "true"},
	} {
		t.Run(tc.query+"/"+tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, tc.query, tc.src, ""); got != tc.want {
				t.Fatalf("%s=%q, want %q", tc.query, got, tc.want)
			}
		})
	}
}

// An inspection that needs the one value a variable holds is not answered
// with one of several producers, whichever is written first.
func TestSharedLocalOnly(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"one assignment", `function f($a){$u=scandir($a); probe($u);}`, "scandir($a)"},
		{"two branches", `function f($a, $c){if($c){$u=scandir($a);}else{$u=glob($a);} probe($u);}`, ""},
		{"two branches, literal second", `function f($a, $c){if($c){$u=scandir($a);}else{$u=[];} probe($u);}`, ""},
		{"one path only", `function f($a, $c){$u=[]; if($c){$u=scandir($a);} probe($u);}`, ""},
		{"one path, the other unknown", `function f($a, $c, $u){if($c){$u=scandir($a);} probe($u);}`, ""},
		{"loop back-edge", `function f($a, $c){$u=scandir($a); while($c--){probe($u); $u=[];}}`, ""},
		{"same assignment on every pass", `function f($a, $c){while($c--){$u=scandir($a); probe($u);}}`, "scandir($a)"},
		{"copy", `function f($a){$u=scandir($a); $v=$u; probe($v);}`, "scandir($a)"},
		{"not a variable", `function f($a){probe(scandir($a));}`, "scandir($a)"},
		{"parameter", `function f($a){probe($a);}`, ""},
		{"scope the flow does not follow", `function f($a){$u=scandir($a); probe($u); yield 1;}`, "scandir($a)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sharedDefectResults(t, "only", tc.src, ""); got != tc.want {
				t.Fatalf("only=%q, want %q", got, tc.want)
			}
			if tc.want == "" {
				return
			}
			if got := sharedDefectResults(t, "origin", tc.src, ""); got != tc.want {
				t.Fatalf("origin=%q, want %q", got, tc.want)
			}
		})
	}
}

// Dense unrelated writes must not make each known nonmatching result scan
// the scope's preceding events. Keep both literal and unrelated API cases.
func BenchmarkNativeLocalNonmatchingResults(b *testing.B) {
	for _, value := range []string{"'known'", "parse_url($a)"} {
		var source strings.Builder
		source.WriteString("<?php function f(string $a){")
		for range 128 {
			source.WriteString("$u=" + value + ";probe($u);")
		}
		source.WriteString("}")
		file := syntax.Parse("dense.php", []byte(source.String()), syntax.Options{})
		var results []string
		rule := sharedDefectProbe{"origin-of-scandir", &results}
		engine, err := analysis.NewEngine([]analysis.Rule{rule}, analysis.Config{Only: []string{rule.ID()}})
		if err != nil {
			b.Fatal(err)
		}
		b.Run(value, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				results = results[:0]
				engine.Analyze(file)
			}
		})
	}
}

func TestSharedDerivedStringGuards(t *testing.T) {
	for _, tc := range []struct {
		name, condition, between string
		want                     string
	}{
		{"suffix", "str_ends_with($v, '/')", "", "true"},
		{"prefix", "str_starts_with($v, '/')", "", "true"},
		{"contains", "str_contains($v, '/')", "", "true"},
		{"empty needle", "str_ends_with($v, '')", "", "false"},
		{"position", "strpos($v, '$') !== FALSE", "", "true"},
		{"empty position needle", "strpos($v, '') !== false", "", "false"},
		{"strict offset", "$v[0] === 'x'", "", "true"},
		{"empty offset", "$v[0] === ''", "", "true"}, // null[0] is null, never ''

		{"different variable", "str_ends_with($other, '/')", "", "false"},
		{"rewrite", "str_ends_with($v, '/')", "$v=other();", "false"},
		{"false branch", "!str_ends_with($v, '/')", "", "false"},
		{"unpacked", "str_ends_with(...$v)", "", "false"},
		{"named", "str_ends_with(needle:'/', haystack:$v)", "", "true"},
		{"duplicate", "str_ends_with($v, '/', haystack:$v)", "", "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(" + tc.condition + "){" + tc.between + "probe($v);}}"
			if got := sharedDefectResults(t, "string-nonnull", source, ""); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
	for _, source := range []string{
		`function f($s,$c){$v=preg_replace('/x/','',$s);if(str_ends_with($v,'/')){while($c){probe($v);$v=preg_replace('/y/','',$s);}}}`,
		`function f($s,$c){$v=preg_replace('/x/','',$s);while($c){if(str_ends_with($v,'/')){again: probe($v);$v=preg_replace('/y/','',$s);goto again;}}}`,
		`namespace N; function str_ends_with($v,$n){return true;} function f($s){$v=preg_replace('/x/','',$s);if(str_ends_with($v,'/')){probe($v);}}`,
	} {
		if got := sharedDefectResults(t, "string-nonnull", source, ""); got != "false" {
			t.Fatalf("unexpected guard: %q", got)
		}
	}
}

func TestSharedJoinedWarningBoundary(t *testing.T) {
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
			if got := sharedDefectResults(t, "warning-unknown", source, ""); got != fmt.Sprint(tc.want) {
				t.Fatalf("got %q want %v", got, tc.want)
			}
		})
	}
	for _, source := range []string{
		`set_error_handler(fn()=>throw new E());probe(fopen('x','r'));`,
		`function f($c){while($c){probe(fopen('x','r'));set_error_handler(fn()=>throw new E());}}`,
		`function f(){again: probe(fopen('x','r'));set_error_handler(fn()=>throw new E());goto again;}`,
		"function f(){" + strings.Repeat("$p='x';", 2050) + "probe(fopen('x','r'));}",
	} {
		if got := sharedDefectResults(t, "warning-unknown", source, ""); got != "true" {
			t.Fatalf("unexpected boundary %q", got)
		}
	}
}

func TestSharedDerivedStringGuardBoundaries(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		// substr() returns a string, whatever the guard tested.
		{`function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(false===$i=strpos($v,'$')){return;}$n=1;$v=substr($v,$i);probe($v);}`, "true"},
		{`function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(false===$i=strpos($v,'$')){return;}$n=1;$v=other();probe($v);}`, "false"},
		{`function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(false===$i=strpos($v,'$')){return;}$n=1;probe($v);}`, "true"},
		{`function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(str_ends_with($v,'/')){$v=substr(probe($v),1);}}`, "true"},
		{`function f($o){if(str_ends_with($o->p,'/')){probe($o->p);}}`, "false"},
		{`function f($o,$k){if(str_ends_with($$k,'/')){probe($$k);}}`, "false"},
		{`function f($v){if(str_ends_with($v,'/')){$v='changed';probe($v);}}`, "true"}, // a literal
		{`function f($v){if(str_ends_with($v,'/')){$v=other();probe($v);}}`, "false"},
		{"function f($v){" + strings.Repeat("$x=1;", 513) + "probe($v);}", "false"},
	} {
		if got := sharedDefectResults(t, "string-nonnull", tc.source, ""); got != tc.want {
			t.Fatalf("got %q want %q for %s", got, tc.want, tc.source)
		}
	}
	// An unknown write history retains no reference.
	writes := nativeLocalWrites{}
	writes.share(&syntax.Variable{Name: "v"}, 1)
	if len(writes.shared) != 0 {
		t.Fatal("exhausted history retained a reference")
	}
}

func BenchmarkNativeJoinedFailureGuards(b *testing.B) {
	for _, tc := range []struct{ name, query, source string }{
		{"derived", "string-nonnull", `<?php function f($s,$c){$v=$s;if($c){$v=preg_replace('/x/','',$v);}if(false===$i=strpos($v,'$'))return;$n=1;probe($v);}`},
		{"handler", "warning-unknown", `<?php function f($p){set_error_handler(fn()=>throw new E());probe(fopen($p,'r'));}`},
	} {
		b.Run(tc.name, func(b *testing.B) {
			file := syntax.Parse("guard.php", []byte(tc.source), syntax.Options{})
			var results []string
			rule := sharedDefectProbe{tc.query, &results}
			engine, err := analysis.NewEngine([]analysis.Rule{rule}, analysis.Config{Only: []string{rule.ID()}})
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			for b.Loop() {
				results = results[:0]
				engine.Analyze(file)
			}
		})
	}
}

func TestSharedReferenceWorkBudget(t *testing.T) {
	for _, tc := range []struct {
		bindings int
		want     string
	}{{20, "5"}, {160, ""}} {
		source := "function f(){" + strings.Repeat("$alias=&$n;unset($n);", tc.bindings) + "$n=5;work();probe($n);}"
		if got := sharedDefectResults(t, "origin", source, ""); got != tc.want {
			t.Fatalf("bindings%d got %q want %q", tc.bindings, got, tc.want)
		}
	}
	for _, tc := range []struct {
		jumps int
		want  string
	}{{nativeLocalWriteBudget, "true"}, {nativeLocalWriteBudget + 1, "false"}} {
		source := "function f($c,$n){while($c){" + strings.Repeat("if($c){break;}", tc.jumps) + "}probe($n);}"
		if got := sharedDefectResults(t, "writes-known", source, ""); got != tc.want {
			t.Fatalf("jumps%d got %q want %q", tc.jumps, got, tc.want)
		}
	}
	unknown := nativeLocalWrites{}
	if !unknown.reached("n", 0, 100) {
		t.Fatal("unknown write history proved isolated")
	}
	// All nested walks spend one allowance; running out inside either operand
	// must not let the other operand turn the result into a proof of isolation.
	for _, tc := range []struct {
		name     string
		writes   int
		calls    int
		starts   int
		releases int
		jumps    int
	}{
		{"outer writes", nativeLocalWriteBudget, 0, 3, 0, 0},
		{"outer calls", nativeLocalWriteBudget, 1, 3, 0, 0},
		{"starts", nativeLocalWriteBudget, 1, nativeLocalWriteBudget, 0, 0},
		{"releases", nativeLocalWriteBudget, 1, 4, nativeLocalWriteBudget - 4, 0},
		{"jumps", 5, 1, 4, 1, nativeLocalWriteBudget},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := nativeLocalWrites{known: true, shared: map[string][]uint32{"n": make([]uint32, tc.starts)}, released: map[string][]nativeLocalRelease{"n": make([]nativeLocalRelease, tc.releases)}, writes: make([]nativeLocalWrite, tc.writes), calls: make([]uint32, tc.calls), jumps: make([]uint32, tc.jumps)}
			for i := range writes.writes {
				writes.writes[i] = nativeLocalWrite{name: "other", at: 50}
			}
			for i := range writes.calls {
				writes.calls[i] = 50
			}
			for i := range writes.shared["n"] {
				writes.shared["n"][i] = 100
			}
			if tc.releases > 0 {
				for i := range writes.shared["n"] {
					writes.shared["n"][i] = uint32(i + 1)
				}
				for i := range writes.released["n"] {
					writes.released["n"][i].at = uint32(1000 + i)
				}
			}
			if tc.jumps > 0 {
				writes.released["n"][0] = nativeLocalRelease{at: 20, owner: syntax.Span{Start: 0, End: 30}}
				for i := range writes.jumps {
					writes.jumps[i] = uint32(1000 + i)
				}
			}
			if tc.name == "outer writes" {
				for i := range writes.shared["n"] {
					writes.shared["n"][i] = 50
				}
				writes.shared["heavy"] = []uint32{200, 200, 200, 200, 200, 200}
				writes.writes[0] = nativeLocalWrite{name: "heavy", at: 60}
				for i := 1; i < len(writes.writes); i++ {
					writes.writes[i].at = 40
				}
			}
			if !writes.reached("n", 0, 100) {
				t.Fatal("exhausted query proved isolation")
			}
		})
	}
}

func BenchmarkNativeReferenceLifetimes(b *testing.B) {
	for _, bindings := range []int{1, 160} {
		b.Run(fmt.Sprint(bindings), func(b *testing.B) {
			writes := nativeLocalWrites{known: true, shared: map[string][]uint32{"n": {}}, released: map[string][]nativeLocalRelease{}, calls: []uint32{10000}}
			for i := range bindings {
				writes.shared["n"] = append(writes.shared["n"], uint32(i*4+1))
				writes.released["n"] = append(writes.released["n"], nativeLocalRelease{at: uint32(i*4 + 2), owner: syntax.Span{Start: 0, End: 20000}})
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_ = writes.reached("n", 0, 20000)
			}
		})
	}
}
