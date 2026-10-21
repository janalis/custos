package pdoexeczerorejectedasfailure

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewPdoExecZeroRejectedAsFailure(t *testing.T) {
	const head = "function f(PDO $p, PDO $q, $s, $m){"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"direct throw", "if(!$p->exec($s)){throw new E();}}", 1},
		{"loose false return", "if($p->exec($s)==false){return false;}}", 1},
		{"assigned and parenthesised", "if(!($n=$p->exec($s))){return false;}}", 1},
		{"single statement return", "if(!$p->exec($s))return false;}", 1},
		{"single statement throw", "if(!$p->exec($s))throw new E();}", 1},
		{"message from error info", "if(!$p->exec($s)){throw new E($p->errorInfo()[2]);}}", 1},
		{"error info then plain statements", "if(!$p->exec($s)){$i=$p->errorInfo();echo 'x';throw new E($i[2]);}}", 1},
		{"error code case insensitive direct", "if(!$p->exec($s)){$c=$p->ERRORCODE();return false;}}", 1},
		{"other receiver consulted", "if(!$p->exec($s)){if($q->errorCode()!=='00000'){throw new E();}}}", 1},
		{"unrelated guard", "if(!$p->exec($s)){if($m){throw new E();}}}", 1},
		{"unrelated method", "if(!$p->exec($s)){if($p->inTransaction()){throw new E();}}}", 1},
		{"dynamic method", "if(!$p->exec($s)){if($p->$m()){throw new E();}}}", 1},
		{"second exit unguarded", "if(!$p->exec($s)){if($m){return false;}if($p->errorCode()!=='00000'){throw new E();}}}", 1},
		{"guarded by error code", "if(!$p->exec($s)){if($p->errorCode()!=='00000'){throw new E($p->errorInfo()[2]);}}}", 0},
		{"guarded by error info", "if(!$p->exec($s)){if($p->errorInfo()[0]!=='00000'){return false;}}}", 0},
		{"state held in variable", "if($p->exec($s)==false){$c=$p->errorCode();if($c!=='00000'){return false;}}}", 0},
		{"early return on zero rows", "if(!$p->exec($s)){if($p->errorCode()==='00000'){return 0;}throw new E();}}", 0},
		{"throw expression guard", "if(!$p->exec($s)){$p->errorCode()==='00000'||throw new E();}}", 0},
		{"guard in single statement branch", "if(!$p->exec($s))if($p->errorCode()!=='00000')throw new E();}", 0},
		{"exit in nested scope only", "if(!$p->exec($s)){$g=function(){throw new E();};}}", 0},
		{"unreachable exits", "if(!$p->exec($s)){if(false){throw new E();}if(false){return false;}}}", 0},
		{"return other value", "if(!$p->exec($s)){return null;}}", 0},
		{"no rejection", "if(!$p->exec($s)){echo 'none';}}", 0},
		{"strict comparison", "if($p->exec($s)===false){throw new E();}}", 0},
		{"loose comparison with zero", "if($p->exec($s)==0){throw new E();}}", 0},
		{"truthy condition", "if($p->exec($s)){throw new E();}}", 0},
		{"assigned truthy condition", "if($n=$p->exec($s)){throw new E();}}", 0},
		{"other unary", "if(-$p->exec($s)){throw new E();}}", 0},
		{"result not in condition", "$n=$p->exec($s);if(!$n){throw new E();}}", 0},
		{"other method", "if(!$p->query($s)){throw new E();}}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := expansionCryptoDBProbe{"PdoExecZeroRejectedAsFailure", []syntax.NodeKind{syntax.KMethodCall}, func(ctx *analysis.Context, n syntax.Node) {
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
