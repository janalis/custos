package flow

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

// A variable read can have several producers: across a loop back-edge, or
// after an assignment made on some path only. Each keeps the failure values
// the guards on its own paths exclude.
func TestReachingProducers(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		// want lists, per producer in join order, its call name (or the node
		// kind) followed by the excluded failure values.
		want   string
		single bool
	}{
		{"single", `$u=parse_url($a); probe($u);`, "parse_url", true},
		{"single, tested", `$u=parse_url($a); if($u === false){return;} probe($u);`, "parse_url!false", true},
		{"loop", `$u=parse_url($a); if($u === false){return;} while($c--){probe($u); $u=scandir($n);}`, "parse_url!false scandir", false},
		{"loop, tested inside", `$u=parse_url($a); while($c--){if(!$u){return;} probe($u); $u=scandir($n);}`, "parse_url!false!null scandir!false!null", false},
		{"loop, element tested inside", `$u=parse_url($a); while($c--){if(isset($u['host'])){probe($u);} $u=scandir($n);}`, "parse_url!false!null scandir!false!null", false},
		{"loop, element not empty", `$u=parse_url($a); while($c--){if(!empty($u['a']['b'])){probe($u);} $u=scandir($n);}`, "parse_url!false!null scandir!false!null", false},
		{"loop, key tested", `$u=parse_url($a); while($c--){if(array_key_exists('host', $u)){probe($u);} $u=scandir($n);}`, "parse_url!false!null scandir!false!null", false},
		{"one path", `$u=parse_url($a); if($u === false){return;} if($c){$u=scandir($n);} probe($u);`, "parse_url!false scandir", false},
		{"one path, both tested alike", `$u=parse_url($a); if($u === null){return;} if($c){$u=scandir($n); if($u === null){return;}} probe($u);`, "parse_url!null scandir!null", false},
		{"both paths", `if($c){$u=parse_url($a);}else{$u=scandir($n);} probe($u);`, "parse_url scandir", false},
		{"same producer on both paths", `$u=parse_url($a); if($c){echo 1;} probe($u);`, "parse_url", true},
		{"unassigned on one path", `if($c){$u=parse_url($a);} probe($u);`, "parse_url", false},
		{"unassigned before a loop", `while($c--){probe($u); $u=parse_url($n);}`, "parse_url", false},
		{"unknown on one path", `$u=parse_url($a); if($c){extract($o);} probe($u);`, "parse_url", false},
		{"unknown on every path", `function f($o){extract($o); if($o){echo 1;} probe($u);}`, "", false},
		{"read before a do-while that assigns", `probe($u); do{$u=parse_url($n);}while($c--);`, "", false},
		{"assigned before a do-while that reassigns", `$u=parse_url($a); probe($u); do{$u=scandir($n);}while($c--);`, "parse_url", true},
		{"do-while", `$u=parse_url($a); do{probe($u); $u=scandir($n);}while($c--);`, "parse_url scandir", false},
		{"for, tested by its condition", `for($u=parse_url($a); $u !== false; $u=scandir($n)){probe($u);}`, "parse_url!false scandir!false", false},
		{"for, tested by the last condition", `for($u=parse_url($a); $c, $u; $u=scandir($n)){probe($u);}`, "parse_url!false!null scandir!false!null", false},
		{"for, tested by an earlier condition only", `for($u=parse_url($a); $u, $c; $u=scandir($n)){probe($u);}`, "parse_url scandir", false},
		{"for without a condition", `for($u=parse_url($a); ; $u=scandir($n)){probe($u);}`, "parse_url scandir", false},
		{"for, after the loop", `for($u=parse_url($a); !$u; $u=scandir($n)){echo 1;} probe($u);`, "parse_url!false!null scandir!false!null", false},
		{"fallback, both failed exits", `if(($u=parse_url($a)) === false && ($u=scandir($n)) === false){return;} probe($u);`, "parse_url!false scandir!false", false},
		{"fallback, either succeeded", `if(($u=parse_url($a)) !== false || ($u=scandir($n)) !== false){probe($u);}`, "parse_url!false scandir!false", false},
		{"fallback, negated", `if(!($u=parse_url($a)) && !($u=scandir($n))){return;} probe($u);`, "parse_url!false!null scandir!false!null", false},
		{"fallback, constant first", `if(false === ($u=parse_url($a)) && false === ($u=scandir($n))){return;} probe($u);`, "parse_url!false scandir!false", false},
		{"fallback, second not tested", `if(($u=parse_url($a)) === false && ($u=scandir($n)) > $c){return;} probe($u);`, "parse_url!false scandir", false},
		{"fallback, second tests another value", `if(($u=parse_url($a)) === false && $c === false){$u=scandir($n);} probe($u);`, "parse_url scandir", false},
		{"fallback, second compared with true", `if(($u=parse_url($a)) === false && ($u=scandir($n)) === true){return;} probe($u);`, "parse_url!false scandir", false},
		{"fallback skipped with an earlier result in place", `foreach($l as $n){if($skip || ($u=parse_url($n)) !== false){probe($u);} $skip=!$skip;}`, "parse_url", false},
		{"fallback exit skipped with an earlier result in place", `foreach($l as $n){if(!$skip && ($u=parse_url($n)) === false){continue;} probe($u); $skip=!$skip;}`, "parse_url", false},
		{"fallback skipped with the same result in place", `$u=($w=parse_url($a)); if($c || ($u=$w) !== false){probe($u);}`, "parse_url", true},
		{"fallback into one producer", `if($c && ($u=parse_url($a)) === false){return;} probe($u);`, "parse_url!false", false},
		{"skipped operand pins nothing", `$u=parse_url($a); if($u === false || $c){return;} probe($u);`, "parse_url!false", true},
		{"literal and call", `$u=['host'=>'x']; while($c--){probe($u); $u=parse_url($n);}`, "Array parse_url", false},
		{"copied", `$u=parse_url($a); if($c){$u=scandir($n);} $v=$u; probe($v);`, "parse_url scandir", false},
		{"too many", `$u=a(); if($c){$u=b();} if($c){$u=c();} if($c){$u=d();} if($c){$u=e();} if($c){$u=f();} probe($u);`, "", false},
		{"too many, then one more path", `$u=a(); if($c){$u=b();} if($c){$u=c();} if($c){$u=d();} if($c){$u=e();} while($c--){probe($u); if($c){$u=f();}}`, "", false},
		{"unsupported scope", `$u=parse_url($a); switch($c){case 1: $u=scandir($n);} probe($u);`, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, f := environment(t, tc.src, nil)
			found, single := e.Reaching(argument(callsNamed(f, "probe")[0]))
			var got []string
			for _, o := range found {
				name := strings.TrimPrefix(fmt.Sprintf("%T", o.Expr), "*syntax.")
				if call, ok := o.Expr.(*syntax.FuncCall); ok {
					name = call.Name.(*syntax.Name).Value
				}
				if o.NonFalse {
					name += "!false"
				}
				if o.NonNull {
					name += "!null"
				}
				got = append(got, name)
			}
			if strings.Join(got, " ") != tc.want || single != tc.single {
				t.Fatalf("reaching=%q single=%v", strings.Join(got, " "), single)
			}
		})
	}
	e, _ := environment(t, `1;`, nil)
	if found, known := e.Reaching(nil); found != nil || known {
		t.Fatal("absent expression has producers")
	}
}

