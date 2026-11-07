package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestBorrowedResourceEffects(t *testing.T) {
	for _, tc := range []struct {
		definition, invoke, operation string
		known                         bool
	}{
		{`function useIt($h){fread($h,1);}`, `useIt($h);`, "fread", true},
		{`function useIt($h){fread($h,1);return tmpfile();}`, `$other=useIt($h);`, "fread", true},
		{`function useIt($h){fread($h,1);}`, `$h=unknown();useIt($h);`, "fread", false},
		{`function useIt($h,&$other){fread($h,1);}`, `$other='x';useIt($h,$other);`, "fread", false},
		{`function useIt($h,...$other){fread($h,1);}`, `useIt($h);`, "fread", false},
		{`final class R{static function useIt($h){fread($h,1);}}`, `$cb=R::useIt(...);$cb($h);`, "fread", true},
		{`function useIt($h){fwrite($h,'x');}`, `useIt($h);`, "fwrite", true},
		{`function useIt($h){if(!fclose($h)){exit;}}`, `useIt($h);`, "fclose", true},
		{`function useIt($h){fclose($h);}`, `useIt($h);`, "fclose", false},
		{`function useIt($h){if(rand()){fread($h,1);}}`, `useIt($h);`, "fread", false},
		{`function useIt($h){fread($h,1);opaque($h);}`, `useIt($h);`, "fread", false},
		{`function useIt($h){fread($h,1);yield 1;}`, `useIt($h);`, "fread", false},
		{`function useIt($a,$b){fread($a,1);fwrite($b,'x');}`, `useIt($h,$h);`, "fread", false},
		{`final class R{static function useIt($h){fread($h,1);}}`, `R::useIt($h);`, "fread", true},
		{`final class R{function useIt($h){fread($h,1);}}`, `$worker=new R;$worker->useIt($h);`, "fread", true},
		{`class R{function __construct($h){fread($h,1);}}`, `$worker=new R($h);`, "fread", false},
	} {
		e, file := summarized(t, tc.definition+`$h=tmpfile();`+tc.invoke+`probe($h);`)
		at := callsNamed(file, "probe")[0]
		if state, known := e.StateBefore(at, argument(at), tc.operation); known != tc.known {
			t.Fatalf("%s %s: known=%v state=%+v", tc.definition, tc.invoke, known, state)
		}
	}
}

func TestCrossFileBorrowedResourceEffects(t *testing.T) {
	definitions := `<?php function useIt($h){if(!fclose($h)){exit;}}`
	callerSource := `<?php function caller(){$h=tmpfile();useIt($h);probe($h);}`
	ix := stubsIndexForResource(t, definitions, callerSource)
	defs := syntax.Parse("definitions.php", []byte(definitions), syntax.Options{Version: phpversion.Max})
	snapshot := NewSnapshot(Extract(defs, ix, phpversion.Max, nil, Options{}))
	caller := syntax.Parse("caller.php", []byte(callerSource), syntax.Options{Version: phpversion.Max})
	e, out := Summarize(caller, resourceTypes(caller, ix), snapshot, Options{})
	at := callsNamed(caller, "probe")[0]
	if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
		t.Fatal("cross-file checked close lost")
	}
	replacement := syntax.Parse("definitions.php", []byte(`<?php function useIt($h){fclose($h);}`), syntax.Options{Version: phpversion.Max})
	next := snapshot.WithFile(Extract(replacement, ix, phpversion.Max, nil, Options{}))
	if out.Current(snapshot, next) {
		t.Fatal("resource effect edit did not invalidate caller")
	}
	revised, _ := Summarize(caller, resourceTypes(caller, ix), next, Options{})
	if _, closed := revised.StateBefore(at, argument(at), "fclose"); closed {
		t.Fatal("unchecked replacement retained close")
	}
}

func TestBorrowedEffectsRequireClosedNativeDispatch(t *testing.T) {
	for _, source := range []string{
		`class R{function useIt($h){fread($h,1);}}function caller(R $worker){$h=tmpfile();$worker->useIt($h);probe($h);}`,
		`final class R{function useIt($h){fread($h,1);}}function caller($worker){/** @var R $worker */ $h=tmpfile();$worker->useIt($h);probe($h);}`,
		`class R{static function useIt($h){fread($h,1);}function caller(){$h=tmpfile();static::useIt($h);probe($h);}}`,
	} {
		e, file := summarized(t, source)
		at := callsNamed(file, "probe")[0]
		if _, known := e.StateBefore(at, argument(at), "fread"); known {
			t.Fatalf("unproven dispatch projected an effect: %s", source)
		}
	}
}
