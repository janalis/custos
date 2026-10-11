package shortstreamreadunchecked

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewLengthGuards(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ShortStreamReadUnchecked"}})
	if err != nil {
		t.Fatal(err)
	}
	const four, one = "$v=fread($h,4);", "$v=fread($h,1);"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"lower bound exit", four + "if(strlen($v)<4){return;}unpack('N',$v);", 0},
		{"mirrored lower bound exit", four + "if(4>strlen($v)){return;}unpack('N',$v);", 0},
		{"at most exit", four + "if(strlen($v)<=3){return;}unpack('N',$v);", 0},
		{"mirrored at most exit", four + "if(3>=strlen($v)){return;}unpack('N',$v);", 0},
		{"at least branch", four + "if(strlen($v)>=4){unpack('N',$v);}", 0},
		{"mirrored at least branch", four + "if(4<=strlen($v)){unpack('N',$v);}", 0},
		{"mirrored greater branch", four + "if(3<strlen($v)){unpack('N',$v);}", 0},
		{"greater branch", four + "if(strlen($v)>3){unpack('N',$v);}", 0},
		{"loose equal branch", four + "if(strlen($v)==4){unpack('N',$v);}", 0},
		{"angle unequal exit", four + "if(strlen($v)<>4){return;}unpack('N',$v);", 0},
		{"negated equality exit", four + "if(!(strlen($v)===4)){return;}unpack('N',$v);", 0},
		{"keyword disjunction exit", four + "if($v===false or strlen($v)!==4){return;}unpack('N',$v);", 0},
		{"keyword conjunction branch", four + "if(strlen($v)===4 and $ok){unpack('N',$v);}", 0},
		{"earlier exit", four + "if(strlen($v)!==4){return;}$n=1;if($n){unpack('N',$v);}", 0},
		{"else branch", four + "if(strlen($v)!==4){$n=0;}else{unpack('N',$v);}", 0},
		{"elseif branch", four + "if($n){}elseif(strlen($v)===4){unpack('N',$v);}", 0},
		{"after failed elseif", four + "if($n){}elseif(strlen($v)!==4){}else{unpack('N',$v);}", 0},
		{"ternary then", four + "$r=strlen($v)===4?unpack('N',$v):null;", 0},
		{"ternary else", four + "$r=strlen($v)!==4?null:unpack('N',$v);", 0},
		{"short ternary else", four + "$r=strlen($v)!==4?:unpack('N',$v);", 0},
		{"disjunction right operand", four + "$r=strlen($v)!==4||unpack('N',$v);", 0},
		{"unbraced body", four + "if(strlen($v)===4)unpack('N',$v);", 0},
		{"isset branch", four + "if(isset($v[3])){unpack('N',$v);}", 0},
		{"byte encoding constant case", four + "if(mb_strlen($v,'8BIT')===4){unpack('N',$v);}", 0},
		{"single byte non-zero length", one + "if(strlen($v)!==0){unpack('C',$v);}", 0},
		{"single byte zero length exit", one + "if(strlen($v)==0){return;}unpack('C',$v);", 0},
		{"single byte positive length", one + "if(strlen($v)>0){unpack('C',$v);}", 0},
		{"single byte truthy", one + "if($v){unpack('C',$v);}", 0},
		{"single byte not empty", one + "if(!empty($v)){unpack('C',$v);}", 0},
		{"single byte loose empty exit", one + "if($v==''){return;}unpack('C',$v);", 0},
		{"single byte loose failure", one + "if(false!=$v){unpack('C',$v);}", 0},
		{"single byte separate exits", one + "if($v===false){return;}if(''===$v){return;}unpack('C',$v);", 0},

		{"upper bound branch", four + "if(strlen($v)<=4){unpack('N',$v);}", 1},
		{"smaller branch", four + "if(strlen($v)<5){unpack('N',$v);}", 1},
		{"other length", four + "if(strlen($v)===3){unpack('N',$v);}", 1},
		{"non-zero length of four", four + "if(strlen($v)!==0){unpack('N',$v);}", 1},
		{"unknown bound", four + "if(strlen($v)===$n){unpack('N',$v);}", 1},
		{"arithmetic on length", four + "if(strlen($v)+0===4){unpack('N',$v);}", 1},
		{"negated other operator", four + "if(-strlen($v)){unpack('N',$v);}", 1},
		{"truthy four bytes", four + "if($v){unpack('N',$v);}", 1},
		{"falsy branch", one + "if(!$v){unpack('C',$v);}", 1},
		{"empty branch", one + "if(empty($v)){unpack('C',$v);}", 1},
		{"only non-empty", one + "if($v!==''){unpack('C',$v);}", 1},
		{"only non-false", one + "if($v!==false){unpack('C',$v);}", 1},
		{"other sentinel", one + "if($v!==null&&$v!=='x'){unpack('C',$v);}", 1},
		{"true sentinel", one + "if($v!=true){unpack('C',$v);}", 1},
		{"at least exit", four + "if(strlen($v)>=4){return;}unpack('N',$v);", 1},
		{"greater exit", four + "if(strlen($v)>3){return;}unpack('N',$v);", 1},
		{"difference exit", four + "if(strlen($v)-4){return;}unpack('N',$v);", 1},
		{"isset other offset", four + "if(isset($v[2])){unpack('N',$v);}", 1},
		{"isset unknown offset", four + "if(isset($v[$i])){unpack('N',$v);}", 1},
		{"isset whole value", four + "if(isset($v)){unpack('N',$v);}", 1},
		{"isset other value", four + "if(isset($w[3])){unpack('N',$v);}", 1},
		{"isset several", four + "if(isset($v[3],$w)){unpack('N',$v);}", 1},
		{"missing offset branch", four + "if(!isset($v[3])){unpack('N',$v);}", 1},
		{"character encoding", four + "if(mb_strlen($v,'UTF-8')===4){unpack('N',$v);}", 1},
		{"dynamic encoding", four + "if(mb_strlen($v,$enc)===4){unpack('N',$v);}", 1},
		{"other function", four + "if(count($v)===4){unpack('N',$v);}", 1},
		{"user strlen", "namespace App;function strlen($s){return 4;}" + four + "if(strlen($v)===4){\\unpack('N',$v);}", 1},
		{"disjunction branch", four + "if(strlen($v)===4||$ok){unpack('N',$v);}", 1},
		{"conjunction exit", four + "if(strlen($v)!==4&&$ok){return;}unpack('N',$v);", 1},
		{"exit with else", four + "if(strlen($v)!==4){return;}else{$n=1;}unpack('N',$v);", 1},
		{"exit with elseif", four + "if(strlen($v)!==4){return;}elseif($n){$n=1;}unpack('N',$v);", 1},
		{"later guard", four + "unpack('N',$v);if(strlen($v)!==4){return;}", 1},
		{"read in condition", four + "if(unpack('N',$v)){$n=1;}", 1},
		{"read in elseif condition", four + "if(strlen($v)!==4){}elseif(unpack('N',$v)){}", 1},
		{"ternary condition", four + "$r=unpack('N',$v)?1:2;", 1},
		{"left operand", four + "$r=unpack('N',$v)&&strlen($v)===4;", 1},
		{"concatenated", four + "$r='x'.unpack('N',$v);", 1},
		{"reassigned after guard", four + "if(strlen($v)!==4){return;}$v=fread($h,4);unpack('N',$v);", 1},
		{"guard outside closure", four + "if(strlen($v)!==4){return;}$f=function()use($h){$v=fread($h,4);unpack('N',$v);};", 1},
		{"inline read", "if(strlen(fread($h,4))===4){unpack('N',fread($h,4));}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
