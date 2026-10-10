package flow

import (
	"strconv"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestRefinementRejectsUnprovedGuards(t *testing.T) {
	for _, guard := range []string{
		`unknown($x)`, `strlen('literal')`, `in_array($x, ['safe'], false)`,
		`in_array($x, [], true)`, `in_array($x, $values, true)`, `in_array($x, [$other], true)`,
		`in_array($x, [...$values], true)`, `in_array($x, [&$other], true)`,
		`str_contains($x,$needle)`, `str_contains('literal',"\r")`,
	} {
		e, f := environment(t, `$x=$_GET['x'];if (`+guard+`){probe($x);}`, nil)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		if v.Safe&Header != 0 || v.Safe&URL != 0 {
			t.Fatalf("%s proved unsafe validation: %+v", guard, v)
		}
	}
	e, f := environment(t, `$x=$_GET['x'];if (strlen($x)!==4){return;}probe($x);`, nil)
	if n, ok := e.ExactLength(argument(callsNamed(f, "probe")[0])); !ok || n != 4 {
		t.Fatalf("length=%d/%v", n, ok)
	}
	e, f = environment(t, `$x=unknown();if ($x===null){return;}probe($x);`, nil)
	if e.Excludes(argument(callsNamed(f, "probe")[0]), "null") {
		t.Fatal("unknown result became complete")
	}
	e, f = environment(t, `$x=preg_replace('/x/','y','z');if ($x===null){return;}probe($x);`, nil)
	x := argument(callsNamed(f, "probe")[0])
	if !e.Excludes(x, "null") || e.Excludes(x, "unsupported") {
		t.Fatal("strict null guard")
	}
}

func TestRequestStateAndUnknownEffects(t *testing.T) {
	e, f := environment(t, `ob_start();echo 'x';ob_end_flush();probe(1);ob_start();ob_end_clean();probe(2);session_start();unknown();probe(3);`, nil)
	calls := callsNamed(f, "probe")
	if _, ok := e.GlobalStateBefore(calls[0], "output"); !ok {
		t.Fatal("output state")
	}
	if _, ok := e.GlobalStateBefore(calls[0], "committed-output"); !ok {
		t.Fatal("committed state")
	}
	if _, ok := e.GlobalStateBefore(calls[1], "output-buffer"); ok {
		t.Fatal("ended buffer retained")
	}
	if _, ok := e.GlobalStateBefore(calls[2], "session"); ok {
		t.Fatal("unknown call retained request state")
	}
	if _, ok := e.GlobalStateBefore(calls[2], "unsupported"); ok {
		t.Fatal("unsupported request domain")
	}
	if !e.Value(nil).Complete {
		t.Fatal("nil expression")
	}
	if _, ok := e.StateBefore(calls[0], argument(calls[0])); ok {
		t.Fatal("scalar allocation")
	}
	e, f = environment(t, `switch($x){}probe(1);`, nil)
	if _, ok := e.GlobalStateBefore(callsNamed(f, "probe")[0], "output"); ok {
		t.Fatal("unsupported scope retained state")
	}
}

func TestUnsupportedAndExceptionalScopes(t *testing.T) {
	for _, src := range []string{
		`$x=$_GET[$key];$s=$_SERVER[$key];probe($s);`,
		`$x=$_FILES['a']['name'];$y=$_FILES['a']['unexpected'];probe($y);`,
		`${$name}='x';probe(${$name});`,
		`while($flag){break 2;}probe(1);`, `while($flag){continue 2;}probe(1);`,
		`$x='a';if($flag){$x='b';}elseif($other){$x='c';}else{$x='d';}probe($x);`,
		`$f=fn($x)=>$x;probe($f);`,
		`class C{static function f($x){return $x;}}C::f($_GET['x']);`,
		`Unknown::f();`,
	} {
		e, f := environment(t, src, nil)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if x, ok := n.(syntax.Expr); ok {
				e.Value(x)
			}
			return true
		})
	}
	deep := strings.Repeat("!", MaxDepth*8+1) + "$x"
	e, f := environment(t, `$x=$_GET['x'];probe(`+deep+`);`, nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("expression depth budget")
	}
	g := graph{blocks: make([]block, MaxBlocks), complete: true}
	if g.add(nil, 0, 0) != 0 || g.complete {
		t.Fatal("CFG budget")
	}
	if g.stmt(nil, 2, exits{}) != 2 {
		t.Fatal("nil statement continuation")
	}
}

