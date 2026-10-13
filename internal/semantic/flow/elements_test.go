package flow

import (
	"fmt"
	"strings"
	"testing"
)

// The elements tracked for an array belong to its value: they follow it into
// another variable and leave with it when the variable is given another.
func TestElementsFollowTheirArray(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"replaced by an array holding request data", `$a=['c'=>'ls']; $b=['c'=>$_GET['x']]; $a=$b; probe($a['c']);`, true},
		{"replaced by a clean array", `$a=['c'=>$_GET['x']]; $b=['c'=>'ls']; $a=$b; probe($a['c']);`, false},
		{"copy written afterwards", `$a=['c'=>'ls']; $b=$a; $b['c']=$_GET['x']; probe($a['c']);`, false},
		{"either of two arrays", `$a=['c'=>'ls']; $b=['c'=>$_GET['x']]; $a=$k ? $b : $a; probe($a['c']);`, true},
		{"either of two arrays, element clean in both", `$a=['c'=>'ls','k'=>$_GET['x']]; $b=['c'=>'pwd','k'=>$_GET['y']]; $d=$k ? $a : $b; probe($d['c']);`, false},
		{"either of two arrays, element in one only", `$a=['c'=>'ls','k'=>$_GET['x']]; $b=['k'=>$_GET['y']]; $d=$k ? $a : $b; probe($d['c']);`, true},
		{"short choice", `$b=['c'=>'ls','k'=>$_GET['x']]; $a=$b ?: []; probe($a['c']);`, true},
		{"array cast", `$a=['c'=>'ls']; $b=['c'=>$_GET['x']]; $a=(array) $b; probe($a['c']);`, true},
		{"array cast of a clean element", `$b=['c'=>'ls','k'=>$_GET['x']]; $a=(array) $b; probe($a['c']);`, false},
		{"another operator", `$b=['c'=>'ls','k'=>$_GET['x']]; $a=@$b; probe($a['c']);`, true},
		{"chained assignment", `$a=$b=['c'=>'ls','k'=>$_GET['x']]; probe($a['c']);`, true},
		{"superglobal", `$a=['c'=>'ls']; $a=$_POST; probe($a['c']);`, true},
		{"paths with no element in common", `if($f){$a=['c'=>'ls','k'=>$_GET['x']];}else{$a=['d'=>'ls','j'=>$_GET['x']];} probe($a['c']);`, true},
		{"paths with one element in common", `if($f){$a=['c'=>'ls','k'=>$_GET['x']];}else{$a=['d'=>'ls','k'=>$_GET['x']];} probe($a['c']);`, true},
		{"paths with the same elements", `if($f){$a=['c'=>'ls','k'=>$_GET['x']];}else{$a=['c'=>'ls','k'=>$_GET['x']];} probe($a['c']);`, false},
		{"pinned by a comparison", `$a=['c'=>$_GET['x']]; if($a === 'none'){$a['k']=1; probe($a['c']);}`, false},
		{"copy of an array with a referenced element", `$a=['c'=>'ls']; $s=&$a['c']; $b=$a; $s=$_GET['x']; probe($b['c']);`, true},
		{"copy taken before the reference", `$a=['c'=>'ls']; $b=$a; $s=&$a['c']; $s=$_GET['x']; probe($b['c']);`, false},
		{"written through a bound name", `$a=['c'=>'ls']; $b=&$a; $b['c']=$_GET['x']; probe($a['c']);`, true},
		{"other element written through a bound name", `$a=['c'=>'ls']; $b=&$a; $b['k']=$_GET['x']; probe($a['c']);`, false},
		{"literal assigned through a bound name", `$a=['c'=>'ls']; $b=&$a; $b=['c'=>$_GET['x']]; probe($a['c']);`, true},
		{"more elements than are tracked", `$a=[]; $a['c']='ls';` + manyKeys(MaxSources) + `$a['k']=$_GET['x']; probe($a['c']);`, true},
		{"as many elements as are tracked", `$a=[]; $a['c']='ls';` + manyKeys(MaxSources-2) + `$a['k']=$_GET['x']; probe($a['c']);`, false},
		// PHP folds a constant offset to an integer or a string key.
		{"hexadecimal key", `$a=[1=>'ls']; $a[0x1]=$_GET['x']; probe($a[1]);`, true},
		{"octal key", `$a=[1=>'ls']; $a[01]=$_GET['x']; probe($a[1]);`, true},
		{"separated digits", `$a=[10=>'ls']; $a[1_0]=$_GET['x']; probe($a[10]);`, true},
		{"float key", `$a=[1=>'ls']; $a[1.7]=$_GET['x']; probe($a[1]);`, true},
		{"negative float key", `$a=[-1=>'ls']; $a[-1.7]=$_GET['x']; probe($a[-1]);`, true},
		{"float key out of range", `$a=[1=>'ls']; $a[1e30]=$_GET['x']; probe($a[1]);`, true},
		{"integer key out of range", `$a=[1=>'ls']; $a[0x1ffffffffffffffff]=$_GET['x']; probe($a[1]);`, true},
		{"true key", `$a=[1=>'ls']; $a[true]=$_GET['x']; probe($a[1]);`, true},
		{"false key", `$a=['ls']; $a[FALSE]=$_GET['x']; probe($a[0]);`, true},
		{"null key", `$a=[''=>'ls']; $a[null]=$_GET['x']; probe($a['']);`, true},
		{"other constant key", `$a=[1=>'ls']; $a[KEY]=$_GET['x']; probe($a[1]);`, true},
		{"negated constant key", `$a=[1=>'ls']; $a[-true]=$_GET['x']; probe($a[1]);`, true},
		{"integer string key", `$a=[1=>'ls']; $a['1']=$_GET['x']; probe($a[1]);`, true},
		{"integer key read as a string", `$a=['1'=>'ls']; $a[1]=$_GET['x']; probe($a['1']);`, true},
		{"negative integer key", `$a=['-1'=>'ls']; $a[-1]=$_GET['x']; probe($a['-1']);`, true},
		{"negated string key", `$a=[1=>'ls']; $a[-'1']=$_GET['x']; probe($a[1]);`, true},
		{"string that only looks like an integer", `$a=['01'=>'ls']; $a[1]=$_GET['x']; probe($a['01']);`, false},
		{"other integer key", `$a=[1=>'ls']; $a[0x2]=$_GET['x']; probe($a[1]);`, false},
		{"escaped string key", `$a=['a b'=>'ls']; $a["a\x20b"]=$_GET['x']; probe($a['a b']);`, true},
		// The globals table entry of a variable in scope is that variable.
		{"globals entry at file level", `$x=$_GET['a']; probe($GLOBALS['x']);`, true},
		{"clean globals entry at file level", `$x='ls'; $y=$_GET['a']; probe($GLOBALS['x']);`, false},
		{"globals entry of a bound global", `function f(){global $x; $x=$_GET['a']; probe($GLOBALS['x']);}`, true},
		{"globals entry of a local", `function f(){$x=$_GET['a']; probe($GLOBALS['x']);}`, false},
		{"element of a globals entry", `$x=['k'=>$_GET['a']]; probe($GLOBALS['x']['k']);`, true},
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

