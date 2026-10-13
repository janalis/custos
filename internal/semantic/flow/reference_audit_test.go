package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestReferenceCallsCannotRetainPriorValueProof(t *testing.T) {
	for _, source := range []string{
		`function clean(&$value){$value='safe';}$x=$_GET['x'];clean($x);probe($x);`,
		`function clean(&$value){$value='safe';}$x=$_GET['x'];clean(value:$x);probe($x);`,
		`class Replace{function __construct(&$value){$value='safe';}}$x=$_GET['x'];new Replace(value:$x);probe($x);`,
		`$x=$_GET['x'];preg_match('/a/','a',matches:$x);probe($x);`,
	} {
		t.Run(source, func(t *testing.T) {
			e, f := summarized(t, source)
			x := argument(callsNamed(f, "probe")[0])
			if e.Value(x).Complete || e.Tainted(x, HTML) {
				t.Fatalf("reference call retained stale value: %+v", e.Value(x))
			}
		})
	}
	// A reference to one element replaces that element at most: the array
	// keeps the request data its other elements may hold, and stays an array.
	for _, tc := range []struct {
		src               string
		tainted, nonFalse bool
	}{
		{`function clean(&$value){$value='safe';}$x=[$_GET['x'], $_GET['y']];clean($x[0]);probe($x);`, true, false},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} $o->m($u['host']); probe($u);}`, false, true},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} $o->m($u['a']['b']); probe($u);}`, false, true},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} lookup($u['host']); probe($u);}`, false, true},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} preg_match('/a/', $a, $u['m']); probe($u);}`, false, true},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} $o->m($u); probe($u);}`, false, false},
		{`function f($o, $a){$u=fopen($a,'r'); if($u === false){return;} $o->m($u['host']->p); probe($u);}`, false, true},
		{`function f($o, $a){$u=['k'=>'v']; preg_match('/a/', $_GET['s'], $u['m']); probe($u);}`, true, false},
		{`function f($o, $a){$u=['k'=>'v', 'c'=>$_GET['c']]; $o->m($u['k']); probe($u['k']);}`, true, false},
		// The elements beside the one passed keep what was known of them.
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>$_GET['c']]; $o->m($u['k']); probe($u['c']);}`, false, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>['j'=>$_GET['c']]]; $o->m($u['k']['j']); probe($u['c']);}`, false, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>$_GET['c']]; $o->m($u[$a]); probe($u['c']);}`, true, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>['j'=>$_GET['c']]]; $o->m($u[$a]['j']); probe($u['c']);}`, true, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>$_GET['c']]; $o->m($u['c']); probe($u['c']);}`, true, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>$_GET['c']]; $o->m($u); probe($u['c']);}`, true, false},
		{`function f($o, $a){$u=['k'=>$_GET['c']]; $o->m($u['k']); probe($u);}`, true, false},
		{`function f($o, $a){$u=['c'=>'ls', 'k'=>$_GET['c']]; $r=&$u; $o->m($r['k']); probe($u['c']);}`, true, false},
	} {
		e, f := environment(t, tc.src, nil)
		x := argument(callsNamed(f, "probe")[0])
		if e.Tainted(x, HTML) != tc.tainted || e.Excludes(x, "false") != tc.nonFalse {
			t.Errorf("%s: value=%+v", tc.src, e.Value(x))
		}
		// What the array held is no longer known once an element was bound.
		if _, resolved := e.Resolve(x); resolved && strings.Contains(tc.src, "['host']); probe($u)") {
			t.Errorf("%s: resolved", tc.src)
		}
	}
	// An unresolved callee may take the argument by reference, so its value
	// is no longer known; it may also leave it alone, so the request data
	// stays possible.
	for _, source := range []string{
		`$x=$_GET['x'];unknown($x);probe($x);`,
		`$x=$_GET['x'];unknown(value:$x);probe($x);`,
		`$x=$_GET['x'];new $class($x);probe($x);`,
	} {
		e, f := environment(t, source, nil)
		x := argument(callsNamed(f, "probe")[0])
		if _, resolved := e.Resolve(x); resolved || !e.Tainted(x, HTML) {
			t.Fatalf("%s: value %+v", source, e.Value(x))
		}
	}
	for _, source := range []string{
		`function read($value){}$x=$_GET['x'];read($x);probe($x);`,
		`function read($value){}$x=$_GET['x'];read(value:$x);probe($x);`,
		`class Read{function __construct($value){}}$x=$_GET['x'];new Read($x);probe($x);`,
	} {
		e, f := environment(t, source, nil)
		if !e.Tainted(argument(callsNamed(f, "probe")[0]), HTML) {
			t.Fatal("known by-value parameter discarded input proof")
		}
	}
}

func TestConstructorEscapesInvalidateResourceAndSessionHistory(t *testing.T) {
	for _, source := range []string{
		`class Reopen{function __construct(&$handle){$handle=fopen('other','r');}}$h=fopen('original','r');fclose($h);new Reopen($h);fread($h,1);`,
		`class Reader{function __construct($handle){}}$h=fopen('original','r');fclose($h);new Reader($h);fread($h,1);`,
		`$h=fopen('original','r');fclose($h);new Unknown($h);fread($h,1);`,
	} {
		e, f := environment(t, source, nil)
		read := callsNamed(f, "fread")[0]
		if _, known := e.StateBefore(read, argument(read), "fclose"); known {
			t.Fatal("constructor escape retained earlier close")
		}
	}
	e, f := environment(t, `session_write_close();new Unknown;probe();`, nil)
	if _, known := e.GlobalStateBefore(callsNamed(f, "probe")[0], "session"); known {
		t.Fatal("unknown constructor retained global session state")
	}
}

func TestMalformedReferenceArgumentsDoNotInventWrites(t *testing.T) {
	e, _ := environment(t, "", nil)
	f := frame{vars: map[string]Value{"x": {Complete: true}}}
	list := &syntax.ArgList{Args: []syntax.Expr{&syntax.Literal{Raw: "1"}, &syntax.Arg{Value: &syntax.Literal{Raw: "2"}}}}
	e.invalidateReferences(&syntax.FuncCall{}, list, &f, &scopeResult{})
	if !f.vars["x"].Complete {
		t.Fatal("malformed independent arguments invalidated unrelated variable")
	}
}

// summarized returns an environment that knows the summaries of its own file,
// as every analysed file does.
func summarized(t testing.TB, src string) (*Env, *syntax.File) {
	t.Helper()
	e, f := environment(t, src, nil)
	snapshot := NewSnapshot()
	for range 3 {
		snapshot = NewSnapshot(Extract(f, e.types.Index, e.types.PHP, snapshot))
	}
	return New(f, e.types, snapshot), f
}

// What a function of the project leaves in a reference argument comes from
// its summary: the request data it writes, the arguments it copies there,
// or nothing when every path overwrites it with clean data.
func TestReferenceParametersCarryWhatTheCalleeLeaves(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"callee writes request data", `function fill(&$x){$x=$_GET['a'];} $v='ls'; fill($v); probe($v);`, true},
		{"callee writes request data on one path", `function fill(&$x, $c){if($c){$x=$_GET['a'];}} $v='ls'; fill($v, 1); probe($v);`, true},
		{"callee overwrites with clean data", `function clean(&$x){$x='safe';} $v=$_GET['a']; clean($v); probe($v);`, false},
		{"callee overwrites on one path only", `function clean(&$x, $c){if($c){$x='safe';}} $v=$_GET['a']; clean($v, 1); probe($v);`, true},
		{"callee never writes", `function look(&$x){return strlen($x);} $v=$_GET['a']; look($v); probe($v);`, true},
		{"callee never writes a clean value", `function look(&$x){return strlen($x);} $v='ls'; look($v); probe($v);`, false},
		{"callee escapes in place", `function esc(&$x){$x=htmlspecialchars($x);} $v=$_GET['a']; esc($v); probe($v);`, false},
		{"callee copies another argument", `function put(&$x, $y){$x=$y;} $v='ls'; put($v, $_GET['a']); probe($v);`, true},
		{"callee copies a clean argument", `function put(&$x, $y){$x=$y;} $v=$_GET['a']; put($v, 'ls'); probe($v);`, false},
		{"callee writes an element", `function fill(&$x){$x['k']=$_GET['a'];} $v=[]; fill($v); probe($v);`, true},
		{"element argument", `function fill(&$x){$x=$_GET['a'];} $v=['c'=>'ls', 'u'=>'x']; fill($v['c']); probe($v['c']);`, true},
		{"element argument, sibling", `function fill(&$x){$x=$_GET['a'];} $v=['c'=>'ls', 'u'=>'x']; fill($v['u']); probe($v['c']);`, false},
		{"callee passes it on", `function fill(&$x){$x=$_GET['a'];} function outer(&$y){fill($y);} $v='ls'; outer($v); probe($v);`, true},
		{"recursive callee", `function rec(&$x, $n){if($n){rec($x, $n-1);} $x=$_GET['a'];} $v='ls'; rec($v, 2); probe($v);`, true},
		{"callee throws before writing", `function fill(&$x){if(!$x){throw new E;} $x='safe';} $v=$_GET['a']; try{fill($v);}catch(E $e){} probe($v);`, true},
		{"conditionally declared callee", `if(!function_exists('clean')){function clean(&$x){$x='safe';}} $v=$_GET['a']; clean($v); probe($v);`, true},
		{"generator callee", `function gen(&$x){$x='safe'; yield 1;} $v=$_GET['a']; gen($v); probe($v);`, true},
		// A method called on an object is the one resolved unless a class of
		// the project extending the receiver's declares another.
		{"replaced method overwrites", `class C{function clean(&$x){$x='safe';}} class D extends C{function clean(&$x){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"method replaced two classes below", `class C{function clean(&$x){$x='safe';}} class D extends C{} class E extends D{function clean(&$x){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"method replaced through a trait", `trait T{function clean(&$x){}} class C{function clean(&$x){$x='safe';}} class D extends C{use T;} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"method replaced in an anonymous class", `class C{function clean(&$x){$x='safe';}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);} f(new class extends C{function clean(&$x){}});`, true},
		{"method replaced under an alias of an alias", `class C{function clean(&$x){$x='safe';}} class_alias('C', 'CA'); class_alias('CA', 'CB'); class_alias('CB', 'CA'); class D extends CB{function clean(&$x){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"private method called from outside its class", `class C{private function clean(&$x){$x='safe';} function __call($n,$a){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"method replaced under an alias of the class", `class C{function clean(&$x){$x='safe';}} class_alias('C', 'CAlias'); class D extends CAlias{function clean(&$x){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"method replaced in a second declaration", `class C{function clean(&$x){$x='safe';}} if(PHP_VERSION_ID < 1){class D{}}else{class D extends C{function clean(&$x){}}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"receiver created here", `class C{function clean(&$x){$x='safe';}} function f(){$c=new C; $v=$_GET['a']; $c->clean($v); probe($v);}`, false},
		{"receiver replaced through a reference", `class C{function clean(&$x){$x='safe';}} class O{function clean(&$x){}} function f(){$c=new C; $r=&$c; $r=new O; $v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"receiver passed to an unresolved callee by reference", `class C{function clean(&$x){$x='safe';}} function f(){$c=new C; swap($c); $v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"receiver kept in a property", `class C{function clean(&$x){$x='safe';}} class H{private C $c; function f(){$v=$_GET['a']; $this->c->clean($v); probe($v);}}`, false},
		{"classes extending each other", `class C extends D{function clean(&$x){$x='safe';}} class D extends C{} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, false},
		{"method no class replaces", `class C{function clean(&$x){$x='safe';}} class D extends C{function other(){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, false},
		{"method replaced beside the receiver's class", `class B{function clean(&$x){$x='safe';}} class C extends B{} class D extends B{function clean(&$x){}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, false},
		{"method of an unknown receiver class", `function f(Unknown $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, true},
		{"trait method through this", `trait T{function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; $this->clean($v); probe($v);}}`, true},
		{"replaceable method writes request data", `class C{function fill(&$x){$x=$_GET['a'];}} function f(C $c){$v='ls'; $c->fill($v); probe($v);}`, true},
		{"static method named by its class", `class V{static function clean(&$x){$x='safe';}} $v=$_GET['a']; V::clean($v); probe($v);`, false},
		{"static method bound late, replaced", `class V{static function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; static::clean($v); probe($v);}} class W extends V{static function clean(&$x){}}`, true},
		{"static method bound late, not replaced", `class V{static function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; static::clean($v); probe($v);}}`, false},
		{"method of a final class", `final class W{function clean(&$x){$x='safe';}} function f(W $w){$v=$_GET['a']; $w->clean($v); probe($v);}`, false},
		{"numbers left in the argument", `function ints(&$x){$x=array_map('intval', $x);} $v=$_GET['ids']; ints($v); probe(implode(',', $v));`, false},
		{"numbers left in the argument, keys", `function ints(&$x){$x=array_map('intval', $x);} $v=$_GET['ids']; ints($v); probe(implode(',', array_keys($v)));`, true},
		{"numbers left by a replaced method", `class C{function ints(&$x){$x=array_map('intval', $x);}} class D extends C{function ints(&$x){}} function f(C $c){$v=$_GET['ids']; $c->ints($v); probe(implode(',', $v));}`, true},
		{"final method overwrites", `class C{final function clean(&$x){$x='safe';}} function f(C $c){$v=$_GET['a']; $c->clean($v); probe($v);}`, false},
		{"private method overwrites", `class C{private function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; $this->clean($v); probe($v);}}`, false},
		// `self::` and `parent::` name the class whose method runs, declared
		// there or in a class above it.
		{"static method through self", `class C{private static function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; self::clean($v); probe($v);}}`, false},
		{"open method through self", `class C{function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; self::clean($v); probe($v);}}`, false},
		{"inherited method through self", `class P{static function clean(&$x){$x='safe';}} class C extends P{function f(){$v=$_GET['a']; self::clean($v); probe($v);}}`, false},
		{"inherited method through self writes request data", `class P{static function fill(&$x){$x=$_GET['a'];}} class C extends P{function f(){$v='ls'; self::fill($v); probe($v);}}`, true},
		{"overridden method through parent", `class P{function clean(&$x){$x='safe';}} class C extends P{function clean(&$x){} function f(){$v=$_GET['a']; parent::clean($v); probe($v);}}`, false},
		{"method of the grandparent through parent", `class G{function clean(&$x){$x='safe';}} class P extends G{} class C extends P{function f(){$v=$_GET['a']; parent::clean($v); probe($v);}}`, false},
		{"parent without a parent class", `class C{function f(){$v=$_GET['a']; parent::clean($v); probe($v);}}`, true},
		{"self in a trait", `trait T{static function clean(&$x){$x='safe';} function f(){$v=$_GET['a']; self::clean($v); probe($v);}}`, true},
		{"inherited final method through this", `class P{final function clean(&$x){$x='safe';}} class C extends P{function f(){$v=$_GET['a']; $this->clean($v); probe($v);}}`, false},
		{"inherited open method through this", `class P{function clean(&$x){$x='safe';}} class C extends P{function f(){$v=$_GET['a']; $this->clean($v); probe($v);}}`, false},
		{"inherited method through this, replaced below", `class P{function clean(&$x){$x='safe';}} class C extends P{function f(){$v=$_GET['a']; $this->clean($v); probe($v);}} class D extends C{function clean(&$x){}}`, true},
		{"inherited open method through this writes request data", `class P{function fill(&$x){$x=$_GET['a'];}} class C extends P{function f(){$v='ls'; $this->fill($v); probe($v);}}`, true},
		{"constructor overwrites", `class C{function __construct(&$x){$x='safe';}} $v=$_GET['a']; new C($v); probe($v);`, false},
		{"callee calls into the project", `function look(&$x){other();} $v=$_GET['a']; look($v); probe($v);`, true},
		{"callee calls into the project, clean", `function look(&$x){other();} $v='ls'; look($v); probe($v);`, false},
		{"callee unsets then writes", `function f(&$x){unset($x); $x='ls';} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee rebinds to a local", `function f(&$x){$y='ls'; $x=&$y;} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee rebinds to an element", `function f(&$x, $l){$x=&$l['k']; $x='ls';} $v=$_GET['a']; f($v, []); probe($v);`, true},
		{"callee iterates by reference with it", `function f(&$x, $l){foreach($l as &$x){$x='ls';}} $v=$_GET['a']; f($v, []); probe($v);`, true},
		{"callee makes it static", `function f(&$x){static $x; $x='ls';} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee makes it global", `function f(&$x){global $x; $x='ls';} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee writes request data, then unsets", `function f(&$x){$x=$_GET['z']; unset($x); $x='ls';} $v='ls'; f($v); probe($v);`, true},
		{"callee overwrites, then unsets", `function f(&$x){$x='safe'; unset($x);} $v=$_GET['a']; f($v); probe($v);`, false},
		{"callee unsets on one path", `function f(&$x, $c){if($c){unset($x);} $x='ls';} $v=$_GET['a']; f($v, 1); probe($v);`, true},
		{"callee unsets on one path, written before", `function f(&$x, $c){$x='safe'; if($c){unset($x);} $x='ls';} $v=$_GET['a']; f($v, 1); probe($v);`, false},
		{"callee writes through a bound name after unsetting", `function f(&$x){$y=&$x; unset($x); $y=$_GET['z'];} $v='ls'; f($v); probe($v);`, true},
		{"callee overwrites through a bound name after unsetting", `function f(&$x){$y=&$x; unset($x); $y='safe';} $v=$_GET['a']; f($v); probe($v);`, false},
		{"callee unsets twice", `function f(&$x){unset($x); $x=$_GET['z']; unset($x);} $v='ls'; f($v); probe($v);`, false},
		{"callee rebinds through a destructuring", `function f(&$x){$a=['ls']; [&$x]=$a;} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee rebinds through list()", `function f(&$x){$a=['ls']; list(&$x)=$a; $x='pwd';} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee replaces its reference variadic", `function f(&...$xs){$xs=['ls'];} $v=$_GET['a']; f($v); probe($v);`, true},
		{"callee writes an element of its reference variadic", `function f(&...$xs){$xs[0]=$_GET['z'];} $v='ls'; f($v); probe($v);`, true},
		{"callee leaves its reference variadic", `function f(&...$xs){return count($xs);} $v='ls'; f($v); probe($v);`, false},
		{"closure unsets its reference capture", `$v='ls'; $g=function() use (&$v){$v=$_GET['z']; unset($v);}; $g(); probe($v);`, true},
		{"named argument", `function fill($c, &$x){$x=$_GET['a'];} $v='ls'; fill(x: $v, c: 1); probe($v);`, true},
		{"callee too large to summarise", `function big(&$x){` + strings.Repeat("$x=[$x,$x];", 2) + `eval('1;'); $x='safe';} $v=$_GET['a']; big($v); probe($v);`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := summarized(t, tc.src)
			calls := callsNamed(f, "probe")
			if got := e.Tainted(argument(calls[len(calls)-1]), HTML); got != tc.taint {
				t.Fatalf("taint=%v value=%+v", got, e.Value(argument(calls[len(calls)-1])))
			}
		})
	}
}