func TestSnapshotsAreReplacedWithoutStaleContracts(t *testing.T) {
	a := &File{Path: "a.php", Summaries: map[string]Summary{"f": {Return: Value{Complete: true}}}}
	b := &File{Path: "b.php", Summaries: map[string]Summary{"g": {Return: Value{Complete: true}}}}
	s := NewSnapshot(a, b)
	replaced := s.WithFile(&File{Path: "a.php", Summaries: map[string]Summary{}})
	if _, ok := s.lookup("f"); !ok {
		t.Fatal("source snapshot mutated")
	}
	if _, ok := replaced.lookup("f"); ok {
		t.Fatal("stale replacement")
	}
	if _, ok := replaced.lookup("g"); !ok {
		t.Fatal("unrelated file lost")
	}
	if !s.Equal(NewSnapshot(a, b)) || s.Equal(nil) || !(*Snapshot)(nil).Equal(nil) {
		t.Fatal("snapshot equality")
	}
	if _, ok := s.WithoutFile("a.php").lookup("g"); !ok {
		t.Fatal("unrelated removal")
	}
	e, f := environment(t, `class C { function f($x) { $inner=function(){yield 1;};return $x;} } function input(){return $_GET['x'];} function sink(){system($_GET['x']);} function route($x){header('Location: '.$x);} function net($h,$x){curl_setopt($h,CURLOPT_URL,$x);}`, nil)
	file := Extract(f, e.types.Index, phpversion.Max, nil)
	if file.Summaries["c::f"].Generator {
		t.Fatal("nested deferred scope escaped")
	}
	snapshot := NewSnapshot(file)
	local, g := environment(t, `$x=input();probe($x);sink();route($_GET['x']);net(curl_init(),$_GET['x']);`, snapshot)
	local.types.Index = e.types.Index
	if !local.Tainted(argument(callsNamed(g, "probe")[0]), HTML) {
		t.Fatal("external return summary")
	}
	if sinks := local.Sinks(callsNamed(g, "sink")[0]); len(sinks) != 1 || len(sinks[0].Argument.Sources) != 1 {
		t.Fatalf("external sink summary: %+v", sinks)
	}
	if sinks := local.Sinks(callsNamed(g, "route")[0]); len(sinks) != 2 {
		t.Fatalf("location contexts: %+v", sinks)
	}
	if sinks := local.Sinks(callsNamed(g, "net")[0]); len(sinks) != 1 || sinks[0].Context != URL {
		t.Fatalf("curl option context: %+v", sinks)
	}
}

func TestIncompleteSummaryArgumentsCannotProveTaint(t *testing.T) {
	e, f := environment(t, `function f($x){return $x;}function sink($x){system($x);}`, nil)
	snapshot := NewSnapshot(Extract(f, e.types.Index, phpversion.Max, nil))
	local, g := environment(t, `probe(f());sink();`, snapshot)
	local.types.Index = e.types.Index
	if local.Tainted(argument(callsNamed(g, "probe")[0]), HTML) {
		t.Fatal("missing required argument")
	}
	if sinks := local.Sinks(callsNamed(g, "sink")[0]); len(sinks) != 0 {
		t.Fatalf("missing sink argument: %+v", sinks)
	}
}

func TestMalformedOrDynamicCallsDiscardContracts(t *testing.T) {
	e, f := environment(t, `function noargs(){} noargs();$object->$method();Unknown::$method();`, nil)
	if _, ok := e.BoundArguments(f.Stmts[0]); ok {
		t.Fatal("non-call contract")
	}
	call := callsNamed(f, "noargs")[0]
	copy := *call
	copy.Args = nil
	if _, ok := e.bind(&copy, nil); !ok {
		t.Fatal("zero argument contract")
	}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.MethodCall, *syntax.StaticCall:
			if _, _, ok := e.declaration(n); ok {
				t.Fatal("dynamic method declaration")
			}
		}
		return true
	})
	for _, raw := range []string{"", "bare"} {
		if _, ok := literalText(&syntax.Literal{LitKind: syntax.LitString, Raw: raw}); ok {
			t.Fatal("malformed literal accepted")
		}
	}
	frame := frame{vars: map[string]Value{}, states: map[uint32]State{}}
	e.refine(nil, &frame, true)
	e, f = environment(t, `$x=$_GET['x'];in_array($x,['a'],true);str_contains($x,"\r");`, nil)
	arrayCall := *callsNamed(f, "in_array")[0]
	args := *arrayCall.Args
	arrayCall.Args = &args
	args.Args = append([]syntax.Expr(nil), args.Args...)
	args.Args[1] = &syntax.Literal{Raw: "invalid"}
	e.refine(&arrayCall, &frame, true)
	args.Args[0] = &syntax.Literal{Raw: "invalid"}
	e.refine(&arrayCall, &frame, true)
	contains := *callsNamed(f, "str_contains")[0]
	containsArgs := *contains.Args
	contains.Args = &containsArgs
	containsArgs.Args = append([]syntax.Expr(nil), containsArgs.Args...)
	containsArgs.Args[1] = &syntax.Literal{Raw: "invalid"}
	e.refine(&contains, &frame, false)
}

