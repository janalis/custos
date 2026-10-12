package flow

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// A value stays resolved only while no write the transfer functions cannot
// attribute may have replaced it.
func TestIndirectWritesDiscardResolvedValues(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		resolved  bool
	}{
		{"plain", `$x=1; probe($x);`, true},
		// A step of a known value resolves to the step itself, which a reader
		// evaluates; a step of an unknown value resolves to nothing.
		{"post increment", `$x=1; $x++; probe($x);`, true},
		{"pre decrement", `$x=1; --$x; probe($x);`, true},
		{"increment before copy", `$x=1; $x++; $y=$x; probe($y);`, true},
		{"increment of a parameter", `function f($x){$x++; probe($x);}`, false},
		{"increment in a do-while", `$x=1; do{$x++;}while($c); probe($x);`, false},
		{"assignment before a do-while", `$x=1; do{$y=2;}while($c); probe($x);`, true},
		{"static local across a call into the project", `function f(){static $x; $x=1; work(); probe($x);}`, false},
		{"static local across a builtin call", `function f(){static $x; $x=1; strlen('a'); probe($x);}`, true},
		{"increment in one branch", `$x=1; if($c){$x++;} probe($x);`, false},
		{"compound assignment", `$x=1; $x+=2; probe($x);`, true},
		{"compound assignment of a parameter", `function f($x){$x.='a'; probe($x);}`, false},
		{"coalescing assignment", `$x=1; $x??=2; probe($x);`, false},
		{"compound assignment of another variable", `$x=1; $y='a'; $y.='b'; probe($x);`, true},
		{"compound assignment of an element", `$x=1; $y=[]; $y['k'].='b'; probe($x);`, true},
		{"element assignment of the array itself", `$x=[1]; $x['k']='b'; probe($x);`, false},
		{"property assignment", `$x=1; $this->p=2; probe($x);`, true},
		{"dynamic globals entry written in a function", `function f($k){$x=1; $GLOBALS[$k]=$_GET['b']; probe($x);}`, true},
		{"destructuring of other variables", `$x=1; [$a, $b]=$r; probe($x);`, true},
		{"destructuring", `$x=1; [$a, $x]=$r; probe($x);`, false},
		{"variable variable assignment", `$x=1; $$n=2; probe($x);`, false},
		// A reference stored where the analysis does not follow it.
		{"reference kept in a property, property written", `$x=1; $o->p=&$x; $o->p=5; probe($x);`, false},
		{"reference kept in a property, other name written", `$x=1; $o->p=&$x; $y=2; probe($x);`, true},
		{"reference kept in a static property", `$x=1; Reg::$p=&$x; Reg::$p=5; probe($x);`, false},
		{"reference kept in a property element", `$x=1; $o->a[0]=&$x; $o->a[0]=5; probe($x);`, false},
		{"reference kept in a property, compound write", `$x=1; $o->p=&$x; $o->p+=6; probe($x);`, false},
		{"reference kept in a property, step", `$x=1; $o->p=&$x; $o->p--; probe($x);`, false},
		{"reference kept in a property, unset", `$x=1; $o->p=&$x; unset($o->p); probe($x);`, false},
		{"reference kept in an array slot", `$x=1; $a[0]=&$x; $a[0]=5; probe($x);`, false},
		{"reference kept in a variable variable", `$x=1; $$n=&$x; $o->p=5; probe($x);`, false},
		{"property written without a kept reference", `$x=1; $r=&$y; $o->p=5; probe($x);`, true},
		{"definite alias then kept reference", `$x=1; $r=&$x; $o->p=&$r; $o->p=5; probe($x);`, false},
		{"two definite aliases", `$x=1; $r=&$x; $q=&$r; $q=2; probe($x);`, true},
		{"alias rebound to another variable", `$x=1; $r=&$x; $r=&$y; $r=2; probe($x);`, true},
		{"alias of an element", `$x=[1]; $r=&$x[0]; $r=2; probe($x);`, false},
		{"alias of a call result", `$x=1; $r=&f(); $r=2; probe($x);`, true},
		{"aliases joined from two branches", `$x=1; if($c){$r=&$x;}else{$r=&$y;} $r=2; probe($x);`, false},
		{"alias bound in one branch only", `$x=1; if($c){$r=&$x;} $r=2; probe($x);`, false},
		// The globals table is the file-level scope.
		{"globals entry assigned at file level", `$x=1; $GLOBALS['x']=5; probe($x);`, true},
		{"globals entry appended at file level", `$x=1; $GLOBALS['x'].='a'; probe($x);`, false},
		{"globals entry element at file level", `$x=[1]; $GLOBALS['x']['k']=5; probe($x);`, false},
		{"other globals entry at file level", `$x=1; $GLOBALS['y']=5; probe($x);`, true},
		{"dynamic globals entry at file level", `$x=1; $GLOBALS[$k]=5; probe($x);`, false},
		{"globals entry stepped at file level", `$x=1; $GLOBALS['x']++; probe($x);`, false},
		{"globals entry of a bound name", `function f(){global $x; $x=1; $GLOBALS['x']=5; probe($x);}`, true},
		{"globals entry of an unbound name", `function f(){$x=1; $GLOBALS['x']=5; probe($x);}`, true},
		{"dynamic globals entry, bound name", `function f($k){global $x; $x=1; $GLOBALS[$k]=5; probe($x);}`, false},
		{"dynamic globals entry, unbound name", `function f($k){global $y; $x=1; $GLOBALS[$k]=5; probe($x);}`, true},
		{"other variable unset", `$x=1; unset($y); probe($x);`, true},
		{"other variable declared static", `function f(){$x=1; static $y; probe($x);}`, true},
		{"other variable declared global", `function f(){$x=1; global $y; probe($x);}`, true},
		{"compound assignment in a loop", `$x=1; while($c){probe($x); $x+=1;}`, false},
		{"compound assignment through an alias", `$r=&$x; $x=1; $r+=1; probe($x);`, true},
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
		{"reference item written through its array", `$x=1; $w=[&$x]; $w[0]=2; probe($x);`, false},
		{"element write beside a reference alias", `$r=&$x; $x=1; $w=[]; $w[0]=2; probe($x);`, true},
		{"reference item without call", `$x=1; $w=['k'=>&$x]; probe($x);`, true},
		{"reference alias written", `$x=1; $r=&$x; $x=1; $r=2; probe($x);`, true},
		{"reference alias rebound", `$r=&$x; $x=1; probe($x);`, true},
		{"reference alias stepped", `$r=&$x; $x=1; $r++; probe($x);`, true},
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
		{"appended in a loop", `$n=$_GET['a']; foreach($x as $k){$n.=$k;} probe($n);`, true},
		{"rebuilt in a loop", `$n=$_GET['a']; foreach($x as $k){$n=$n . $k;} probe($n);`, true},
		{"request data appended in a branch", `$n='ls'; if($c){$n.=$_GET['a'];} probe($n);`, true},
		{"request data appended in a loop", `$n='ls'; foreach($_GET['a'] as $k){$n.=$k;} probe($n);`, true},
		{"request data keys appended in a loop", `$n='ls'; foreach($_GET as $k=>$v){$n.=$k;} probe($n);`, true},
		{"escaped elements appended in a loop", `$n='ls'; foreach($_GET['a'] as $k){$n.=htmlspecialchars($k);} probe($n);`, false},
		{"elements of a clean list appended in a loop", `$n='ls'; foreach(['a','b'] as $k){$n.=$k;} probe($n);`, false},
		{"clean text appended", `$n='ls'; $n.=' -l'; probe($n);`, false},
		{"unknown text appended to clean text", `$n='ls'; $n.=$this->flag; probe($n);`, false},
		{"unknown text appended to escaped data", `$n=htmlspecialchars($_GET['a']); $n.=$this->flag; probe($n);`, false},
		{"request data beside unknown text", `$y=$this->table . $_GET['a']; probe($y);`, true},
		{"request data interpolated beside unknown text", `$y="{$this->table} {$_GET['a']}"; probe($y);`, true},
		{"unknown call result beside unknown text", `$y=$this->table . unknown($_GET['a']); probe($y);`, false},
		{"number subtracted in place", `$n=$_GET['a']; $n-=1; probe($n);`, false},
		{"number added in place may unite arrays", `$n=$_GET['a']; $n+=1; probe($n);`, true},
		{"request data coalesced in place", `$n=null; $n??=$_GET['a']; probe($n);`, true},
		{"superglobal appended in place", `$_GET['a'].='x'; $n=1; probe($n);`, false},
		{"passed to an unresolved function", `$n=$_GET['a']; other($n); probe($n);`, true},
		{"passed to an unresolved method", `$n=$_GET['a']; $o->other($n); probe($n);`, true},
		{"clean value passed to an unresolved function", `$n=1; other($n); probe($n);`, false},
		{"passed to a declared reference parameter", `function clean(&$v){$v='safe';} $n=$_GET['a']; clean($n); probe($n);`, false},
		{"passed to a builtin output parameter", `$n=$_GET['a']; preg_match('/a/', 'a', $n); probe($n);`, false},
		{"request data imported into the scope", `extract($_GET); probe($cmd);`, true},
		{"query string imported into the scope", `parse_str($_SERVER['QUERY_STRING']); probe($cmd);`, true},
		{"query string parsed into an array", `parse_str($_SERVER['QUERY_STRING'], $q); probe($cmd);`, false},
		{"clean array imported into the scope", `extract(['cmd'=>'ls']); probe($cmd);`, false},
		{"unknown array imported into the scope", `$x=$_GET['a']; extract(foo()); probe($x);`, true},
		{"unknown array imported over a clean value", `$x='ls'; extract(foo()); probe($x);`, false},
		{"request data imported in one branch", `if($c){extract($_GET);} probe($cmd);`, true},
		{"request data imported in both branches", `if($c){extract($_GET);}else{extract($_POST);} probe($cmd);`, true},
		{"request data imported, variable reassigned", `extract($_GET); $cmd='ls'; probe($cmd);`, false},
		{"element stepped", `$x=$_GET; $x['n']++; probe($x['c']);`, true},
		{"element used as a loop target", `$x=$_GET; foreach([1] as $x['k']){} probe($x['c']);`, true},
		{"property element used as a loop target", `$x=$_GET['a']; foreach([1] as $this->p['k']){} probe($x);`, true},
		{"variable variable stepped", `$x=$_GET['a']; $$n++; probe($x);`, true},
		{"request data stored in an element", `$x=[]; $x['k']=$_GET['a']; probe($x['k']);`, true},
		{"request data destructured", `[$a, [$b]]=$_GET['a']; probe($b);`, true},
		{"clean data destructured", `$b=$_GET['a']; [$a, $b]=[1, 2]; probe($b);`, false},
		{"property written beside request data", `$x=$_GET['a']; $this->p=1; probe($x);`, true},
		{"reference alias written with request data", `$x='ls'; $r=&$x; $r=$_GET['a']; probe($x);`, true},
		{"unknown call result", `$y=conf('bin') . ' ' . $_GET['a']; probe($y);`, true},
		{"unknown call on request data", `$y=valid($_GET['a']); probe($y);`, false},
		{"unpacked request data argument", `$y=sprintf(...$_GET['a']); probe($y);`, false},
		// Elements with a literal key are read from their own last write.
		{"clean element beside a request element", `$a=[]; $a['k']=$_GET['x']; $a['c']='ls'; probe($a['c']);`, false},
		{"request element", `$a=[]; $a['k']=$_GET['x']; $a['c']='ls'; probe($a['k']);`, true},
		{"unknown element", `$a=[]; $a['k']=$_GET['x']; $a['c']='ls'; probe($a[$i]);`, true},
		{"whole array", `$a=[]; $a['k']=$_GET['x']; $a['c']='ls'; probe($a);`, true},
		{"untracked element", `$a=[]; $a['k']=$_GET['x']; probe($a['z']);`, true},
		{"element overwritten with clean data", `$a=[]; $a['k']=$_GET['x']; $a['k']='ls'; probe($a['k']);`, false},
		{"element after an unknown key write", `$a=[]; $a['c']='ls'; $a[$i]=$_GET['x']; probe($a['c']);`, true},
		{"element after an appended write", `$a=[]; $a['c']='ls'; $a[]=$_GET['x']; probe($a['c']);`, true},
		{"element after the array is replaced", `$a=[]; $a['c']='ls'; $a=$_GET; probe($a['c']);`, true},
		{"element after the array is passed on", `$a=[]; $a['c']='ls'; fill($a); $a['k']=$_GET['x']; probe($a['c']);`, true},
		{"element appended in place", `$a=[]; $a['c']='ls '; $a['c'].=$_GET['x']; probe($a['c']);`, true},
		{"other element appended in place", `$a=[]; $a['c']='ls'; $a['k']='x'; $a['k'].=$_GET['x']; probe($a['c']);`, false},
		{"nested element", `$a=[]; $a['c']='ls'; $a['k']['j']=$_GET['x']; probe($a['c']);`, false},
		{"nested request element", `$a=[]; $a['k']=[]; $a['k']['j']=$_GET['x']; probe($a['k']['j']);`, true},
		{"element written on one path", `$a=[]; $a['c']=$_GET['x']; if($f){$a['c']='ls';} probe($a['c']);`, true},
		{"element written on both paths", `$a=[]; $a['k']=$_GET['x']; if($f){$a['c']='ls';}else{$a['c']='pwd';} probe($a['c']);`, false},
		{"element tracked on one path only", `$a=[]; $a['k']=$_GET['x']; if($f){$a['c']='ls';} probe($a['c']);`, true},
		{"element stepped", `$a=[]; $a['k']=$_GET['x']; $a['c']='ls'; $a['c']++; probe($a['c']);`, true},
		{"literal array element", `$a=['k'=>$_GET['x'], 'c'=>'ls']; probe($a['c']);`, false},
		{"literal array request element", `$a=['k'=>$_GET['x'], 'c'=>'ls']; probe($a['k']);`, true},
		{"literal list element", `$a=['ls', $_GET['x']]; probe($a[0]);`, false},
		{"literal list request element", `$a=['ls', $_GET['x']]; probe($a[1]);`, true},
		{"literal array with mixed keys", `$a=['k'=>'ls', $_GET['x']]; probe($a['k']);`, false},
		{"literal array numbered after a key", `$a=[4=>'ls', $_GET['x'], 'ls']; probe($a[5]);`, true},
		{"literal array numbered past a request element", `$a=[4=>'ls', $_GET['x'], 'ls']; probe($a[6]);`, false},
		{"literal array numbered after a negative key", `$a=[-5=>'ls', $_GET['x']]; probe($a[-5]);`, true},
		{"literal array with a dynamic key", `$a=[$i=>$_GET['x'], 'c'=>'ls']; probe($a['c']);`, true},
		{"literal array with a spread", `$a=[...$_GET, 'c'=>'ls']; probe($a['c']);`, true},
		{"long literal array", `$a=[` + strings.Repeat(`'x', `, MaxSources) + `$_GET['x']]; probe($a[0]);`, true},
		{"empty literal array", `$a=[]; probe($a['c']);`, false},
		{"superglobal element written", `$_GET['a']='ls'; probe($_GET['a']);`, true},
		{"element after a holder call", `$a=[]; $a['c']='ls'; $f=function() use (&$a){$a['c']=$_GET['x'];}; $f(); probe($a['c']);`, true},
		{"element after an import", `$a=[]; $a['c']='ls'; extract($_GET); probe($a['c']);`, true},
		{"element after a globals table writer", `function w(){$GLOBALS[$k]=$_GET['x'];} $a=[]; $a['c']='ls'; w(); probe($a['c']);`, false},
		{"element of an aliased array", `$a=[]; $r=&$a; $a['c']='ls'; $a['k']=$_GET['x']; probe($a['c']);`, false},
		// Adding in place to a value that is not an array is arithmetic.
		{"number added to a literal", `$t=0; $t+=$_GET['q']; probe($t);`, false},
		{"number added twice", `$t=0; $t+=$_GET['q']; $t+=$_GET['r']; probe($t);`, false},
		{"number added to a string", `$t="a$b"; $t+=$_GET['q']; probe($t);`, false},
		{"number added after a step", `$t=0; $t++; $t+=$_GET['q']; probe($t);`, false},
		{"number added after a product", `$t=2; $t*=3; $t+=$_GET['q']; probe($t);`, false},
		{"array united with an array literal", `$t=[]; $t+=$_GET['q']; probe($t);`, true},
		{"array united after a union", `$t=[]; $t+=$u; $t+=$_GET['q']; probe($t);`, true},
		{"number added to an unknown value", `$t=$u; $t+=$_GET['q']; probe($t);`, true},
		{"number added after a long chain", `$t=0; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=1; $t+=$_GET['q']; probe($t);`, true},
		// What an import cannot create or replace.
		{"session after an import", `extract($_GET); probe($_SESSION['user']);`, false},
		{"environment after an import", `extract($_GET); probe($_ENV['BIN']);`, false},
		{"globals table after an import", `extract($_GET); probe($GLOBALS['bin']);`, false},
		{"object after an import", `extract($_GET); probe($this);`, false},
		{"import skipping existing variables", `$cmd='ls'; extract($_GET, EXTR_SKIP); probe($cmd);`, false},
		{"import prefixing every variable", `$cmd='ls'; extract($_GET, \EXTR_PREFIX_ALL, 'p'); probe($cmd);`, false},
		{"import skipping, new variable", `extract($_GET, EXTR_SKIP); probe($cmd);`, true},
		{"import overwriting", `$cmd='ls'; extract($_GET, EXTR_OVERWRITE); probe($cmd);`, true},
		{"import with dynamic flags", `$cmd='ls'; extract($_GET, $flags); probe($cmd);`, true},
		// Definite aliases share one value.
		{"clean write through an alias", `$x=$_GET['a']; $r=&$x; $r='ls'; probe($x);`, false},
		{"clean write through the aliased name", `$x=$_GET['a']; $r=&$x; $x='ls'; probe($r);`, false},
		{"escaped through an alias", `$x=$_GET['a']; $y=&$x; $y=htmlspecialchars($y); probe($x);`, false},
		{"request data through an alias", `$x='ls'; $r=&$x; $r=$_GET['a']; probe($x);`, true},
		{"alias reads the shared value", `$x=$_GET['a']; $r=&$x; probe($r);`, true},
		// Loops over request data.
		{"upload size in a loop", `foreach($_FILES as $f){probe($f['size']);}`, false},
		{"upload name in a loop", `foreach($_FILES as $f){probe($f['name']);}`, true},
		{"upload entry in a loop", `foreach($_FILES as $f){probe($f);}`, true},
		{"size key of other request data", `foreach($_GET as $f){probe($f['size']);}`, true},
		{"mapped to integers", `foreach(array_map('intval', $_GET['ids']) as $id){probe($id);}`, false},
		{"mapped to integers, joined", `$y=implode(',', array_map('\intval', $_GET['ids'])); probe($y);`, false},
		{"mapped to strings", `foreach(array_map('trim', $_GET['ids']) as $id){probe($id);}`, true},
		{"mapped by a closure", `foreach(array_map(fn($v) => $v, $_GET['ids']) as $id){probe($id);}`, true},
		{"mapped by an unknown function", `foreach(array_map('cleanup', $_GET['ids']) as $id){probe($id);}`, true},
		// PHP functions working on a reference argument.
		{"sorted in place", `$a=$_GET['args']; sort($a); probe($a);`, true},
		{"shifted in place", `$a=$_GET['args']; $first=array_shift($a); probe($a);`, true},
		{"clean data sorted in place", `$a=[2, 1]; sort($a); probe($a);`, false},
		{"converted to an integer in place", `$a=$_GET['id']; settype($a, 'integer'); probe($a);`, false},
		{"converted to a string in place", `$a=$_GET['id']; settype($a, 'string'); probe($a);`, true},
		{"converted to a dynamic type in place", `$a=$_GET['id']; settype($a, $t); probe($a);`, true},
		{"converted with a spread", `$a=$_GET['id']; settype($a, ...$t); probe($a);`, true},
		{"converted without a type", `$a=$_GET['id']; settype($a); probe($a);`, true},
		{"match groups of request data", `preg_match('/(.*)/', $_GET['a'], $m); probe($m[1]);`, true},
		{"match groups of clean data", `$m=$_GET['a']; preg_match('/(.*)/', 'abc', $m); probe($m);`, false},
		{"query string parsed into an array, element", `parse_str($_SERVER['QUERY_STRING'], $q); probe($q['c']);`, true},
		{"scanned values", `sscanf($_GET['a'], '%s %s', $x, $y); probe($y);`, true},
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

// An unknown value is not an untainted one: request data a binding held, or
// that a reference holder writes to it, survives the invalidation.
func TestInvalidationKeepsRequestData(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"tainted, holder writes a constant", `$n=$_GET['a']; $f=function() use (&$n){$n=1;}; $f(); probe($n);`, true},
		{"sum, holder writes request data", `$n=$_GET['a'] + 1; $f=function() use (&$n){$n=$_GET['b'];}; $f(); probe($n);`, true},
		{"length, holder writes request data", `$n=strlen($_GET['a']); $f=function() use (&$n){$n=$_GET['b'];}; $f(); probe($n);`, true},
		{"length, holder writes a constant", `$n=strlen($_GET['a']); $f=function() use (&$n){$n=1;}; $f(); probe($n);`, false},
		{"escaped, holder writes a constant", `$n=htmlspecialchars($_GET['a']); $f=function() use (&$n){$n=1;}; $f(); probe($n);`, false},
		{"constant, holder writes request data", `$n=1; $f=function() use (&$n){$n=$_GET['b'];}; $f(); probe($n);`, true},
		{"constant, holder writes escaped data", `$n=1; $f=function() use (&$n){$n=htmlspecialchars($_GET['b']);}; $f(); probe($n);`, false},
		{"constant, holder writes a constant", `$n=1; $f=function() use (&$n){$n=2;}; $f(); probe($n);`, false},
		{"constant, nested scope writes request data", `$n=1; $f=function() use (&$n){$g=function(){$n=$_GET['b'];};}; $f(); probe($n);`, false},
		{"constant, holder never called", `$n=1; $f=function() use (&$n){$n=$_GET['b'];}; probe($n);`, false},
		{"invoked holder literal", `$n=1; (function() use (&$n){$n=$_GET['b'];})(); probe($n);`, true},
		{"holder created after the sink", `$n=$_GET['a']; probe($n); $f=function() use (&$n){$n=1;}; $f();`, true},
		{"clean, holder created after the sink", `$n=1; probe($n); $f=function() use (&$n){$n=$_GET['b'];}; $f();`, false},
		{"constant passed to an unknown function", `$n=1; other($n); probe($n);`, false},
		{"global written with request data", `function swap(){global $n; $n=$_GET['b'];} $n=1; swap(); probe($n);`, true},
		{"global written twice", `function a(){global $n; $n=$_GET['a'];} function b(){global $n; $n=$_POST['b'];} $n=1; a(); probe($n);`, true},
		{"global written with a constant", `function swap(){global $n; $n=2;} $n=1; swap(); probe($n);`, false},
		{"tainted global written with a constant", `function swap(){global $n; $n=2;} $n=$_GET['a']; swap(); probe($n);`, true},
		{"other global written with request data", `function swap(){global $m; $m=$_GET['b'];} $n=1; swap(); probe($n);`, false},
		{"globals entry written with request data", `function swap(){$GLOBALS['n']=$_GET['b'];} $n=1; swap(); probe($n);`, true},
		{"other globals entry written", `function swap(){$GLOBALS['m']=$_GET['b'];} $n=1; swap(); probe($n);`, false},
		{"dynamic globals entry written", `function swap($k){$GLOBALS[$k]=$_GET['b'];} $n=1; swap('n'); probe($n);`, false},
		{"globals table read", `function peek(){return count($GLOBALS);} $n=$_GET['a']; peek(); probe($n);`, true},
		{"holder invoked as an object", `$n=1; $f=function() use (&$n){$n=$_GET['b'];}; $f->__invoke(); probe($n);`, true},
		{"holder bound and called", `$n=1; $f=function() use (&$n){$n=$_GET['b'];}; $f->call(new stdClass); probe($n);`, true},
		{"holder destructures request data", `$n=1; $f=function() use (&$n){[$n]=$_GET['b'];}; $f(); probe($n);`, true},
		{"holder stores request data in an element", `$n=[]; $f=function() use (&$n){$n['k']=$_GET['b'];}; $f(); probe($n['k']);`, true},
		{"holder appends request data", `$n='ls'; $f=function() use (&$n){$n.=$_GET['b'];}; $f(); probe($n);`, true},
		{"holder overwrites request data before it ends", `$n=1; $f=function() use (&$n){$n=$_GET['b']; $n=2;}; $f(); probe($n);`, false},
		{"nested holder", `$n=1; $f=function() use (&$n){$g=function() use (&$n){$n=$_GET['b'];}; $g();}; $f(); probe($n);`, true},
		{"function-level global", `function setcmd(){global $cmd; $cmd=$_GET['c'];} function run(){global $cmd; setcmd(); probe($cmd);}`, true},
		{"function-level global, clean writer", `function setcmd(){global $cmd; $cmd='ls';} function run(){global $cmd; setcmd(); probe($cmd);}`, false},
		{"dynamic globals entry written with request data", `function swap($k){$GLOBALS[$k]=$_GET['b'];} $n=1; swap('n'); probe($n);`, false},
		{"closure and global holder", `function swap(){global $n; $n=$_GET['b'];} $n=1; $f=function() use (&$n){$n=$_POST['c'];}; $f(); probe($n);`, true},
		{"alias write keeps earlier request data", `$q=&$n; $r=&$w; $n=$_GET['a']; $r=1; probe($n);`, true},
		{"alias write of a clean value", `$r=&$n; $n=1; $r=2; probe($n);`, false},
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

func TestInvalidationDropsEverythingButRequestData(t *testing.T) {
	e, f := environment(t, `$n=$_GET['a']; if(!$n){return;} $f=function() use (&$n){$n=1;}; $f(); probe($n);`, nil)
	x := argument(callsNamed(f, "probe")[0])
	if _, resolved := e.Resolve(x); resolved || e.Excludes(x, "false") || !e.Tainted(x, HTML) {
		t.Fatalf("value=%+v", e.Value(x))
	}
}

// Shared bindings are walked on every write and call: the work is charged to
// the scope budget, and each writer function is analysed once.
func TestSharedBindingsStayBounded(t *testing.T) {
	for name, build := range map[string]func(*strings.Builder){
		"global statements": func(src *strings.Builder) {
			src.WriteString("function w(){")
			for i := range 5000 {
				fmt.Fprintf(src, "global $g%d;", i)
			}
			src.WriteString("} $x=1; w(); probe($x);")
		},
		"reference aliases": func(src *strings.Builder) {
			for i := range 2000 {
				fmt.Fprintf(src, "$r%d=&$v%d; $v%d=%d;", i, i, i, i)
			}
			src.WriteString("probe($v0);")
		},
		"imports": func(src *strings.Builder) {
			for i := range 2000 {
				fmt.Fprintf(src, "$v%d=%d; extract($a%d);", i, i, i)
			}
			src.WriteString("probe($v0);")
		},
	} {
		t.Run(name, func(t *testing.T) {
			var src strings.Builder
			build(&src)
			e, f := environment(t, src.String(), nil)
			start := time.Now()
			e.Value(argument(callsNamed(f, "probe")[0]))
			if d := time.Since(start); d > testbudget.Of(2*time.Second) {
				t.Fatalf("flow took %s", d)
			}
		})
	}
}

// A function of ordinary length keeps its flow facts: a hundred locals carried
// through a few hundred blocks stay far below the budgets.
func longFunction(locals, blocks int) string {
	var src strings.Builder
	src.WriteString("function f($cfg, $rows){ $html=''; $out=[];")
	for i := range locals {
		fmt.Fprintf(&src, "$v%d=%d;", i, i)
	}
	for i := range blocks {
		switch i % 3 {
		case 0:
			fmt.Fprintf(&src, "if($v%d > %d){$out['k%d']=$v%d; $html.='<p>' . $v%d;}else{$v%d=%d;}", i%locals, i, i, (i+1)%locals, (i+2)%locals, (i+3)%locals, i)
		case 1:
			fmt.Fprintf(&src, "foreach($rows as $row){$html.=$row . $v%d; $out['n']+=1;}", i%locals)
		default:
			fmt.Fprintf(&src, "while($v%d < %d){$v%d++; $out[]=$v%d;}", i%locals, i, i%locals, (i+1)%locals)
		}
	}
	src.WriteString("$z=$_GET['z']; probe('tool ' . $z); $s=gzdecode($cfg['blob']); check($s); return $html;}")
	return src.String()
}

func TestLongFunctionsKeepTheirFacts(t *testing.T) {
	for _, size := range [][2]int{{100, 80}, {120, 150}, {200, 200}, {300, 400}} {
		e, f := environment(t, longFunction(size[0], size[1]), nil)
		start := time.Now()
		if !e.Tainted(argument(callsNamed(f, "probe")[0]), Shell) {
			t.Fatalf("%d locals, %d blocks: request data lost", size[0], size[1])
		}
		if _, resolved := e.Resolve(argument(callsNamed(f, "check")[0])); !resolved {
			t.Fatalf("%d locals, %d blocks: producer lost", size[0], size[1])
		}
		if d := time.Since(start); d > testbudget.Of(2*time.Second) {
			t.Fatalf("%d locals, %d blocks: flow took %s", size[0], size[1], d)
		}
	}
}

func BenchmarkLongFunction(b *testing.B) {
	e, f := environment(b, longFunction(100, 80), nil)
	x := argument(callsNamed(f, "probe")[0])
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if !New(f, e.types, nil).Tainted(x, Shell) {
			b.Fatal("request data lost")
		}
	}
}

