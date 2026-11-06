package flow

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestLocalContractsAreDemandDriven(t *testing.T) {
	e, file := environment(t, `function value(){return 'safe';}function wrap(){return value();}function unused(){opaque();}probe(wrap());`, nil)
	local := NewLocal(file, e.types, nil, Options{}, nil)
	if len(local.scopes) != 0 || len(local.local.cache) != 0 {
		t.Fatal("inventory followed callable bodies")
	}
	value := local.Value(argument(callsNamed(file, "probe")[0]))
	scalar, known := knownScalar(value)
	if !known || scalar.s != "safe" {
		t.Fatalf("missing demanded contract: %+v", value)
	}
	if _, computed := local.local.cache["unused"]; computed || local.scopes[functionNamed(file, "unused")] != nil {
		t.Fatal("unrelated callable was followed")
	}
	if _, known, own := local.localContract("wrap"); !known || !own {
		t.Fatal("cached contract disappeared")
	}
	_, all := Summarize(file, e.types, nil, Options{})
	if len(all.Summaries) != 3 {
		t.Fatal("project publication became partial")
	}
}

func TestLocalRecursionIsQueryOrderIndependent(t *testing.T) {
	for _, source := range []string{
		`function a($x){if($x){return b($x);}return $_GET['x'];}function b($x){return a($x);}`,
		`function a(){return b();}function b(){return c();}function c(){return a();}`,
		`function a($x){return b($x);}function b($x){return $x;}function c($x){return b($x);}`,
	} {
		e, file := environment(t, source, nil)
		first := NewLocal(file, e.types, nil, Options{}, nil)
		second := NewLocal(file, e.types, nil, Options{}, nil)
		first.contract("a")
		first.contract("b")
		second.contract("b")
		second.contract("a")
		if !reflect.DeepEqual(first.local.cache, second.local.cache) {
			t.Fatalf("query order changed contracts: %s", source)
		}
	}
}

func TestLocalOverlayShadowsAndRemoves(t *testing.T) {
	e, old := environment(t, `function source(){return $_GET['x'];}function removed(){return $_GET['x'];}`, nil)
	_, published := Summarize(old, e.types, nil, Options{})
	snapshot := NewSnapshot(published)
	fresh, file := environment(t, `function source(){return 'safe';}probe(source());`, snapshot)
	local := NewLocal(file, fresh.types, snapshot, Options{}, nil)
	if local.Tainted(argument(callsNamed(file, "probe")[0]), HTML) {
		t.Fatal("old buffer contract remained visible")
	}
	if _, known := local.contract("removed"); known {
		t.Fatal("removed declaration remained visible")
	}
	if sum, known := snapshot.lookup("source"); !known || !tainted(sum.Return) {
		t.Fatal("local query mutated published snapshot")
	}
	// Flattening an incoming overlay still preserves shadowing and ambiguity.
	overlay := snapshot.WithFile(&File{Path: "other.php", Summaries: map[string]Summary{"source": {}}})
	ambiguous := NewLocal(file, fresh.types, overlay, Options{}, nil)
	if _, known, own := ambiguous.localContract("source"); known || own {
		t.Fatal("ambiguous declaration acquired a local contract")
	}
	duplicate, dups := environment(t, `if($x){function f(){return 1;}}else{function f(){return 2;}}`, nil)
	if _, known, own := NewLocal(dups, duplicate.types, nil, Options{}, nil).localContract("f"); known || own {
		t.Fatal("duplicate declaration acquired a local contract")
	}
}

