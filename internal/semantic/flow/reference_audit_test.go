package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestReferenceCallsCannotRetainPriorValueProof(t *testing.T) {
	for _, source := range []string{
		`function clean(&$value){$value='safe';}$x=$_GET['x'];clean($x);probe($x);`,
		`function clean(&$value){$value='safe';}$x=$_GET['x'];clean(value:$x);probe($x);`,
		`function clean(&$value){$value='safe';}$x=[$_GET['x']];clean($x[0]);probe($x);`,
		`$x=$_GET['x'];unknown($x);probe($x);`,
		`$x=$_GET['x'];unknown(value:$x);probe($x);`,
		`class Replace{function __construct(&$value){$value='safe';}}$x=$_GET['x'];new Replace(value:$x);probe($x);`,
		`$x=$_GET['x'];new $class($x);probe($x);`,
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
	e.invalidateReferences(&syntax.FuncCall{}, list, &f)
	if !f.vars["x"].Complete {
		t.Fatal("malformed independent arguments invalidated unrelated variable")
	}
}
