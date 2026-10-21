package curlresponsezerorejected

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestReviewCurlResponseZeroRejected(t *testing.T) {
	const on = "curl_setopt($c,CURLOPT_RETURNTRANSFER,true);"
	const fail = "if(!curl_exec($c)){throw new RuntimeException();}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"plain negation", on + fail, 1},
		{"other option keeps proof", on + "curl_setopt($c,CURLOPT_TIMEOUT,5);" + fail, 1},
		{"parenthesized option", "curl_setopt($c,(CURLOPT_RETURNTRANSFER),1);" + fail, 1},
		{"file sink", on + "curl_setopt($c,CURLOPT_FILE,$fp);" + fail, 0},
		{"write callback sink", on + "curl_setopt($c,CURLOPT_WRITEFUNCTION,function($c,$d){return strlen($d);});" + fail, 0},
		{"transfer restored after sink", "curl_setopt($c,CURLOPT_FILE,$fp);" + on + fail, 1},
		{"dynamic option", on + "curl_setopt($c,$opt,false);" + fail, 0},
		{"class constant option", on + "curl_setopt($c,Opt::RETURN,false);" + fail, 0},
		{"user constant option", "const MY_OPT=19913;" + on + "curl_setopt($c,MY_OPT,false);" + fail, 0},
		{"unknown transfer value", on + "curl_setopt($c,CURLOPT_RETURNTRANSFER,$flag);" + fail, 0},
		{"unknown options array", on + "curl_setopt_array($c,$opts);" + fail, 0},
		{"transfer restored after unknown array", "curl_setopt_array($c,$opts);" + on + fail, 1},
		{"literal array enables", "curl_setopt_array($c,[CURLOPT_TIMEOUT=>5,CURLOPT_RETURNTRANSFER=>true]);" + fail, 1},
		{"literal array keeps proof", on + "curl_setopt_array($c,[CURLOPT_TIMEOUT=>5]);" + fail, 1},
		{"literal array disables", on + "curl_setopt_array($c,[CURLOPT_RETURNTRANSFER=>false]);" + fail, 0},
		{"array sink after transfer", "curl_setopt_array($c,[CURLOPT_RETURNTRANSFER=>true,CURLOPT_FILE=>$fp]);" + fail, 0},
		{"array transfer after sink", "curl_setopt_array($c,[CURLOPT_FILE=>$fp,CURLOPT_RETURNTRANSFER=>true]);" + fail, 1},
		{"array dynamic key", on + "curl_setopt_array($c,[$opt=>false]);" + fail, 0},
		{"array spread", on + "curl_setopt_array($c,[...$opts]);" + fail, 0},
		{"unknown call on handle", on + "configure($c);" + fail, 0},
		{"other builtin call on handle", on + "curl_reset($c);" + fail, 0},
		{"conditional write", on + "if($f){curl_setopt($c,CURLOPT_FILE,$fp);}" + fail, 0},
		{"other handle sink", on + "curl_setopt($d,CURLOPT_FILE,$fp);" + fail, 1},
		{"double negation", on + "if(!!curl_exec($c)){throw new RuntimeException();}", 0},
		{"triple negation", on + "if(!!!curl_exec($c)){throw new RuntimeException();}", 1},
		{"negated loose false", on + "if(!(curl_exec($c)==false)){throw new RuntimeException();}", 0},
		{"loose false", on + "if(curl_exec($c)==false){return false;}", 1},
		{"negated assignment", on + "if(!($r=curl_exec($c))){return false;}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := engine(t).Analyze(syntax.Parse("audit.php", []byte("<?php "+tc.src), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
