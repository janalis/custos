package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewMysqliDrainProbe dispatches on method calls only, as the rule package does.
type reviewMysqliDrainProbe struct{ nativeCheckProbe }

func (reviewMysqliDrainProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func TestReviewMysqliMultiQueryResultsNotDrained(t *testing.T) {
	const (
		open  = "function f(mysqli $db, mysqli $other){if($db->multi_query('SELECT 7; SELECT 11')){"
		store = "if($r=$db->store_result()){$r->free();}"
		query = "$db->query('SELECT 14');}}"
	)
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"undrained", open + query, 1},
		{"do-while next_result", open + "do{" + store + "}while($db->next_result());" + query, 0},
		{"do-while more and next", open + "do{" + store + "}while($db->more_results()&&$db->next_result());" + query, 0},
		{"do-while parenthesised", open + "do{" + store + "}while(($db->more_results())&&($db->next_result()));" + query, 0},
		{"do-while more with next in body", open + "do{" + store + "$db->next_result();}while($db->more_results());" + query, 0},
		{"while next_result", open + "while($db->next_result()){" + store + "}" + query, 0},
		{"while more and next", open + "while($db->more_results()&&$db->next_result()){" + store + "}" + query, 0},
		{"while more with next in body", open + "while($db->more_results()){$db->next_result();}" + query, 0},
		{"while more without next", open + "while($db->more_results()){" + store + "}" + query, 1},
		{"do-while more without next", open + "do{" + store + "}while($db->more_results());" + query, 1},
		{"negated next_result", open + "do{" + store + "}while(!$db->next_result());" + query, 1},
		{"next_result or flag", open + "do{" + store + "}while($db->next_result()||$x);" + query, 1},
		{"next_result compared", open + "do{" + store + "}while($db->next_result()===false);" + query, 1},
		{"next_result in if only", open + "if($db->next_result()){" + store + "}" + query, 1},
		{"next_result in for condition", open + "for(;$db->next_result();){" + store + "}" + query, 1},
		{"next_result in while body condition", open + "while($x){if($db->next_result()){}}" + query, 1},
		{"other connection drained", open + "do{}while($other->next_result());" + query, 1},
		{"query inside while drain loop", open + "while($db->next_result()){$db->query('SELECT 14');}}}", 1},
		{"query inside do-while drain loop", open + "do{$db->query('SELECT 14');}while($db->next_result());}}", 1},
		{"query inside more_results loop", open + "while($db->more_results()){$db->next_result();$db->query('SELECT 14');}}}", 1},
		{"drain before multi_query", "function f(mysqli $db){do{}while($db->next_result());if($db->multi_query('SELECT 7; SELECT 11')){" + query, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := reviewMysqliDrainProbe{nativeCheckProbe{"MysqliMultiQueryResultsNotDrained", CheckMysqliMultiQueryResultsNotDrained}}
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"MysqliMultiQueryResultsNotDrained"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
