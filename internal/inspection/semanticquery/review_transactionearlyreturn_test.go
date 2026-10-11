package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewTransactionReturnProbe dispatches on returns only, as the rule package does.
type reviewTransactionReturnProbe struct{ nativeCheckProbe }

func (reviewTransactionReturnProbe) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KReturn}
}

func TestReviewTransactionEarlyReturn(t *testing.T) {
	const begin = "if(!$p->beginTransaction()){return false;}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"bare return", "if($x){return;}$p->commit();", 1},
		{"closed before return", "if($x){$p->rollBack();return;}$p->commit();", 0},
		{"commit returned", "return $p->commit();", 0},
		{"rollback returned", "return $p->rollBack();", 0},
		{"parenthesised", "return ($p->commit());", 0},
		{"negated", "return !$p->commit();", 0},
		{"silenced", "return @$p->commit();", 0},
		{"assigned", "return $ok=$p->commit();", 0},
		{"compound assigned", "return $ok.=$p->commit();", 0},
		{"skippable coalescing assignment", "return $ok??=$p->commit();", 1},
		{"left operand", "return $p->commit()&&$x;", 0},
		{"always evaluated right operand", "return $x===$p->commit();", 0},
		{"skippable and", "return $x&&$p->commit();", 1},
		{"skippable or", "return $x||$p->commit();", 1},
		{"skippable word and", "return ($x and $p->commit());", 1},
		{"skippable word or", "return ($x or $p->commit());", 1},
		{"skippable coalesce", "return $x??$p->commit();", 1},
		{"both ternary branches", "return $x?$p->commit():$p->rollBack();", 0},
		{"ternary condition", "return $p->commit()?1:2;", 0},
		{"one ternary branch", "return $x?$p->commit():false;", 1},
		{"other ternary branch", "return $x?false:$p->rollBack();", 1},
		{"short ternary", "return $x?:$p->commit();", 1},
		{"every match arm", "return match(true){$x=>$p->commit(),default=>$p->rollBack()};", 0},
		{"match subject", "return match($p->commit()){true=>1,default=>2};", 0},
		{"one match arm", "return match(true){$x=>$p->commit(),default=>false};", 1},
		{"empty match", "return match($x){};", 1},
		{"other method", "return $p->inTransaction();", 1},
		{"other connection", "return $q->commit();", 1},
		{"call argument", "return wrap($p->commit());", 1},
		{"plain value", "return $x;", 1},
		{"direct finally", "try{if($x){return;}$p->commit();}finally{$p->rollBack();}", 0},
		{"guarded finally", "try{if($x){return;}$p->commit();}finally{if($p->inTransaction()){$p->rollBack();}}", 0},
		{"guarded finally commit", "try{if($x){return;}}finally{if($p->inTransaction()){$p->commit();}}", 0},
		{"guarded finally without braces", "try{if($x){return;}$p->commit();}finally{if($p->inTransaction())$p->rollBack();}", 0},
		{"guarded finally after other statements", "try{if($x){return;}$p->commit();}finally{log_it();if($p->inTransaction()){log_it();$p->rollBack();}}", 0},
		{"outer guarded finally", "try{try{if($x){return;}$p->commit();}finally{log_it();}}finally{if($p->inTransaction()){$p->rollBack();}}", 0},
		{"finally under another condition", "try{if($x){return;}$p->commit();}finally{if($x){$p->rollBack();}}", 1},
		{"finally under negated guard", "try{if($x){return;}$p->commit();}finally{if(!$p->inTransaction()){$p->rollBack();}}", 1},
		{"finally guard on other connection", "try{if($x){return;}$p->commit();}finally{if($q->inTransaction()){$q->rollBack();}}", 1},
		{"finally guard closes other connection", "try{if($x){return;}$p->commit();}finally{if($p->inTransaction()){$q->rollBack();}}", 1},
		{"finally guard without close", "try{if($x){return;}$p->commit();}finally{if($p->inTransaction()){log_it();}}", 1},
		{"finally guard closes conditionally", "try{if($x){return;}$p->commit();}finally{if($p->inTransaction()){if($x){$p->rollBack();}}}", 1},
		{"return inside finally", "try{work();}finally{if($x){return;}if($p->inTransaction()){$p->rollBack();}}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewTransactionReturnProbe{nativeCheckProbe{"TransactionEarlyReturn", CheckTransactionEarlyReturn}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				t.Fatal(err)
			}
			src := "<?php function f(PDO $p,PDO $q,$x){" + begin + tc.src + "}"
			if got := e.Analyze(syntax.Parse("transaction.php", []byte(src), syntax.Options{})); len(got) != tc.want {
				t.Fatalf("got %d, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
