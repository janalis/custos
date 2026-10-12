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
			e, f := environment(t, source, nil)
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