func manyKeys(n int) string {
	var src strings.Builder
	for i := range n {
		fmt.Fprintf(&src, "$a['k%d']=1;", i)
	}
	return src.String()
}

// The server chooses some fields of an uploaded file. That holds for an
// entry of the uploads table, not for any array a field was copied into.
func TestUploadFieldsFollowTheEntry(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"size of a listed entry", `foreach($_FILES as $f){probe($f['size']);}`, false},
		{"name of a listed entry", `foreach($_FILES as $f){probe($f['name']);}`, true},
		{"name stored under a server field name", `foreach($_FILES as $f){$meta=['size'=>$f['name']]; probe($meta['size']);}`, true},
		{"server field overwritten with the name", `foreach($_FILES as $f){$f['size']=$f['name']; probe($f['size']);}`, true},
		{"server field after another write", `foreach($_FILES as $f){$f['seen']=1; probe($f['size']);}`, true},
		{"copied entry", `$f=$_FILES['doc']; probe($f['tmp_name']);`, false},
		{"copied entry, client field", `$f=$_FILES['doc']; probe($f['type']);`, true},
		{"optional entry", `$f=$_FILES['doc'] ?? null; probe($f['error']);`, false},
		{"either of two entries", `$f=$c ? $_FILES['a'] : $_FILES['b']; probe($f['size']);`, false},
		{"entry or other request data", `$f=$c ? $_FILES['a'] : $_GET; probe($f['size']);`, true},
		{"copied table", `$all=$_FILES; probe($all['doc']['size']);`, false},
		{"entry named like a server field", `$all=$_FILES; probe($all['size']);`, true},
		{"listed fields of one entry", `foreach($_FILES['docs'] as $f){probe($f['size']);}`, true},
		{"entry unpacked by the loop", `foreach($_FILES as [$f]){probe($f['size']);}`, true},
		{"field of a field", `$f=$_FILES['doc']; $n=$f['name']; probe($n['size']);`, true},
		{"entry joined with text", `$f=$_FILES['doc'] . ''; probe($f['size']);`, true},
		{"entry wrapped in an array", `$f=[$_FILES['doc']]; probe($f['size']);`, true},
		{"entry cast to an array", `$f=(array) $_FILES['doc']; probe($f['size']);`, false},
		{"entry cast to a string", `$f=(string) $_FILES['doc']; probe($f['size']);`, true},
		{"name of one of several files", `probe($_FILES['f']['name'][0]);`, true},
		{"size of one of several files", `probe($_FILES['f']['size'][0]);`, false},
		{"name of a file posted as a[size]", `probe($_FILES['a']['name']['size']);`, true},
		{"name of one of several files of a copied entry", `$f=$_FILES['f']; probe($f['name'][0]);`, true},
		{"name of a file of a copied entry posted as a[size]", `$f=$_FILES['f']; probe($f['name']['size']);`, true},
		{"temporary path of one of several files of a copied entry", `$f=$_FILES['f']; probe($f['tmp_name'][0]);`, false},
		{"unknown field", `probe($_FILES['f']['other']);`, false},
		{"direct server field", `probe($_FILES['doc']['size']);`, false},
		{"direct client field", `probe($_FILES['doc']['name']);`, true},
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