func TestFrameJoinsRequireAgreement(t *testing.T) {
	a := frame{vars: map[string]Value{"x": {Complete: true, Sources: []Source{{Kind: "a"}}}}, states: map[uint32]State{1: {Operation: "open"}}}
	b := frame{vars: map[string]Value{"x": {Complete: true, Sources: []Source{{Kind: "b"}}}}, states: map[uint32]State{1: {Operation: "close"}}}
	if same(a, b) {
		t.Fatal("different origins agreed")
	}
	b.vars = a.vars
	if same(a, b) {
		t.Fatal("different operations agreed")
	}
	joined := join(a, frame{vars: map[string]Value{}, states: map[uint32]State{}})
	if joined.vars["x"].Complete || len(joined.states) != 0 {
		t.Fatal("missing path retained definite facts")
	}
}

func TestTransferBudgetIncludesStraightLineExpressions(t *testing.T) {
	e, f := environment(t, "probe(["+strings.Repeat("1,", MaxTransfers+1)+"]);", nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("straight-line expression budget exhausted without invalidating proof")
	}
	e, f = environment(t, "probe([$_GET['x']]);probe([&$unknown]);", nil)
	calls := callsNamed(f, "probe")
	if !e.Tainted(argument(calls[0]), HTML) || e.Value(argument(calls[1])).Complete {
		t.Fatal("array elements did not retain precise provenance")
	}
}

func TestRetainedStateBudgetInvalidatesOversizedScope(t *testing.T) {
	var source strings.Builder
	for i := 0; i < 500; i++ {
		source.WriteString("$h")
		source.WriteString(strings.Repeat("x", i))
		source.WriteString("=tmpfile();")
	}
	source.WriteString("probe($h);")
	e, f := environment(t, source.String(), nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("retained state budget established definite value")
	}
}

func TestBranchFrameBudgetInvalidatesWideScopes(t *testing.T) {
	var source strings.Builder
	for i := 0; i < 10000; i++ {
		source.WriteString("$v")
		source.WriteString(strings.Repeat("x", i%30))
		source.WriteString("_")
		source.WriteString(strconv.Itoa(i))
		source.WriteString("=1;")
	}
	for i := 0; i < 20; i++ {
		source.WriteString("if($flag){$changed=1;}")
	}
	source.WriteString("probe($v_0);")
	e, f := environment(t, source.String(), nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("wide branch frames established proof beyond budget")
	}
}

func TestIncompleteArrayItemsCannotEstablishTaint(t *testing.T) {
	e, _ := environment(t, "", nil)
	r := e.scope(nil)
	frame := frame{vars: map[string]Value{}, states: map[uint32]State{}}
	value := e.eval(&syntax.Array{Items: []*syntax.ArrayItem{nil}}, &frame, r, 0)
	if !value.Complete {
		t.Fatal("omitted item discarded complete empty array")
	}
}

func TestResourceEscapesSeparateOperationHistories(t *testing.T) {
	for _, source := range []string{
		`$h=fopen('a','r');$alias=$h;fclose($h);unknown($alias);probe($h);fclose($h);fread($h,1);`,
		`$h=fopen('a','r');$box=[$h];fclose($h);unknown($box);probe($h);fclose($h);fread($h,1);`,
		`$h=fopen('a','r');fclose($h);unknown(function()use($h){});probe($h);fclose($h);fread($h,1);`,
	} {
		e, f := environment(t, "function probe($value){} "+source, nil)
		probe := callsNamed(f, "probe")[0]
		value := e.Value(argument(probe))
		if value.Identity == 0 || value.Invalidated == 0 {
			t.Fatal("escaped identity retained stale history")
		}
		if _, ok := e.StateBefore(probe, argument(probe), "fclose"); ok {
			t.Fatal("escaped close retained")
		}
		read := callsNamed(f, "fread")[0]
		if _, ok := e.StateBefore(read, argument(read), "fclose"); !ok {
			t.Fatal("known close after escape failed to establish state")
		}
	}
	e, f := environment(t, `$h=fopen('a','r');if($flag){unknown($h);}probe($h);`, nil)
	if value := e.Value(argument(callsNamed(f, "probe")[0])); value.Complete {
		t.Fatal("conditional reference escape retained definite identity")
	}
}

func TestArgumentOriginWalkSharesExecutionBudget(t *testing.T) {
	e, f := environment(t, "$a=["+strings.Repeat("1,", MaxTransfers/2)+"];unknown($a);probe($a);", nil)
	if e.Value(argument(callsNamed(f, "probe")[0])).Complete {
		t.Fatal("argument-origin walk escaped execution budget")
	}
}