func TestLocalContractsRebaseExternalDependencies(t *testing.T) {
	defs := `<?php function source(){return $_GET['x'];}`
	caller := `<?php function wrap(){return source();}probe(wrap());`
	ix := stubsIndexForResource(t, defs, caller)
	file := syntax.Parse("caller.php", []byte(caller), syntax.Options{Version: phpversion.Max})
	makeSnapshot := func(source string) *Snapshot {
		file := syntax.Parse("definitions.php", []byte(source), syntax.Options{Version: phpversion.Max})
		return NewSnapshot(Extract(file, ix, phpversion.Max, nil, Options{}))
	}
	local := NewLocal(file, resourceTypes(file, ix), makeSnapshot(defs), Options{}, nil)
	at := argument(callsNamed(file, "probe")[0])
	if !local.Tainted(at, HTML) {
		t.Fatal("external source lost")
	}
	rebased := local.Rebase(makeSnapshot(`<?php function source(){return 'safe';}`))
	if rebased == local || rebased.Tainted(at, HTML) {
		t.Fatal("changed external dependency retained stale facts")
	}
	if !local.Tainted(at, HTML) {
		t.Fatal("rebase changed original environment")
	}
}

func TestLocalDiscoveryBudget(t *testing.T) {
	e, file := environment(t, `function f(){return 1;}`, nil)
	local := NewLocal(file, e.types, nil, Options{}, nil)
	if got := local.materializeLocal("f", 0); got.Fixed || got.Return.Complete {
		t.Fatal("exhausted discovery established a contract")
	}
	if gaps := local.Gaps(); len(gaps) != 1 || gaps[0].Reason != "flow-budget" {
		t.Fatalf("missing discovery status: %+v", gaps)
	}
}

func BenchmarkLocalSummaryDemand(b *testing.B) {
	var source strings.Builder
	source.WriteString(`function needed($v){return $v;}probe(needed('safe'));`)
	for i := range 64 {
		fmt.Fprintf(&source, `function unused%d($v){if($v){return $v+1;}return $v-1;}`, i)
	}
	e, file := environment(b, source.String(), nil)
	at := argument(callsNamed(file, "probe")[0])
	for _, lazy := range []bool{false, true} {
		b.Run(fmt.Sprint(lazy), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var env *Env
				if lazy {
					env = NewLocal(file, e.types, nil, Options{}, nil)
				} else {
					var own *File
					env, own = Summarize(file, e.types, nil, Options{})
					env = env.Rebase((*Snapshot)(nil).WithFile(own))
				}
				env.Value(at)
			}
		})
	}
}

func TestLocalWorkersReuseOnlyFinishedScopes(t *testing.T) {
	e, file := environment(t, `function dep($v){return $v;}function wrap($v){return dep($v);}function other($v){return strlen($v);}`, nil)
	local := NewLocal(file, e.types, nil, Options{}, nil)
	wrap := functionNamed(file, "wrap")
	local.scope(wrap)
	prior := local.scopes[wrap]
	local.contract("wrap")
	if !prior.finished || local.scopes[wrap] != prior {
		t.Fatal("finished query scope was recomputed")
	}
	other := functionNamed(file, "other")
	local.scope(other)
	local.contract("other")
	if !local.globalsReady {
		t.Fatal("global inventory did not finish")
	}
}

func TestLocalWorkersShareRecordIdentities(t *testing.T) {
	e, file := environment(t, `function make(){return ['h'=>tmpfile()];}function wrap(){return make();}function other(){return make();}`, nil)
	local := NewLocal(file, e.types, nil, Options{}, nil)
	local.contract("wrap")
	local.contract("other")
	ids := map[uint32]bool{}
	for _, id := range local.recordSites {
		if ids[id] {
			t.Fatal("independent worker allocations share an identity")
		}
		ids[id] = true
	}
	if len(ids) < 2 {
		t.Fatal("worker allocation sites were not retained")
	}
}

func TestLocalDeferredGuardAndRebase(t *testing.T) {
	e, file := environment(t, `function f(){return 'safe';}`, nil)
	builds := 0
	guard := func(build func()) { builds++; build() }
	local := NewLocal(file, e.types, nil, Options{}, guard)
	first, known := local.contract("f")
	if !known || !first.Fixed || builds != 1 {
		t.Fatal("deferred guard did not run")
	}
	local.contract("f")
	if builds != 1 {
		t.Fatal("cached contract was rebuilt")
	}
	rebased := local.Rebase(NewSnapshot())
	rebased.contract("f")
	if builds != 2 {
		t.Fatal("rebase dropped the deferred build guard")
	}
}