// A callback returning numbers cleans the values of an array, not the keys
// the request chose for them.
func TestScalarCallbacksCleanValuesNotKeys(t *testing.T) {
	const ids = `$ids=array_map('intval', $_GET['ids']); `
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"joined values", ids + `probe(implode(',', $ids));`, false},
		{"joined values with request glue", ids + `probe(implode($_GET['glue'], $ids));`, true},
		{"joined keys", ids + `probe(implode(',', array_keys($ids)));`, true},
		{"whole array", ids + `probe($ids);`, true},
		{"one value", ids + `probe($ids[0]);`, false},
		{"listed value", ids + `foreach($ids as $k=>$v){probe($v);}`, false},
		{"listed key", ids + `foreach($ids as $k=>$v){probe($k);}`, true},
		{"fewer values", ids + `probe(implode(',', array_unique($ids)));`, false},
		{"fewer values, keys", ids + `probe(implode(',', array_keys(array_filter($ids))));`, true},
		{"imported as variables", ids + `extract($ids); probe($page);`, false},
		{"merged with numbers", ids + `probe(implode(',', array_merge($ids, [0], array_map('intval', $_POST['more']))));`, false},
		{"merged with text", ids + `probe(implode(',', array_merge($ids, $_POST['more'])));`, true},
		{"merged, keys", ids + `probe(implode(',', array_keys(array_replace($ids, [0]))));`, true},
		{"without some values", ids + `probe(implode(',', array_diff($ids, $_POST['skip'])));`, false},
		{"numbers of a list", `$ids=array_map('intval', explode(',', $_GET['ids'])); probe($ids);`, false},
		{"numbers of a list kept in a variable", `$parts=explode(',', $_GET['ids']); $ids=array_map('intval', $parts); probe($ids);`, false},
		{"numbers of a list built by the project", `$ids=array_map('intval', explode_all($_GET['ids'])); probe($ids);`, false},
		{"numbers of a superglobal element", `$ids=array_map('intval', ($_GET['ids'])); probe($ids);`, true},
		{"chunks", ids + `foreach(array_chunk($ids, 2) as $chunk){probe(implode(',', array_keys($chunk)));}`, false},
		{"chunks under their keys", ids + `foreach(array_chunk($ids, 2, true) as $chunk){probe(implode(',', array_keys($chunk)));}`, true},
		{"sorted in place", ids + `sort($ids); probe(implode(',', $ids));`, false},
		{"request data pushed", ids + `array_push($ids, $_GET['x']); probe(implode(',', $ids));`, true},
		{"walked by a callback", ids + `array_walk($ids, $f); probe(implode(',', $ids));`, true},
		{"passed to a function of the project", ids + `fill($ids); probe(implode(',', $ids));`, true},
		{"array cast", `$ids=(array) array_map('intval', $_GET['ids']); probe(implode(',', $ids));`, false},
		{"another operator", `$ids=@array_map('intval', $_GET['ids']); probe(implode(',', $ids));`, true},
		{"default", `$ids=array_map('intval', $_GET['ids']) ?? []; probe(implode(',', $ids));`, false},
		{"union with another array", `$ids=array_map('intval', $_GET['ids']) + [0]; probe(implode(',', $ids));`, true},
		{"either of two", `$ids=$c ? array_map('intval', $_GET['a']) : array_map('floatval', $_GET['b']); probe(implode(',', $ids));`, false},
		{"either numbers or text", `$ids=$c ? array_map('intval', $_GET['a']) : $_GET['b']; probe(implode(',', $ids));`, true},
		{"wrapped in an array", `$ids=[array_map('intval', $_GET['ids'])]; probe(implode(',', $ids));`, true},
		{"several arrays", `$ids=array_map('intval', $_GET['a'], $_GET['b']); probe($ids);`, false},
		{"clean array", `$ids=array_map('intval', [1, 2]); probe($ids);`, false},
		{"text callback", `$ids=array_map('trim', $_GET['ids']); probe(implode(',', $ids));`, true},
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

