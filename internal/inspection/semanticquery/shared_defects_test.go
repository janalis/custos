package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
	case "same":
		result = fmt.Sprint(NativeSameValue(ctx, a, b))
	case "stable":
		result = fmt.Sprint(expansionDSame(ctx, a, b))
	case "false":
		result = fmt.Sprint(NativeSentinelGuard(ctx, a, "false"))
	case "local-false":
		result = fmt.Sprint(NativeLocalSentinelGuard(ctx, a, "false"))
	case "length":
		result = fmt.Sprint(NativeLengthGuard(ctx, a, 4))
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

// A local resolves to its plain assignment only while no other write of the
// variable lies between that assignment and the read.
func TestSharedLocalValueSeesEveryWrite(t *testing.T) {
	for _, tc := range []struct{ name, src, want string }{
		{"plain", `$n=5; probe($n);`, "5"},
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
		{"unresolved method argument", `function f($o){$n=5; work(); $o->fill($n); probe($n);}`, "5"},
		{"reference method argument", `class W { function fill(&$x){} } function f(W $o){$n=5; work(); $o->fill($n); probe($n);}`, ""},
		{"value method argument", `class W { function fill($x){} } function f(W $o){$n=5; work(); $o->fill($n); probe($n);}`, "5"},
		{"unpacked method argument", `class W { function fill($x){} } function f(W $o){$n=[5]; work(); $o->fill(...$n); probe($n);}`, ""},
		{"builtin method value argument", `function f(DateTime $o){$n=5; work(); $o->format($n); probe($n);}`, "5"},
		{"unresolved static argument", `function f(){$n=5; work(); Unknown::fill($n); probe($n);}`, "5"},
		{"reference static argument", `class W { static function fill(&$x){} } function f(){$n=5; work(); W::fill($n); probe($n);}`, ""},
		{"builtin static value argument", `function f(){$n=5; work(); DateTime::createFromFormat($n, 'x'); probe($n);}`, "5"},
		{"unresolved constructor argument", `function f(){$n=5; work(); new Unknown($n); probe($n);}`, "5"},
		{"dynamic constructor argument", `function f($c){$n=5; work(); new $c($n); probe($n);}`, "5"},
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
		// Property reads have no flow value: only the written guard proves them.
		{"false", "property branch", `function f($o){if($o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property constant-first branch", `function f($o){if(false !== $o->h){probe($o->h);}}`, "true"},
		{"false", "property else arm", `function f($o){if($o->h === false){return 0;}else{probe($o->h);}}`, "true"},
		{"false", "property wrong arm", `function f($o){if($o->h !== false){return 0;}else{probe($o->h);}}`, "false"},
		{"false", "property elseif arm", `function f($o, $c){if($c){return 0;}elseif($o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property after elseif", `function f($o, $c){if($c){return 0;}elseif($o->h === false){return 1;}else{probe($o->h);}}`, "true"},
		{"false", "property after first test", `function f($o, $c){if($o->h === false){return 0;}elseif($c){probe($o->h);}}`, "true"},
		{"false", "property later elseif", `function f($o, $c, $d){if($c){return 0;}elseif($o->h === false){return 1;}elseif($d){probe($o->h);}}`, "true"},
		{"false", "property unrelated elseif", `function f($o, $c, $d){if($c){return 0;}elseif($d){probe($o->h);}elseif($o->h === false){return 1;}}`, "false"},
		{"false", "property spaced else if", `function f($o, $c){if($o->h === false){return 0;}else if($c){probe($o->h);}}`, "true"},
		{"false", "property conjunction operand", `function f($o){return $o->h !== false && probe($o->h);}`, "true"},
		{"false", "property word conjunction operand", `function f($o){return ($o->h !== false and probe($o->h));}`, "true"},
		{"false", "property disjunction operand", `function f($o){return $o->h === false || probe($o->h);}`, "true"},
		{"false", "property word disjunction operand", `function f($o){return ($o->h === false or probe($o->h));}`, "true"},
		{"false", "property wrong disjunction", `function f($o){return $o->h !== false || probe($o->h);}`, "false"},
		{"false", "property left operand", `function f($o){return probe($o->h) && $o->h !== false;}`, "false"},
		{"false", "property other operator", `function f($o){return ($o->h !== false) + probe($o->h);}`, "false"},
		{"false", "property ternary arm", `function f($o){return $o->h !== false ? probe($o->h) : 0;}`, "true"},
		{"false", "property ternary else arm", `function f($o){return $o->h === false ? 0 : probe($o->h);}`, "true"},
		{"false", "property wrong ternary arm", `function f($o){return $o->h !== false ? 0 : probe($o->h);}`, "false"},
		{"false", "property short ternary", `function f($o){return $o->h === false ?: probe($o->h);}`, "true"},
		{"false", "property ternary condition", `function f($o){return probe($o->h) ? $o->h !== false : 0;}`, "false"},
		{"false", "property negated guard", `function f($o){if(!($o->h === false)){probe($o->h);}}`, "true"},
		{"false", "property conjunction guard", `function f($o, $c){if($c && $o->h !== false){probe($o->h);}}`, "true"},
		{"false", "property false conjunction", `function f($o, $c){if($c && $o->h === false){return 0;} probe($o->h);}`, "false"},
		{"false", "property disjunction exit", `function f($o, $c){if($c || $o->h === false){return 0;} probe($o->h);}`, "true"},
		{"false", "property true disjunction", `function f($o, $c){if($c || $o->h !== false){probe($o->h);}}`, "false"},
		{"false", "property preceding exit", `function f($o){if($o->h === false){return 0;} probe($o->h);}`, "true"},
		{"false", "property other value", `function f($o){if($o->g === false){return 0;} probe($o->h);}`, "false"},
		{"false", "repeated call", `function f($o){if($o->h() === false){return 0;} probe($o->h());}`, "false"},
		{"false", "closure boundary", `function f($o){if($o->h === false){return 0;} $g=function() use ($o){probe($o->h);};}`, "false"},
		{"local-false", "property branch", `function f($o){if(\FALSE !== ($o->h)){probe($o->h);}}`, "false"},
		{"local-false", "callable branch", `function f($o){$g=$o->pick(); if(\FALSE !== ($g)){probe($g);}}`, "true"},
		{"length", "no input", `function f($s){probe();}`, "false"},
		{"length", "flow fact", `function f($s){if(strlen($s) !== 4){return;} echo 1; probe($s);}`, "true"},
		{"length", "property branch", `function f($o){if(strlen($o->s) === 4){probe($o->s);}}`, "true"},
		{"length", "property constant-first", `function f($o){if(4 === strlen($o->s)){probe($o->s);}}`, "true"},
		{"length", "property other length", `function f($o){if(strlen($o->s) === 5){probe($o->s);}}`, "false"},
		{"length", "property ternary", `function f($o){return strlen($o->s) === 4 ? probe($o->s) : 0;}`, "true"},
		{"length", "property other test", `function f($o){if(count($o->s) === 4){probe($o->s);}}`, "false"},
		{"length", "property loose test", `function f($o){if(strlen($o->s) == 4){probe($o->s);}}`, "false"},
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