func TestOriginSetsCompareByContent(t *testing.T) {
	x, y := &syntax.Literal{Raw: "1"}, &syntax.Literal{Raw: "2"}
	a := &originSet{items: []Origin{{Expr: x}, {Expr: y}}}
	for _, tc := range []struct {
		other *originSet
		same  bool
	}{
		{a, true},
		{&originSet{items: []Origin{{Expr: x}, {Expr: y}}}, true},
		{nil, false},
		{&originSet{items: []Origin{{Expr: x}}}, false},
		{&originSet{items: []Origin{{Expr: x}, {Expr: y, NonNull: true}}}, false},
		{&originSet{items: []Origin{{Expr: x}, {Expr: y}}, full: true}, false},
	} {
		if got := sameOrigins(a, tc.other); got != tc.same {
			t.Errorf("sameOrigins(%+v)=%v", tc.other, got)
		}
	}
	if (*originSet)(nil).excluding(true, true) != nil {
		t.Error("no producers gained a guard")
	}
}

// A summary keeps no syntax of the file it was computed from: the producers
// of a returned or sunk variable stay behind.
func TestSummariesKeepNoProducers(t *testing.T) {
	const src = `function pick($a, $c){$u=parse_url($a); if($c){$u=scandir($a);} system($u); return $u;}`
	e, f := environment(t, src, nil)
	sum := Extract(f, e.types.Index, e.types.PHP, nil).Summaries["pick"]
	if sum.Return.origins != nil || sum.Return.Expr != nil || len(sum.Sinks) != 1 || sum.Sinks[0].Argument.origins != nil {
		t.Fatalf("summary=%+v", sum)
	}
	again := Extract(syntax.Parse("test.php", []byte("<?php "+src), syntax.Options{Version: e.types.PHP}), e.types.Index, e.types.PHP, nil)
	if !NewSnapshot(again).Holds(&File{Path: "test.php", Summaries: map[string]Summary{"pick": sum}}) {
		t.Fatal("the same file summarized twice differs")
	}
}