// A function of the project hands on the numbers it made.
func TestScalarArraysCrossFunctions(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"numbers of an array built by the project", `function pick(){return $_GET['ids'];} $ids=array_map('intval', pick()); probe(implode(',', $ids));`, false},
		{"returned by a function", `function ids(){return array_map('intval', $_GET['ids']);} probe(implode(',', ids()));`, false},
		{"returned by a function, keys", `function ids(){return array_map('intval', $_GET['ids']);} probe(implode(',', array_keys(ids())));`, true},
		{"returned by a method", `class R{static function ids(){return array_map('intval', $_GET['ids']);}} probe(implode(',', R::ids()));`, false},
		{"returned by a function as numbers or text", `function ids($c){if($c){return $_GET['raw'];} return array_map('intval', $_GET['ids']);} probe(implode(',', ids(1)));`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			e = New(f, e.types, NewSnapshot(Extract(f, e.types.Index, e.types.PHP, nil)))
			x := argument(callsNamed(f, "probe")[0])
			if got := e.Tainted(x, HTML); got != tc.taint {
				t.Fatalf("taint=%v value=%+v", got, e.Value(x))
			}
		})
	}
}

// What a PHP function leaves in a reference argument depends on the
// parameter: a number or a boolean is no text, a callback is no data.
func TestReferenceArgumentsFollowTheirParameter(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"replacement count", `$s=str_replace('a', 'b', $_GET['s'], $count); probe($count);`, false},
		{"similarity", `similar_text($_GET['a'], 'b', $pct); probe($pct);`, false},
		{"pattern replacement count", `$r=preg_replace('/a/', 'b', $_GET['s'], -1, $cnt); probe($cnt);`, false},
		{"command status", `exec($_GET['c'], $out, $code); probe($code);`, false},
		{"command output", `exec($_GET['c'], $out, $code); probe($out);`, false},
		{"scanned numbers", `sscanf($_GET['a'], '%d %5d%%', $x, $y); probe($y);`, false},
		{"scanned number and text", `sscanf($_GET['a'], '%d %s', $x, $y); probe($x);`, true},
		{"scanned with an unfinished conversion", `sscanf($_GET['a'], '%d %', $x); probe($x);`, true},
		{"scanned with a format of the request", `sscanf($_GET['a'], $_GET['f'], $x); probe($x);`, true},
		{"scanned without a format", `sscanf($_GET['a']); $x=1; probe($x);`, false},
		{"sorted by a callback of the request", `$l=['b', 'a']; usort($l, $_GET['cmp']); probe($l);`, false},
		{"match groups", `preg_match('/(.*)/', $_GET['a'], $m); probe($m);`, true},
		{"bound to a statement", `function f(mysqli_stmt $st){$x='k'; $y=$_GET['a']; $st->bind_param('ss', $x, $y); probe($y);}`, true},
		{"bound to a statement by a function", `function f($st){$y=$_GET['a']; mysqli_stmt_bind_param($st, 's', $y); probe($y);}`, true},
		{"bound to a prepared statement", `function f(PDOStatement $st){$y=$_GET['a']; $st->bindParam(1, $y); probe($y);}`, true},
		{"clean value bound to a statement", `function f($st){$y='k'; mysqli_stmt_bind_param($st, 's', $y); probe($y);}`, false},
		{"result bound to a statement", `function f(mysqli_stmt $st){$y=$_GET['a']; $st->bind_result($y); probe($y);}`, false},
		{"column bound to a statement", `function f(PDOStatement $st){$y=$_GET['a']; $st->bindColumn(1, $y); probe($y);}`, false},
		{"value given to a variadic parameter", `$c=$_GET['x']; $m=max(1, 2, $c); probe($c);`, true},
		{"value given past the parameters", `$c=$_GET['x']; strlen('a', $c); probe($c);`, false},
		{"value given to no parameter", `$c=$_GET['x']; strlen(text: $c); probe($c);`, false},
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

