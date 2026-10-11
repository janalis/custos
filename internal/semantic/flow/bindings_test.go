package flow

import (
	"fmt"
	"strings"
	"testing"
)

// A value stays resolved only while no write the transfer functions cannot
// attribute may have replaced it.
func TestIndirectWritesDiscardResolvedValues(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		resolved  bool
	}{
		{"plain", `$x=1; probe($x);`, true},
		{"post increment", `$x=1; $x++; probe($x);`, false},
		{"pre decrement", `$x=1; --$x; probe($x);`, false},
		{"increment before copy", `$x=1; $x++; $y=$x; probe($y);`, false},
		{"copy before increment", `$x=1; $y=$x; $x++; probe($y);`, true},
		{"post increment yields old value", `$x=1; $y=$x++; probe($y);`, true},
		{"pre increment yields itself", `$x=1; $y=++$x; probe($y);`, true},
		{"element increment", `$x=[1]; $x[0]++; probe($x);`, false},
		{"other element increment", `$x=1; $y=[1]; $y[0]++; probe($x);`, true},
		{"property increment", `$x=1; $this->n++; probe($x);`, true},
		{"variable variable increment", `$x=1; $$n++; probe($x);`, false},
		{"loop counter", `$x=1; while($c){ probe($x); $x++; }`, false},
		{"reference capture then call", `$x=1; $f=function() use (&$x){$x=2;}; $f(); probe($x);`, false},
		{"reference capture without call", `$x=1; $f=function() use (&$x){$x=2;}; probe($x);`, true},
		{"reference capture then callback builtin", `$x=1; $f=function() use (&$x){$x=2;}; array_map($f, [1]); probe($x);`, false},
		{"reference capture then plain builtin", `$x=1; $f=function() use (&$x){$x=2;}; strlen('a'); probe($x);`, true},
		{"reference capture then construction", `$x=1; $f=function() use (&$x){$x=2;}; new Runner($f); probe($x);`, false},
		{"reference capture then dynamic construction", `$x=1; $f=function() use (&$x){$x=2;}; new $c($f); probe($x);`, false},
		{"reference capture then native construction", `$x=1; $f=function() use (&$x){$x=2;}; new ArrayObject([]); probe($x);`, true},
		{"invoked reference capture", `$x=1; (function() use (&$x){$x=2;})(); probe($x);`, false},
		{"invoked value capture", `$x=1; (function() use ($x){$x=2;})(); probe($x);`, true},
		{"later reference capture", `$x=1; probe($x); $f=function() use (&$x){$x=2;}; $f();`, true},
		{"value capture", `$x=1; $f=function() use ($x){}; $f(); probe($x);`, true},
		{"arrow function", `$x=1; $f=fn() => $x + 1; $f(); probe($x);`, true},
		{"reference item then call", `$x=1; $w=['k'=>&$x]; unknown($w); probe($x);`, false},
		{"reference item without call", `$x=1; $w=['k'=>&$x]; probe($x);`, true},
		{"reference alias written", `$x=1; $r=&$x; $x=1; $r=2; probe($x);`, false},
		{"reference alias rebound", `$r=&$x; $x=1; probe($x);`, true},
		{"reference alias stepped", `$r=&$x; $x=1; $r++; probe($x);`, false},
		{"reference element alias", `$x=[1]; $r=&$x[0]; $x=[1]; $r=2; probe($x);`, false},
		{"reference loop", `$x=[1]; foreach($x as &$q){$q=3;} probe($x);`, false},
		{"reference loop in target", `$x=[[1]]; foreach($x as [&$q]){$q=3;} probe($x);`, false},
		{"nested reference loop in target", `$x=[[[1]]]; foreach($x as [[&$q]]){$q=3;} probe($x);`, false},
		{"value loop", `$x=[1]; foreach($x as $q){$q=3;} probe($x);`, true},
		{"destructuring loop target", `$x=1; foreach($rows as [$a,$x]){} probe($x);`, false},
		{"list loop target", `$x=1; foreach($rows as list($a,$x)){} probe($x);`, false},
		{"reference list loop target", `$x=[[1]]; foreach($x as list(&$q)){$q=3;} probe($x);`, false},
		{"skipped destructuring slot", `$x=1; foreach($rows as [,$a]){} probe($x);`, true},
		{"element loop target", `$x=[1]; foreach($rows as $x['k']){} probe($x);`, false},
		{"variable variable loop target", `$x=1; foreach($rows as $$n){} probe($x);`, false},
		{"property loop target", `$x=1; foreach($rows as $this->p){} probe($x);`, true},
		{"loop key", `$x=1; foreach($rows as $x => $v){} probe($x);`, false},
		{"catch binding", `$x=1; try{a();}catch(E $x){} probe($x);`, false},
		{"catch without variable", `$x=1; try{strlen('a');}catch(E){} probe($x);`, true},
		{"other catch binding", `$x=1; try{strlen('a');}catch(E $e){} probe($x);`, true},
		{"extract", `$x=1; extract($arr); probe($x);`, false},
		{"parse_str into scope", `$x=1; parse_str($q); probe($x);`, false},
		{"parse_str into array", `$x=1; parse_str($q, $out); probe($x);`, true},
		{"global written by function", `function w(){global $x; $x=3;} $x=1; w(); probe($x);`, false},
		{"global not written before call", `function w(){global $x; $x=3;} $x=1; probe($x); w();`, true},
		{"global variable variable", `function w(){global $$n;} $x=1; w(); probe($x);`, true},
		{"other global", `function w(){global $y; $y=3;} $x=1; w(); probe($x);`, true},
		{"top-level global statement", `global $y; $x=1; w(); probe($x);`, true},
		{"globals table in function", `function w(){$GLOBALS['x']=3;} $x=1; w(); probe($x);`, false},
		{"globals table callback", `function w(){$GLOBALS['x']=3;} $x=1; array_map('w', [1]); probe($x);`, false},
		{"globals table at top level", `$all=$GLOBALS; $x=1; w(); probe($x);`, true},
		{"declared global then call", `function f(){global $x; $x=1; w(); probe($x);}`, false},
		{"declared global without call", `function f(){global $x; $x=1; probe($x);}`, true},
		{"reference parameter then call", `function f(&$x){$x=1; w(); probe($x);}`, false},
		{"reference parameter without call", `function f(&$x){$x=1; probe($x);}`, true},
		{"value parameter then call", `function f($x){$x=1; w(); probe($x);}`, true},
		{"two reference parameters", `function f(&$x, &$y){$x=1; $y=2; probe($x);}`, true},
		{"own reference capture then call", `$f=function() use (&$x){$x=1; w(); probe($x);};`, false},
		{"own value capture then call", `$f=function() use ($x){$x=1; w(); probe($x);};`, true},
		{"reference parameter aliased in a loop", `function f(&$x){$x=1; while($c){probe($x); $r=&$x; $x=1; $r=2;}}`, false},
		{"reference parameters aliased in a loop", `function f(&$x, &$r){$x=1; while($c){probe($x); $r=&$x; $x=1; $r=2;}}`, false},
		{"reference to a variable variable", `$x=1; $y=2; $r=&$$n; $x=1; $r=2; probe($x);`, true},
		{"declared global variable variable", `function f(){$x=1; global $$n; $x=1; probe($x);}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			if _, known := e.Resolve(argument(callsNamed(f, "probe")[0])); known != tc.resolved {
				t.Fatalf("resolved=%v", known)
			}
		})
	}
}

func TestTaintFollowsValuesNotSelectors(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"match constant arms", `$y=match($_GET['a']){'a'=>'b', default=>'c'}; probe($y);`, false},
		{"match request arm", `$y=match($k){'a'=>$_GET['a'], default=>'c'}; probe($y);`, true},
		{"match request condition", `$y=match($k){$_GET['a']=>'b', default=>'c'}; probe($y);`, false},
		{"match without arms", `$y=match($_GET['a']){}; probe($y);`, false},
		{"match arm assignment", `$x=$_GET['a']; $y=match($k){1=>$x='safe', default=>1}; probe($x);`, true},
		{"subtraction", `$y=$_GET['a'] - 1; probe($y);`, false},
		{"product", `$y=$_GET['a'] * $_GET['b']; probe($y);`, false},
		{"quotient", `$y=$_GET['a'] / 2; probe($y);`, false},
		{"modulo", `$y=$_GET['a'] % 2; probe($y);`, false},
		{"power", `$y=$_GET['a'] ** 2; probe($y);`, false},
		{"shift", `$y=$_GET['a'] << 1 >> 1; probe($y);`, false},
		{"addition may unite arrays", `$y=$_GET['a'] + $_GET['b']; probe($y);`, true},
		{"byte-wise or", `$y=$_GET['a'] | 'x'; probe($y);`, true},
		{"concatenation", `$y=$_GET['a'] . 'x'; probe($y);`, true},
		{"negation", `$y=-$_GET['a']; probe($y);`, false},
		{"unary plus", `$y=+$_GET['a']; probe($y);`, false},
		{"byte-wise not", `$y=~$_GET['a']; probe($y);`, true},
		{"length", `$y=strlen($_GET['a']); probe($y);`, false},
		{"count", `$y=count($_GET['a']); probe($y);`, false},
		{"position or false", `$y=strpos($_GET['a'], 'x'); probe($y);`, false},
		{"string result", `$y=trim($_GET['a']); probe($y);`, true},
		{"mixed result", `$y=max($_GET['a'], 1); probe($y);`, true},
		{"scalar method result", `$y=(new ArrayObject($_GET['a']))->count(); probe($y);`, false},
		{"string method result", `$d=new DateTime(); $y=$d->format($_GET['a']); probe($y);`, true},
		{"scalar static result", `$y=IntlChar::ord($_GET['a']); probe($y);`, false},
		{"string static result", `$y=Normalizer::normalize($_GET['a']); probe($y);`, true},
		{"short ternary keeps its condition", `$y=$_GET['a'] ?: 'd'; probe($y);`, true},
		{"short ternary of clean condition", `$y=$k ?: 'd'; probe($y);`, false},
		{"long ternary ignores its condition", `$y=$_GET['a'] ? 'x' : 'd'; probe($y);`, false},
		{"stepped request value", `$y=$_GET['a']; $y++; probe($y);`, true},
		{"extract of request", `$y=1; extract($_GET); probe($y);`, true},
		{"extract of clean array", `$y=1; extract(['y'=>2]); probe($y);`, false},
		{"ternary arm pinned by its condition", `$x=$_GET['a']; $y=$x === 'safe' ? $x : 'd'; probe($y);`, false},
		{"conjunction operand pinned by the left", `$x=$_GET['a']; $x === 'safe' && probe($x);`, false},
		{"disjunction operand pinned by the left", `$x=$_GET['a']; $x !== 'safe' || probe($x);`, false},
		{"disjunction operand of an equality", `$x=$_GET['a']; $x === 'safe' || probe($x);`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			x := argument(callsNamed(f, "probe")[0])
			if got := e.Tainted(x, HTML); got != tc.taint {
				t.Fatalf("taint=%v value=%+v", got, e.Value(x))
			}
		})
	}
}

func TestScalarResult(t *testing.T) {
	for declared, want := range map[string]bool{
		"": false, "int": true, "INT|false": true, "?float": true, "bool|null": true, "void": true, "never": true,
		"string": false, "int|string": false, "array|false": false, "mixed": false, "true": true,
	} {
		if got := scalarResult(declared); got != want {
			t.Errorf("scalarResult(%q)=%v", declared, got)
		}
	}
}

func TestGuardsExcludeFailureSentinels(t *testing.T) {
	for _, tc := range []struct {
		name, src         string
		nonFalse, nonNull bool
	}{
		{"unchecked", `$x=fopen('a','r'); probe($x);`, false, false},
		{"strict false exit", `$x=fopen('a','r'); if($x === false){return;} probe($x);`, true, false},
		{"constant-first false exit", `$x=fopen('a','r'); if(false === $x){return;} probe($x);`, true, false},
		{"qualified constant", `$x=fopen('a','r'); if($x === \FALSE){return;} probe($x);`, true, false},
		{"strict null exit", `$x=preg_replace('/a/','b','c'); if(null === $x){return;} probe($x);`, false, true},
		{"strict true branch", `$x=fopen('a','r'); if($x === true){probe($x);}`, false, false},
		{"strict other constant", `$x=fopen('a','r'); if($x === PHP_EOL){return;} probe($x);`, false, false},
		{"truthy branch", `$x=fopen('a','r'); if($x){probe($x);}`, true, true},
		{"falsy branch", `$x=fopen('a','r'); if($x){}else{probe($x);}`, false, false},
		{"falsy exit", `$x=fopen('a','r'); if(!$x){return;} probe($x);`, true, true},
		{"falsy exit with else", `$x=fopen('a','r'); if(!$x){return;}else{echo 1;} probe($x);`, true, true},
		{"falsy branch continues", `$x=fopen('a','r'); if(!$x){echo 1;} probe($x);`, false, false},
		{"elseif arm", `$x=fopen('a','r'); if($c){}elseif($x){probe($x);}`, true, true},
		{"after elseif exit", `$x=fopen('a','r'); if($c){return;}elseif(!$x){return;} probe($x);`, true, true},
		{"after elseif branch", `$x=fopen('a','r'); if($c){echo 1;}elseif(!$x){return;} probe($x);`, false, false},
		{"both arms exit", `$x=fopen('a','r'); if($c){return;}elseif(!$x){return;}else{probe($x);}`, true, true},
		{"assignment in condition", `if($x=fopen('a','r')){probe($x);}`, true, true},
		{"compared assignment", `if(($x=fopen('a','r')) === false){return;} probe($x);`, true, false},
		{"compound assignment", `$x=1; if($x .= 'a'){probe($x);}`, false, false},
		{"variable variable", `if($$x){probe($$x);}`, false, false},
		{"untracked variable", `if($u){probe($u);}`, false, false},
		{"empty exit", `$x=fopen('a','r'); if(empty($x)){return;} probe($x);`, true, true},
		{"empty branch", `$x=fopen('a','r'); if(empty($x)){probe($x);}`, false, false},
		{"isset branch", `$x=preg_replace('/a/','b','c'); if(isset($x)){probe($x);}`, false, true},
		{"isset exit", `$x=preg_replace('/a/','b','c'); if(isset($x)){return;} probe($x);`, false, false},
		{"instanceof branch", `$x=date_create('now'); if($x instanceof DateTime){probe($x);}`, true, true},
		{"instanceof exit", `$x=date_create('now'); if($x instanceof DateTime){return;} probe($x);`, false, false},
		{"loose false exit", `$x=fopen('a','r'); if($x == false){return;} probe($x);`, true, true},
		{"loose null branch", `$x=fopen('a','r'); if(null != $x){probe($x);}`, true, true},
		{"loose other constant", `$x=fopen('a','r'); if($x == true){return;} probe($x);`, false, false},
		{"type predicate", `$x=file_get_contents('a'); if(is_string($x)){probe($x);}`, true, true},
		{"negated type predicate", `$x=file_get_contents('a'); if(!is_string($x)){return;} probe($x);`, true, true},
		{"type predicate exit", `$x=file_get_contents('a'); if(is_string($x)){return;} probe($x);`, false, false},
		{"not boolean", `$x=file_get_contents('a'); if(is_bool($x)){return;} probe($x);`, true, false},
		{"boolean", `$x=file_get_contents('a'); if(is_bool($x)){probe($x);}`, false, false},
		{"not null predicate", `$x=preg_replace('/a/','b','c'); if(is_null($x)){return;} probe($x);`, false, true},
		{"null predicate", `$x=preg_replace('/a/','b','c'); if(is_null($x)){probe($x);}`, false, false},
		{"conjunction operand", `$x=fopen('a','r'); $x !== false && probe($x);`, true, false},
		{"word conjunction operand", `$x=fopen('a','r'); $x and probe($x);`, true, true},
		{"disjunction operand", `$x=fopen('a','r'); $x === false || probe($x);`, true, false},
		{"word disjunction operand", `$x=fopen('a','r'); !$x or probe($x);`, true, true},
		{"wrong disjunction operand", `$x=fopen('a','r'); $x !== false || probe($x);`, false, false},
		{"after conjunction", `$x=fopen('a','r'); $x !== false && work(); probe($x);`, false, false},
		{"coalesce operand", `$x=fopen('a','r'); $y = $x ?? probe($x);`, false, false},
		{"ternary arm", `$x=fopen('a','r'); $y = $x !== false ? probe($x) : 0;`, true, false},
		{"other ternary arm", `$x=fopen('a','r'); $y = $x === false ? 0 : probe($x);`, true, false},
		{"wrong ternary arm", `$x=fopen('a','r'); $y = $x !== false ? 0 : probe($x);`, false, false},
		{"short ternary arm", `$x=fopen('a','r'); $y = $x === false ?: probe($x);`, true, false},
		{"after ternary", `$x=fopen('a','r'); $y = $x ? 1 : 0; probe($x);`, false, false},
		{"loop condition", `function f($h){while(($x=fgets($h)) !== false){probe($x);}}`, true, false},
		{"reassigned after guard", `$x=fopen('a','r'); if(!$x){return;} $x=fopen('b','r'); probe($x);`, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			x := argument(callsNamed(f, "probe")[0])
			if got := e.Excludes(x, "false"); got != tc.nonFalse {
				t.Errorf("non-false=%v value=%+v", got, e.Value(x))
			}
			if got := e.Excludes(x, "null"); got != tc.nonNull {
				t.Errorf("non-null=%v value=%+v", got, e.Value(x))
			}
		})
	}
}

// A literal pinned by a constant-first comparison keeps the tested value's
// origin on the paths that rejoin, exactly like an unguarded read.
func TestMirroredComparisonsKeepOrigin(t *testing.T) {
	for _, tc := range []struct {
		src      string
		resolved bool
	}{
		{`$x=f(); if(1 === $x){} probe($x);`, true},
		{`$x=f(); if($x === 1){} probe($x);`, false},
		{`$x=$_GET['x']; if(4 !== strlen($x)){return;} probe($x);`, true},
	} {
		e, f := environment(t, tc.src, nil)
		if _, known := e.Resolve(argument(callsNamed(f, "probe")[0])); known != tc.resolved {
			t.Errorf("%s resolved=%v", tc.src, known)
		}
	}
	e, f := environment(t, `$x=$_GET['x']; if(4 !== strlen($x)){return;} probe($x);`, nil)
	if n, ok := e.ExactLength(argument(callsNamed(f, "probe")[0])); !ok || n != 4 {
		t.Fatalf("length=%d/%v", n, ok)
	}
	e, f = environment(t, `$x=$_GET['x']; if(0 !== preg_match('/[\r\n]/', $x)){return;} probe($x);`, nil)
	if e.Tainted(argument(callsNamed(f, "probe")[0]), Header) {
		t.Fatal("constant-first line break guard")
	}
}

// A success test inside an expression must not erase the operation history
// of the tested handle on the paths that skip it.
func TestStatesSurviveDisagreeingSuccess(t *testing.T) {
	for _, tc := range []struct {
		src               string
		known, successful bool
	}{
		{`$h=fopen('a','r'); if(!$h || $h === false){return;} fread($h,1);`, true, true},
		{`$h=fopen('a','r'); if($h){echo 1;} fread($h,1);`, true, false},
		{`$h=fopen('a','r'); $h && work(); fread($h,1);`, true, false},
		{`$h=fopen('a','r'); $h && work($h); fread($h,1);`, false, false},
		{`$h=fopen('a','r'); $ok = $h ? 1 : 0; fread($h,1);`, true, false},
		{`$h=fopen('a','r'); while($c){ if($h){ echo 1; } } fread($h,1);`, true, false},
		{`$h=fopen('a','r'); while($c){ fread($h,1); fclose($h); }`, false, false},
	} {
		e, f := environment(t, tc.src, nil)
		c := callsNamed(f, "fread")[0]
		s, ok := e.StateBefore(c, argument(c), "fopen")
		if ok != tc.known || s.Successful != tc.successful {
			t.Errorf("%s state=%+v/%v", tc.src, s, ok)
		}
	}
}

func TestCurlOptionSinkAcceptsQualifiedConstant(t *testing.T) {
	e, f := environment(t, `function fetch($u){$c=curl_init(); curl_setopt($c, \CURLOPT_URL, $u); return curl_exec($c);} function plain($u){$c=curl_init(); curl_setopt($c, \CURLOPT_TIMEOUT, $u);}`, nil)
	file := Extract(f, e.types.Index, e.types.PHP, nil)
	options := func(name string) int {
		n := 0
		for _, sink := range file.Summaries[name].Sinks {
			if sink.Name == "curl_setopt" && sink.Context == URL {
				n++
			}
		}
		return n
	}
	if options("fetch") != 1 || options("plain") != 0 {
		t.Fatalf("summaries=%+v", file.Summaries)
	}
}

func BenchmarkIndirectWrites(b *testing.B) {
	var src strings.Builder
	src.WriteString("function w(){global $v0; $v0=1;} function f(&$out, $rows){")
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&src, "$v%d=$_GET['x']; $v%d++; $g%d=function() use (&$v%d){}; foreach($rows as [$a%d, &$b%d]){} if(!$v%d){return;} $t%d = $v%d !== false ? strlen($v%d) : match($v%d){1=>2, default=>3}; w();", i, i, i, i, i, i, i, i, i, i, i)
	}
	src.WriteString("probe($v39);}")
	e, f := environment(b, src.String(), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env := New(f, e.types, nil)
		env.Value(x)
	}
}
