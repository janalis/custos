package ftppendingtransferacceptedascomplete

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewFtpPendingTransferAcceptedAsComplete(t *testing.T) {
	const call = "ftp_nb_get($f,$l,$r,FTP_BINARY)"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"truthy test", "if(" + call + "){return true;}", 1},
		{"parenthesised", "if((" + call + ")){return true;}", 1},
		{"assigned status", "if($s=" + call + "){return true;}", 1},
		{"loose true", "if(" + call + "==true){return true;}", 1},
		{"loose true on the left", "if(true==" + call + "){return true;}", 1},
		{"loose not false", "if(" + call + "!=false){return true;}", 1},
		{"strict not false always passes", "if(" + call + "!==false){return true;}", 1},
		{"strict not true always passes", "if(" + call + "!==true){return true;}", 1},
		{"double negation", "if(!!" + call + "){return true;}", 1},
		{"negation strictly false", "if(!" + call + "===false){return true;}", 1},
		{"negation strictly not true", "if(!" + call + "!==true){return true;}", 1},
		{"negated test with success else", "if(!" + call + "){return false;}else{return true;}", 1},
		{"negated test, else after elseif", "if(!" + call + "){return false;}elseif($x){return false;}else{echo 'complete';}", 1},
		{"strict false never matches, else always runs", "if(" + call + "===false){return false;}else{return true;}", 1},

		{"failure branch", "if(!" + call + "){return true;}", 0},
		{"failure branch without else", "if(!" + call + "){return true;}return false;", 0},
		{"loose false", "if(" + call + "==false){return true;}", 0},
		{"loose not true", "if(" + call + "!=true){return true;}", 0},
		{"strict true never matches", "if(" + call + "===true){return true;}", 0},
		{"strict false never matches", "if(" + call + "===false){return true;}", 0},
		{"negation strictly true", "if(!" + call + "===true){return true;}", 0},
		{"negation strictly not false", "if(!" + call + "!==false){return true;}", 0},
		{"negated test, failing else", "if(!" + call + "){return true;}else{return false;}", 0},
		{"else if is its own test", "if(!" + call + "){return false;}else if($x){return true;}", 0},
		{"guard on FTP_FINISHED", "if($s=" + call + "){while($s===FTP_MOREDATA){$s=ftp_nb_continue($f);}if($s!==FTP_FINISHED){return false;}return true;}", 0},
		{"qualified FTP_FINISHED", "if($s=" + call + "){if($s!==\\FTP_FINISHED){return false;}echo 'complete';}", 0},
		{"polling loop", "if(" + call + "){while(ftp_nb_continue($f)!==FTP_FINISHED){}return true;}", 0},
		{"FTP_FINISHED only in nested scope", "if($s=" + call + "){$g=function($s){return $s===FTP_FINISHED;};return true;}", 1},
		{"namespaced lookalike constant", "namespace App;const FTP_FINISHED=9;if($s=\\" + call + "){if($s!==FTP_FINISHED){return false;}return true;}", 1},
		{"other constant", "if($s=" + call + "){if($s===FTP_MOREDATA){return false;}return true;}", 1},
		{"compound assignment", "if($s.=" + call + "){return true;}", 0},
		{"assignment target", "if($a[" + call + "]=1){return true;}", 0},
		{"other unary", "if(-" + call + "){return true;}", 0},
		{"ordering comparison with boolean", "if(" + call + ">false){return true;}", 0},
		{"non boolean comparison", "if(" + call + "==1){return true;}", 0},
		{"while condition", "while(" + call + "){return true;}", 0},
		{"statement", call + ";return true;", 0},
		{"first-class callable", "if(ftp_nb_get(...)){return true;}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probe := New()
			e, err := analysis.NewEngine([]analysis.Rule{probe}, analysis.Config{Only: []string{"FtpPendingTransferAcceptedAsComplete"}})
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