// Overlaying one file answers every lookup like the rebuilt collection.
func TestSnapshotOverlayMatchesRebuild(t *testing.T) {
	file := func(path string, names ...string) *File {
		f := &File{Path: path, Summaries: map[string]Summary{}}
		for _, name := range names {
			f.Summaries[name] = Summary{Return: Value{Sources: []Source{{Kind: path + ":" + name}}}}
		}
		return f
	}
	base := NewSnapshot(file("a.php", "only_a", "a_and_b", "all", "moved"), file("b.php", "only_b", "a_and_b", "all"), file("c.php", "all", "only_c"))
	names := []string{"only_a", "only_b", "only_c", "a_and_b", "all", "moved", "fresh", "Only_C", "missing"}
	for _, own := range []*File{
		file("a.php", "only_a", "a_and_b", "all", "moved"),
		file("a.php", "fresh"),
		file("a.php", "only_b", "only_a"),
		file("a.php"),
		file("d.php", "fresh", "only_c", "all"),
		file("b.php", "moved", "only_b"),
	} {
		overlay := base.WithFile(own)
		rebuilt := NewSnapshot(append([]*File{own}, files(base, own.Path)...)...)
		for _, name := range names {
			got, found := overlay.lookup(name)
			want, expected := rebuilt.lookup(name)
			if found != expected || len(got.Return.Sources) != len(want.Return.Sources) || found && got.Return.Sources[0] != want.Return.Sources[0] {
				t.Errorf("%s over %v: %s = %+v/%v, want %+v/%v", own.Path, own.Summaries, name, got, found, want, expected)
			}
		}
		if !overlay.Equal(rebuilt) || !rebuilt.Equal(overlay) {
			t.Errorf("%s: overlay differs from the rebuilt collection", own.Path)
		}
		// An overlay is a snapshot like any other.
		again := overlay.WithFile(file("e.php", "only_e"))
		if _, ok := again.lookup("only_e"); !ok {
			t.Errorf("%s: second overlay lost its file", own.Path)
		}
		if v, ok := again.lookup("fresh"); ok != (len(own.Summaries) > 0 && has(own, "fresh")) || ok && v.Return.Sources[0].Kind != own.Path+":fresh" {
			t.Errorf("%s: second overlay lost the first", own.Path)
		}
		if _, ok := overlay.WithoutFile(own.Path).lookup("fresh"); ok {
			t.Errorf("%s: removed file kept its contract", own.Path)
		}
	}
	if _, ok := (*Snapshot)(nil).WithFile(file("a.php", "only_a")).lookup("only_a"); !ok {
		t.Error("file added to no snapshot")
	}
}

