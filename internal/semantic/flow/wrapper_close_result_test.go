package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestWrapperCloseResultBranches(t *testing.T) {
	for _, tc := range []struct {
		definition, body string
		closed           bool
	}{
		{`function shut($h){return fclose($h);}`, `if(shut($h)){probe($h);}`, true},
		{`function shut($h):bool{return fclose($h);}`, `if(shut($h)!==false){probe($h);}`, true},
		{`function shut($h):int{return fclose($h);}`, `if(shut($h)!==false){probe($h);}`, false},
		{`function shut($h):string{return fclose($h);}`, `if(shut($h)!==false){probe($h);}`, false},
		{`function shut($h){return fclose($h);}function wrap(string $h){return shut($h);}`, `if(wrap($h)){probe($h);}`, false},
		{`function shut($h){return fclose($h);}function wrap(mixed $h){return shut($h);}`, `if(wrap($h)){probe($h);}`, true},
		{`function shut($h){return fclose($h);}`, `if(!shut($h)){probe($h);}`, false},
		{`function shut($h){return fclose($h);}`, `$ok=shut($h);if($ok===true){probe($h);}`, true},
		{`function shut($h){return fclose($h);}`, `if($ok=shut($h)){probe($h);}`, true},
		{`function shut($h){return fclose($h);}`, `shut($h);probe($h);`, false},
		{`function shut($h){return fclose($h);}`, `$ok=shut($h);$ok=true;if($ok){probe($h);}`, false},
		{`function shut($h){return fclose($h);}`, `$ok=shut($h);fread($h,1);if($ok){probe($h);}`, false},
		{`function shut($h){return fclose($h);}`, `$ok=shut($h);opaque($h);if($ok){probe($h);}`, false},
		{`function shut($h){return !fclose($h);}`, `if(shut($h)){probe($h);}`, false},
		{`function shut($h){return pclose($h);}`, `if(shut($h)){probe($h);}`, false},
		{`function shut($h){return fclose($h);}`, `$h=unknown();if(shut($h)){probe($h);}`, false},
		{`function shut($h){return fclose($h);}`, `if(($ok&=shut($h))===true){probe($h);}`, false},
		{`function shut($h){fclose($h);return true;}`, `if(shut($h)){probe($h);}`, false},
		{`function shut($h){fclose($h);return is_int(1);}`, `if(shut($h)){probe($h);}`, false},
		{`function shut($h){$ok=fclose($h);fread($h,1);return $ok;}`, `if(shut($h)){probe($h);}`, false},
		{`function shut($h,$flag){if($flag){return fclose($h);}return true;}`, `if(shut($h,$flag)){probe($h);}`, false},
		{`function shut($h){return fclose($h);}function wrap($h){return shut($h);}`, `if(wrap($h)){probe($h);}`, true},
		{`function shut($h){return fclose($h);}`, `$cb=shut(...);if($cb($h)){probe($h);}`, true},
		{`final class R{static function shut($h){return fclose($h);}}`, `if(R::shut($h)){probe($h);}`, true},
		{`final class R{function shut($h){return fclose($h);}}`, `$r=new R;if($r->shut($h)){probe($h);}`, true},
		{`class R{function shut($h){return fclose($h);}}`, `function caller(R $r){$h=tmpfile();if($r->shut($h)){probe($h);}}`, false},
	} {
		e, file := summarized(t, tc.definition+`$h=tmpfile();`+tc.body)
		at := callsNamed(file, "probe")[0]
		if state, closed := e.StateBefore(at, argument(at), "fclose"); closed != tc.closed {
			t.Fatalf("%s %s: %v %+v", tc.definition, tc.body, closed, state)
		}
	}
}

func TestCrossFileWrapperCloseResult(t *testing.T) {
	definitions := `<?php function shut($h):bool{return fclose($h);}`
	source := `<?php function caller(){$h=tmpfile();$ok=shut($h);if($ok){probe($h);}}`
	ix := stubsIndexForResource(t, definitions, source)
	defs := syntax.Parse("definitions.php", []byte(definitions), syntax.Options{Version: phpversion.Max})
	snapshot := NewSnapshot(Extract(defs, ix, phpversion.Max, nil, Options{}))
	caller := syntax.Parse("caller.php", []byte(source), syntax.Options{Version: phpversion.Max})
	e, out := Summarize(caller, resourceTypes(caller, ix), snapshot, Options{})
	at := callsNamed(caller, "probe")[0]
	if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
		t.Fatal("cross-file close/result coupling lost")
	}
	replacement := syntax.Parse("definitions.php", []byte(`<?php function shut($h):bool{fclose($h);return true;}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(replacement, ix, phpversion.Max, nil, Options{}))
	if out.Current(snapshot, next) {
		t.Fatal("close/result contract edit did not invalidate caller")
	}
	revised, _ := Summarize(caller, resourceTypes(caller, ix), next, Options{})
	if _, closed := revised.StateBefore(at, argument(at), "fclose"); closed {
		t.Fatal("unrelated return value retained close proof")
	}
}
