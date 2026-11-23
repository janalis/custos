package flow

import (
	"testing"

	"github.com/janalis/custos/internal/php/syntax"
)

// TestPartialApplicationBindsReferences: a partial application keeps an
// argument bound to a reference parameter by reference, so each call of the
// closure may write the variable; one bound by value stays as it was.
func TestPartialApplicationBindsReferences(t *testing.T) {
	for src, kept := range map[string]bool{
		`$m='x';$c=preg_match('/a/',?,$m);$c('ab');probe($m);`:                         false,
		`$m='x';$c=unknown_fn(?,$m);$c('ab');probe($m);`:                               false,
		`$m='x';$c=preg_match(?,subject:'ab',matches:$m);$c('/a/');probe($m);`:         false,
		`$m='x';$c=preg_match(pattern:?,subject:'ab',matches:$m);$c('/a/');probe($m);`: false,
		`$m='x';$c=str_pad(string:?,length:4,pad_string:$m);$c('ab');probe($m);`:       true,
		`$m='x';$c=str_pad(?,4,$m);$c('ab');probe($m);`:                                true,
	} {
		e, f := environment(t, src, nil)
		value, known := e.Resolve(argument(callsNamed(f, "probe")[0]))
		_, literal := value.(*syntax.Literal)
		if got := known && literal; got != kept {
			t.Errorf("%s: literal kept = %v (%T)", src, got, value)
		}
	}
}