func files(s *Snapshot, except string) []*File {
	var out []*File
	for path, f := range s.files {
		if path != except {
			out = append(out, f)
		}
	}
	return out
}

func has(f *File, name string) bool { _, ok := f.Summaries[name]; return ok }

func BenchmarkSnapshotOverlay(b *testing.B) {
	var all []*File
	for i := range 2000 {
		f := &File{Path: fmt.Sprintf("f%d.php", i), Summaries: map[string]Summary{}}
		for j := range 10 {
			f.Summaries[fmt.Sprintf("f%d_%d", i, j)] = Summary{}
		}
		all = append(all, f)
	}
	base := NewSnapshot(all...)
	b.ReportAllocs()
	for b.Loop() {
		if _, ok := base.WithFile(all[7]).lookup("f7_3"); !ok {
			b.Fatal("contract lost")
		}
	}
}

func TestTaintOfNothing(t *testing.T) {
	e, _ := environment(t, `probe(1);`, nil)
	if e.Tainted(nil, HTML) {
		t.Fatal("absent expression")
	}
	if _, resolved := e.Resolve(nil); resolved {
		t.Fatal("absent expression resolved")
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
		{"another variable assigned inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && ($t=strlen($s)) > 1){probe($s);}}`, true, false},
		{"passed by reference inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && settype($s, 'bool')){probe($s);}}`, false, false},
		{"passed to an unknown function inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && check($s)){probe($s);}}`, false, false},
		{"passed by named value inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && strlen(string: $s) > 1){probe($s);}}`, true, false},
		{"callable created inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && strlen(...)){probe($s);}}`, true, false},
		{"stepped inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && $s++){probe($s);}}`, false, false},
		{"reassigned inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && ($s=gzdecode($b)) === false){probe($s);}}`, false, false},
		{"tested inside the condition", `function f($b){$s=gzdecode($b); if($s !== false && strlen($s) > 1){probe($s);}}`, true, false},
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
		{`$h=fopen('a','r'); $box['h']=$h; fread($h,1);`, false, false},
		{`$h=fopen('a','r'); $box['n']=1; fread($h,1);`, true, false},
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
