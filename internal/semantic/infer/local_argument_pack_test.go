package infer

import (
	"fmt"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestLocalArgumentPacks(t *testing.T) {
	for _, tc := range []struct {
		before, arguments string
		known             bool
	}{
		{`$options=[CURLOPT_SSL_VERIFYPEER,false];`, `$h,...$options`, true},
		{`$options=['option'=>CURLOPT_SSL_VERIFYPEER,'value'=>false];`, `$h,...$options`, true},
		{`$options=[CURLOPT_SSL_VERIFYPEER,false];$copy=$options;`, `$h,...$copy`, true},
		{`$value=false;$options=[CURLOPT_SSL_VERIFYPEER,$value];$value=true;`, `$h,...$options`, true},
		{`$options=[CURLOPT_SSL_VERIFYPEER,false];$options[1]=true;`, `$h,...$options`, false},
		{`$options=[CURLOPT_SSL_VERIFYPEER,false];$alias=&$options;`, `$h,...$options`, false},
		{`$options=[CURLOPT_SSL_VERIFYPEER,false];unknown($options);`, `$h,...$options`, false},
		{`if($flag){$options=[CURLOPT_SSL_VERIFYPEER,false];}`, `$h,...$options`, false},
		{`/** @var array{0:int,1:false} $options */`, `$h,...$options`, false},
		{`$options=makeOptions();`, `$h,...$options`, false},
		{``, `$h,...$this`, false},
	} {
		e, f := assertionEnv(t, `function f($h){`+tc.before+`probe(curl_setopt(`+tc.arguments+`));}`, false, phpversion.Max)
		args, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall))
		if known != tc.known {
			t.Fatalf("%s: %v", tc.before, known)
		}
		if known {
			if option, ok := args[1].(*syntax.ConstFetch); !ok || option.Name.Value != "CURLOPT_SSL_VERIFYPEER" {
				t.Fatal("wrong local option")
			}
		}
	}
	source := `$p0=[CURLOPT_SSL_VERIFYPEER,false];`
	for i := 1; i < 10; i++ {
		source += fmt.Sprintf(`$p%d=$p%d;`, i, i-1)
	}
	e, f := assertionEnv(t, `function f($h){`+source+`probe(curl_setopt($h,...$p9));}`, false, phpversion.Max)
	if _, known := e.BuiltinArguments(assertionRead(f).(*syntax.FuncCall)); known {
		t.Fatal("pack alias depth ignored")
	}
}
