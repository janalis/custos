package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestNamedGuardArgumentProof(t *testing.T) {
	for _, tc := range []struct {
		src     string
		context Context
		safe    bool
		length  bool
	}{
		{`$x=$_GET['x'];if(preg_match(subject:$x,pattern:'/[\r\n]/')!==0){return;}probe($x);`, Header, true, false},
		{`$x=$_GET['x'];if(preg_match(subject:$x,pattern:'/[\r\n]/',matches:$matches)!==0){return;}probe($x);`, Header, true, false},
		{`$x=$_GET['x'];if(!in_array(strict:true,haystack:['/home','/help'],needle:$x)){return;}probe($x);`, URL, true, false},
		{`$x=$_GET['x'];if(str_contains(needle:"\r",haystack:$x)||str_contains(needle:"\n",haystack:$x)){return;}probe($x);`, Header, true, false},
		{`$x=$_GET['x'];if(!ctype_digit(text:$x)){return;}probe($x);`, SQL, true, false},
		{`$x=$_GET['x'];if(!is_int(value:$x)){return;}probe($x);`, SQL, true, false},
		{`$x=$_GET['x'];if(!is_integer(value:$x)){return;}probe($x);`, SQL, true, false},
		{`$x=fread($h,4);if(strlen(string:$x)!==4){return;}probe($x);`, 0, false, true},
		{`$x=$_GET['x'];if(!in_array(needle:$x,haystack:['/home'])){return;}probe($x);`, URL, false, false},
		{`$x=$_GET['x'];if(!is_int(wrong:$x)){return;}probe($x);`, SQL, false, false},
		{`$x=$_GET['x'];if(preg_match(pattern:'/[\r\n]/',subject:$x,subject:$x)!==0){return;}probe($x);`, Header, false, false},
		{`$x=$_GET['x'];if(!ctype_digit(...$args)){return;}probe($x);`, SQL, false, false},
		{`$x=$_GET['x'];if(!is_int(&$x)){return;}probe($x);`, SQL, false, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			x := argument(callsNamed(f, "probe")[0])
			v := e.Value(x)
			if tc.context != 0 && (v.Safe&tc.context != 0) != tc.safe {
				t.Fatalf("value %+v, want safe %v", v, tc.safe)
			}
			if tc.length && (!v.LengthKnown || v.Length != 4) {
				t.Fatalf("missing exact length: %+v", v)
			}
		})
	}
}

func TestNamedFlockGuard(t *testing.T) {
	for _, src := range []string{
		`$h=fopen('f','c+');if(!flock(operation:LOCK_EX,stream:$h)){return;}fwrite($h,'x');`,
		`$h=fopen('f','c+');if(!flock(operation:LOCK_EX,stream:$h,would_block:$blocked)){return;}fwrite($h,'x');`,
	} {
		t.Run(src, func(t *testing.T) {
			e, f := environment(t, src, nil)
			call := callsNamed(f, "fwrite")[0]
			state, known := e.StateBefore(call, argument(call), "flock")
			if !known || !state.Successful {
				t.Fatalf("lock proof %+v known %v", state, known)
			}
		})
	}
}

func TestGuardBindingValidation(t *testing.T) {
	for _, tc := range []struct {
		src, name string
		valid     bool
		first     string
	}{
		{`preg_match(subject:$subject,pattern:$pattern,matches:$matches);`, "preg_match", true, "$pattern"},
		{`strlen(string:$value);`, "strlen", true, "$value"},
		{`unknown($value);`, "unknown", false, ""},
		{`strlen();`, "strlen", false, ""},
		{`strlen(wrong:$value);`, "strlen", false, ""},
		{`strlen(string:$value,string:$value);`, "strlen", false, ""},
		{`strlen($value,$extra);`, "strlen", false, ""},
		{`strlen(...$args);`, "strlen", false, ""},
		{`strlen(&$value);`, "strlen", false, ""},
		{`strlen(...);`, "strlen", false, ""},
		{`strlen(string:$value,$extra);`, "strlen", false, ""},
		{`printf('%s',$value);`, "printf", false, ""},
		{`time();`, "time", true, ""},
	} {
		t.Run(tc.src, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			c := callsNamed(f, tc.name)[0]
			args, valid := e.guardArguments(c)
			if valid != tc.valid {
				t.Fatalf("valid %v want %v", valid, tc.valid)
			}
			if valid && tc.first != "" {
				expr := args[0].Value
				if string(f.Src[expr.Span().Start:expr.Span().End]) != tc.first {
					t.Fatal("did not preserve declaration ordering")
				}
				found := false
				for _, raw := range c.Args.Args {
					if raw.(*syntax.Arg) == args[0] {
						found = true
					}
				}
				if !found {
					t.Fatal("argument identity changed")
				}
			}
		})
	}
}

func TestRequestFieldTrustBoundaries(t *testing.T) {
	for _, tc := range []struct {
		expr    string
		tainted bool
	}{
		{`$_SERVER['SERVER_SOFTWARE']`, false},
		{`$_SERVER['HTTP_HOST']`, true},
		{`$_FILES['upload']['tmp_name']`, false},
		{`$_FILES['upload']['error']`, false},
		{`$_FILES['upload']['size']`, false},
		{`$_FILES['upload']['name']`, true},
		{`$_FILES['upload']['type']`, true},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			e, f := environment(t, "probe("+tc.expr+");", nil)
			value := argument(callsNamed(f, "probe")[0])
			fact := e.Value(value)
			if !tc.tainted && (!fact.Complete || len(fact.Sources) != 0) {
				t.Fatalf("trusted system field lacks complete source-free proof: %+v", fact)
			}
			if got := e.Tainted(value, Path); got != tc.tainted {
				t.Fatalf("tainted %v want %v: %+v", got, tc.tainted, e.Value(value))
			}
		})
	}
}

func TestGuardNamedArgumentMinimumVersion(t *testing.T) {
	e, f := environment(t, `strlen(string:$value);`, nil)
	e.types.PHP = phpversion.PHP74
	if _, valid := e.guardArguments(callsNamed(f, "strlen")[0]); valid {
		t.Fatal("named binding accepted before PHP8")
	}
}

func BenchmarkNamedGuardBinding(b *testing.B) {
	e, f := environment(b, `preg_match(subject:$subject,pattern:'/x/',matches:$matches);`, nil)
	call := callsNamed(f, "preg_match")[0]
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.guardArguments(call)
	}
}
