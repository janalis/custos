package transactionexceptionwithoutrollback

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewTransactionExceptionWithoutRollback(t *testing.T) {
	const head = "function f(PDO $p, PDO $q, $m){if(!$p->beginTransaction()){return;}try{throw new E();}catch(Throwable $e){"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"plain return", "return false;}}", 1},
		{"bare return", "return;}}", 1},
		{"other receiver in return", "return $q->rollBack();}}", 1},
		{"other method in return", "return $p->inTransaction();}}", 1},
		{"dynamic method in return", "return $p->$m();}}", 1},
		{"rollback in one ternary arm", "return $m ? $p->rollBack() : false;}}", 1},
		{"rollback in else arm only", "return $m ? false : $p->rollBack();}}", 1},
		{"short ternary fallback", "return $m ?: $p->rollBack();}}", 1},
		{"rollback in one match arm", "return match($m){1=>$p->rollBack(),default=>false};}}", 1},
		{"short-circuit operand", "return $m && $p->rollBack();}}", 1},
		{"finally guard on other receiver", "return false;}finally{if($q->inTransaction()){$p->rollBack();}}}", 1},
		{"finally guard on other condition", "return false;}finally{if($m){$p->rollBack();}}}", 1},
		{"finally guard ends other receiver", "return false;}finally{if($p->inTransaction()){$q->rollBack();}}}", 1},
		{"finally guard without rollback", "return false;}finally{if($p->inTransaction()){echo 'open';}}}", 1},
		{"finally rollback nested deeper", "return false;}finally{if($p->inTransaction()){if($m){$p->rollBack();}}}}", 1},
		{"finally without transaction call", "return false;}finally{echo 'done';}}", 1},
		{"returned rollback", "return $p->rollBack();}}", 0},
		{"returned commit", "return $p->commit();}}", 0},
		{"case-insensitive rollback", "return $p->ROLLBACK();}}", 0},
		{"parenthesised rollback", "return ($p->rollBack());}}", 0},
		{"negated rollback", "return !$p->rollBack();}}", 0},
		{"rollback as ternary condition", "return $p->rollBack() ? 1 : 2;}}", 0},
		{"rollback in both ternary arms", "return $m ? $p->rollBack() : $p->commit();}}", 0},
		{"rollback as match subject", "return match($p->rollBack()){true=>1,false=>2};}}", 0},
		{"rollback in every match arm", "return match($m){1=>$p->rollBack(),default=>$p->commit()};}}", 0},
		{"guarded finally rollback", "return false;}finally{if($p->inTransaction()){$p->rollBack();}}}", 0},
		{"guarded finally commit after statement", "return false;}finally{if($p->inTransaction()){echo 'open';$p->commit();}}}", 0},
		{"guarded finally single statement", "return false;}finally{if($p->inTransaction())$p->rollBack();}}", 0},
		{"unguarded finally rollback", "return false;}finally{$p->rollBack();}}", 0},
		{"outer guarded finally", "try{return false;}finally{echo 'inner';}}finally{if($p->inTransaction()){$p->rollBack();}}}", 0},
		{"rollback before return", "$p->rollBack();return false;}}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := expansionCryptoDBProbe{"TransactionExceptionWithoutRollback", []syntax.NodeKind{syntax.KReturn}, func(ctx *analysis.Context, n syntax.Node) {
				check(ctx, n, "m")
			}}
			engine, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{probe.id}})
			if err != nil {
				t.Fatal(err)
			}
			got := engine.Analyze(syntax.Parse("review.php", []byte("<?php "+head+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

type expansionCryptoDBProbe struct {
	id    string
	kinds []syntax.NodeKind
	check func(*analysis.Context, syntax.Node)
}

func (p expansionCryptoDBProbe) ID() string { return p.id }

func (p expansionCryptoDBProbe) Semantic() {}

func (p expansionCryptoDBProbe) Flow() {}

func (p expansionCryptoDBProbe) Kinds() []syntax.NodeKind { return p.kinds }

func (p expansionCryptoDBProbe) Check(ctx *analysis.Context, n syntax.Node) { p.check(ctx, n) }
