package flow

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestForwardContractsRespectDefinitionStrictness(t *testing.T) {
	for _, tc := range []struct {
		defs, caller string
		empty        bool
	}{
		{`<?php declare(strict_types=1);function zero(int $v){return $v;}`, `<?php $xs=array_filter(['00'],'zero');probe($xs);`, true},
		{`<?php function zero(int $v):string{return $v;}`, `<?php declare(strict_types=1);$xs=array_filter(['00'],'zero');probe($xs);`, true},
		{`<?php declare(strict_types=1);function zero(int $v):string{return $v;}`, `<?php $xs=array_filter(['00'],'zero');probe($xs);`, false},
		{`<?php function factory(){return fn(int $v):string=>$v;}`, `<?php declare(strict_types=1);$xs=array_filter(['00'],factory());probe($xs);`, true},
		{`<?php declare(strict_types=1);function factory(){return fn(int $v):string=>$v;}`, `<?php $xs=array_filter(['00'],factory());probe($xs);`, false},
	} {
		ix := stubsIndexForResource(t, tc.defs, tc.caller)
		defs := syntax.Parse("definitions.php", []byte(tc.defs), syntax.Options{Version: phpversion.Max})
		snap := NewSnapshot(Extract(defs, ix, phpversion.Max, nil, Options{}))
		file := syntax.Parse("caller.php", []byte(tc.caller), syntax.Options{Version: phpversion.Max})
		e := New(file, resourceTypes(file, ix), snap, Options{})
		v := e.Value(argument(callsNamed(file, "probe")[0]))
		empty := v.elements != nil && v.elements.whole && !v.elements.optional && len(v.elements.items) == 0
		if empty != tc.empty {
			t.Fatalf("%s / %s: empty=%v", tc.defs, tc.caller, empty)
		}
	}
}

func TestForwardContractEditInvalidatesCallers(t *testing.T) {
	defs := `<?php function zero(int $v){return $v;}`
	caller := `<?php function caller(){$xs=array_filter(['00'],'zero');probe($xs);}`
	ix := stubsIndexForResource(t, defs, caller)
	file := syntax.Parse("definitions.php", []byte(defs), syntax.Options{Version: phpversion.Max})
	first := NewSnapshot(Extract(file, ix, phpversion.Max, nil, Options{}))
	use := syntax.Parse("caller.php", []byte(caller), syntax.Options{Version: phpversion.Max})
	_, published := Summarize(use, resourceTypes(use, ix), first, Options{})
	edited := syntax.Parse("definitions.php", []byte(`<?php function zero(int $v){return $v+1;}`), syntax.Options{Version: phpversion.Max})
	next := first.WithFile(Extract(edited, ix, phpversion.Max, nil, Options{}))
	if published.Current(first, next) {
		t.Fatal("forward contract edit did not invalidate dependent")
	}
	e := New(use, resourceTypes(use, ix), next, Options{})
	v := e.Value(argument(callsNamed(use, "probe")[0]))
	if v.elements != nil && !v.elements.optional && len(v.elements.items) == 0 {
		t.Fatal("old forwarding facts survived the edit")
	}
}

func TestNonDeclarationHasNoReturnCoercion(t *testing.T) {
	if returnDeclaration(nil) != nil {
		t.Fatal("non-declaration acquired a return type")
	}
}

func TestResourceRecordContractEdits(t *testing.T) {
	for _, tc := range []struct{ defs, edited, caller string }{
		{`<?php function pass($r){return $r;}`, `<?php function pass($r){opaque($r);return $r;}`, `<?php function caller(){$r=pass(['h'=>tmpfile()]);fclose($r['h']);probe($r['h']);}`},
		{`<?php function make(){return fn()=>['h'=>tmpfile()];}`, `<?php function make(){return fn()=>['h'=>'safe'];}`, `<?php function caller(){$cb=make();$r=$cb();fclose($r['h']);probe($r['h']);}`},
	} {
		ix := stubsIndexForResource(t, tc.defs, tc.caller)
		defs := syntax.Parse("definitions.php", []byte(tc.defs), syntax.Options{Version: phpversion.Max})
		published := Extract(defs, ix, phpversion.Max, nil, Options{})
		var check func(Value)
		check = func(v Value) {
			if v.Expr != nil || v.Identity != 0 || v.storedIdentity != 0 || v.origins != nil || v.targets != nil || v.returned != nil {
				t.Fatal("execution metadata escaped in a contract")
			}
			if v.elements != nil {
				for _, entry := range v.elements.items {
					check(entry)
				}
			}
		}
		for _, s := range published.Summaries {
			check(s.Return)
			if s.returned != nil {
				check(s.returned.result)
				for _, v := range s.returned.defaults {
					check(v)
				}
			}
		}
		snap := NewSnapshot(published)
		use := syntax.Parse("caller.php", []byte(tc.caller), syntax.Options{Version: phpversion.Max})
		e, out := Summarize(use, resourceTypes(use, ix), snap, Options{})
		at := callsNamed(use, "probe")[0]
		if _, closed := e.StateBefore(at, argument(at), "fclose"); !closed {
			t.Fatal("cross-file resource contract lost")
		}
		edited := syntax.Parse("definitions.php", []byte(tc.edited), syntax.Options{Version: phpversion.Max})
		next := snap.WithFile(Extract(edited, ix, phpversion.Max, nil, Options{}))
		if out.Current(snap, next) {
			t.Fatal("resource contract edit did not invalidate caller")
		}
		updated := New(use, resourceTypes(use, ix), next, Options{})
		if _, closed := updated.StateBefore(at, argument(at), "fclose"); closed {
			t.Fatal("old resource facts survived contract edit")
		}
	}
}