// `unset`, `static` and `global` take a name out of the names bound with
// it: the storage they shared keeps its value under the other names.
func TestRebindingDetachesTheName(t *testing.T) {
	for _, tc := range []struct {
		name, src       string
		resolved, taint bool
	}{
		{"unset then written", `$x=1; $r=&$x; unset($r); $r=2; probe($x);`, true, false},
		{"static then written", `function f(){$x=1; $r=&$x; static $r; $r=2; probe($x);}`, true, false},
		{"global then written", `function f(){$x=1; $r=&$x; global $r; $r=2; probe($x);}`, true, false},
		{"unset keeps request data", `$x=$_GET['a']; $r=&$x; unset($r); probe($x);`, true, true},
		{"unset then given request data", `$x='ls'; $r=&$x; unset($r); $r=$_GET['a']; probe($x);`, true, false},
		{"unset name itself", `$x=$_GET['a']; $r=&$x; unset($r); probe($r);`, false, false},
		{"unset element of a bound name", `$x=[1]; $r=&$x; unset($r[0]); probe($x);`, false, false},
		{"unset on one path", `$x=1; $r=&$x; if($c){unset($r);} $r=2; probe($x);`, false, false},
		{"unset of a kept reference", `$x=1; $o->p=&$x; unset($x); $o->p=5; $x=2; probe($x);`, true, false},
		{"bound on one path to a reference parameter", `function f(&$x, $c){$x=1; if($c){$r=&$x;} $r=2; probe($x);}`, false, false},
		{"bound on the other path to a reference parameter", `function f(&$x, $c){$x=1; if($c){echo 1;}else{$r=&$x;} $r=2; probe($x);}`, false, false},
		{"bound on one path to a bound name", `$x=1; $y=&$x; if($c){$r=&$x;} $r=2; probe($x);`, false, false},
		{"bound on the other path to a bound name", `$x=1; $y=&$x; if($c){echo 1;}else{$r=&$x;} $r=2; probe($x);`, false, false},
		{"reference parameter bound on both paths", `function f(&$x, $c){$r=&$x; $x=1; if($c){echo 1;} $r=2; probe($x);}`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			x := argument(callsNamed(f, "probe")[0])
			if _, got := e.Resolve(x); got != tc.resolved {
				t.Errorf("resolved=%v value=%+v", got, e.Value(x))
			}
			if got := e.Tainted(x, HTML); got != tc.taint {
				t.Errorf("taint=%v value=%+v", got, e.Value(x))
			}
		})
	}
}

// Request data written to one name is read under the names bound with it.
func TestRequestDataReachesBoundNames(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		taint     bool
	}{
		{"written through a dynamic name", `$x='ls'; $$n=$_GET['a']; probe($x);`, true},
		{"written through a reference to a dynamic name", `$r=&$$n; $x='ls'; $r=$_GET['a']; probe($x);`, true},
		{"assigned in a loop target, no element", `$x=$_GET['a']; foreach($l as $o->p[$x='ls']){} probe($x);`, true},
		{"assigned inside a written offset", `$o->seen[$x=$_GET['a']]=true; probe($x);`, true},
		{"overwritten inside a written offset", `$x=$_GET['a']; $o->seen[$x='ls']=true; probe($x);`, false},
		{"assigned inside a dynamic property name", `$y=$o->{$x=$_GET['a']}; probe($x);`, true},
		{"loop variable", `$x='ls'; $r=&$x; foreach($_GET as $r){} probe($x);`, true},
		{"loop key", `$x='ls'; $r=&$x; foreach($_GET as $r=>$v){} probe($x);`, true},
		{"destructuring", `$x='ls'; $r=&$x; [$r]=[$_GET['a']]; probe($x);`, true},
		{"holder of the other name", `$x='ls'; $f=function() use (&$x){$x=$_GET['a'];}; $r=&$x; $f(); probe($r);`, true},
		{"holder writing a constant", `$x='ls'; $f=function() use (&$x){$x='pwd';}; $r=&$x; $f(); probe($r);`, false},
		{"holder of an unrelated name", `$x='ls'; $y='ls'; $f=function() use (&$y){$y=$_GET['a'];}; $r=&$x; $f(); probe($r);`, false},
		{"import named arguments", `$c='ls'; extract(flags: EXTR_OVERWRITE, array: $_GET); probe($c);`, true},
		{"import named arguments, existing kept", `$c='ls'; extract(flags: EXTR_SKIP, array: $_GET); probe($c);`, false},
		{"import named flags", `$c='ls'; extract($_GET, flags: EXTR_PREFIX_ALL, prefix: 'in'); probe($c);`, false},
		{"import keeping existing by reference", `$c='ls'; extract($_GET, EXTR_SKIP | EXTR_REFS); probe($c);`, false},
		{"import by reference", `$c='ls'; extract($_GET, EXTR_REFS | EXTR_OVERWRITE); probe($c);`, true},
		{"import by reference creates variables", `extract($_GET, EXTR_REFS | EXTR_SKIP); probe($c);`, true},
		{"import of a named string", `$c='ls'; parse_str(string: $_SERVER['QUERY_STRING']); probe($c);`, true},
		{"import of nothing", `$c='ls'; extract(); probe($c);`, false},
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

// A handle stored in an array is no longer followed, until an operation on
// its variable says what state it is in.
func TestHandleStoredThenOperatedOnIsTracked(t *testing.T) {
	e, f := environment(t, `$h=fopen('a','r'); $list[]=$h; fclose($h); fread($h,1);`, nil)
	c := callsNamed(f, "fread")[0]
	if s, ok := e.StateBefore(c, argument(c)); !ok || s.Operation != "fclose" {
		t.Fatalf("state=%+v/%v", s, ok)
	}
}

func TestNumericFormat(t *testing.T) {
	for format, want := range map[string]bool{"": true, "v": true, "%d": true, "%d.%d": true, "%*d %3x%%": true, "%s": false, "%d %[a-z]": false, "%": false, "%5": false} {
		if got := numericFormat(format); got != want {
			t.Errorf("numericFormat(%q)=%v", format, got)
		}
	}
}
