package directoryiteratordotentries

import (
	"strings"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestReviewDirectoryIteratorDotEntries(t *testing.T) {
	const kill = "unlink($e->getPathname());"
	loop := func(body string) string { return "foreach(new DirectoryIterator('/tmp') as $e){" + body + "}" }
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		// Early exits before the call.
		{"isDot continue", "if($e->isDot()){continue;}" + kill, 0},
		{"isDot continue, parenthesized", "if(($e->isDot())){continue;}" + kill, 0},
		{"isDot continue, method case", "if($e->ISDOT()){continue;}" + kill, 0},
		{"isDot or isDir continue", "if($e->isDot()||$e->isDir()){continue;}" + kill, 0},
		{"isDot or unrelated continue", "if($e->isDot() or $skip){continue;}" + kill, 0},
		{"isDir continue", "if($e->isDir()){continue;}" + kill, 0},
		{"not isFile continue", "if(!$e->isFile()){continue;}" + kill, 0},
		{"not isLink continue", "if(!$e->isLink()){continue;}" + kill, 0},
		{"double negation", "if(!!$e->isDot()){continue;}" + kill, 0},
		{"both names rejected", "if($e->getFilename()==='.'||$e->getFilename()==='..'){continue;}" + kill, 0},
		{"names rejected loosely, literal first", "if('.'==$e->getBasename()||'..'==$e->getBasename()){continue;}" + kill, 0},
		{"names rejected by two exits", "if($e->getFilename()==='.'){continue;}if($e->getFilename()==='..'){continue;}" + kill, 0},
		{"names rejected by in_array", "if(in_array($e->getFilename(),['.','..'],true)){continue;}" + kill, 0},
		{"not both of two accepted names", "if(!($e->getFilename()!=='.'&&$e->getFilename()!=='..')){continue;}" + kill, 0},
		{"isDot and unrelated continue", "if($e->isDot()&&$skip){continue;}" + kill, 1},
		{"isFile continue", "if($e->isFile()){continue;}" + kill, 1},
		{"not isDot continue", "if(!$e->isDot()){continue;}" + kill, 1},
		{"not isDir continue", "if(!$e->isDir()){continue;}" + kill, 1},
		{"one name rejected", "if($e->getFilename()==='.'){continue;}" + kill, 1},
		{"one name in list", "if(in_array($e->getFilename(),['.','.git'])){continue;}" + kill, 1},
		{"unknown list", "if(in_array($e->getFilename(),$names)){continue;}" + kill, 1},
		{"other needle", "if(in_array($name,['.','..'])){continue;}" + kill, 1},
		{"other function", "if(str_contains($e->getFilename(),'.')){continue;}" + kill, 1},
		{"always true condition leaves the mutation unreachable", "if($e->getFilename()!=='.'||$e->getFilename()!=='..'){continue;}" + kill, 0},
		{"exit off the path", "if($skip){if($e->isDot()){continue;}}" + kill, 1},
		{"one name ruled in", "if($e->getFilename()!=='.'){continue;}" + kill, 1},
		{"saved name test", "$dot=$e->isDot();if($dot){continue;}" + kill, 0},
		{"exit with else", "if($e->isDot()){continue;}else{$x=1;}" + kill, 0},
		{"match arm", "echo match(true){$e->isDot()=>0,default=>" + strings.TrimSuffix(kill, ";") + "};", 0},
		{"jump label after test", "if($e->isDot()){continue;}back:" + kill + "goto back;", 1},
		{"name with suffix argument", "if($e->getBasename('.')==='.'||$e->getBasename('.')==='..'){continue;}" + kill, 1},
		{"other name method", "if($e->getExtension()==='.'||$e->getExtension()==='..'){continue;}" + kill, 1},
		{"unrelated comparison", "if($a==='.'||$b==='..'){continue;}" + kill, 1},
		{"other receiver", "if($o->isDot()){continue;}" + kill, 1},
		{"non variable receiver", "if($this->e->isDot()){continue;}" + kill, 1},
		{"dynamic method", "if($e->$m()){continue;}" + kill, 1},
		{"other unary", "if(-$e->isDot()){continue;}" + kill, 1},
		{"other binary", "if($e->isDot()+1){continue;}" + kill, 1},
		{"exit does not terminate", "if($e->isDot()){log_it();}" + kill, 1},
		{"exit after the call", kill + "if($e->isDot()){continue;}", 1},
		// Enclosing branches.
		{"isFile branch", "if($e->isFile()){" + kill + "}", 0},
		{"isLink branch", "if($e->isLink()){" + kill + "}", 0},
		{"not isDot and isFile branch", "if(!$e->isDot()&&$e->isFile()){" + kill + "}", 0},
		{"not isDot and unrelated branch", "if(!$e->isDot() and $ok){" + kill + "}", 0},
		{"accepted names branch", "if($e->getFilename()!=='.'&&$e->getFilename()!='..'){" + kill + "}", 0},
		{"not in_array branch", "if(!in_array($e->getFilename(),['.','..'])){" + kill + "}", 0},
		{"else of isDot", "if($e->isDot()){log_it();}else{" + kill + "}", 0},
		{"else if after isDot", "if($e->isDot()){log_it();}else if($ok){" + kill + "}", 0},
		{"elseif after isDot", "if($e->isDot()){log_it();}elseif($ok){" + kill + "}", 0},
		{"elseif isFile", "if($ok){log_it();}elseif($e->isFile()){" + kill + "}", 0},
		{"elseif unrelated", "if($ok){log_it();}elseif($other){" + kill + "}", 1},
		{"not isDot or unrelated branch", "if(!$e->isDot()||$force){" + kill + "}", 1},
		{"isDot branch", "if($e->isDot()){" + kill + "}", 1},
		{"else of isFile", "if($e->isFile()){log_it();}else{" + kill + "}", 1},
		{"in_array branch", "if(in_array($e->getFilename(),['.','..'])){" + kill + "}", 1},
		{"one accepted name branch", "if($e->getFilename()!=='.'){" + kill + "}", 1},
		// Short-circuit operands.
		{"and operand", "$e->isFile()&&" + kill, 0},
		{"keyword and operand", "!$e->isDot() and " + kill, 0},
		{"or operand", "$e->isDot()||" + kill, 0},
		{"keyword or operand", "$e->isDir() or " + kill, 0},
		{"left operand", "unlink($e->getPathname())&&$e->isFile();", 1},
		{"inside the condition", "if($e->isDot()||unlink($e->getPathname())){log_it();}", 0},
		{"condition call with later guard", "if(unlink($e->getPathname())&&$e->isFile()){log_it();}", 1},
		{"other operator operand", "$r=$e->isFile()+unlink($e->getPathname());", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzeReview(t, loop(tc.src))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func analyzeReview(t *testing.T, src string) []diagnostic.Finding {
	t.Helper()
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DirectoryIteratorDotEntries"}})
	if err != nil {
		t.Fatal(err)
	}
	return e.Analyze(syntax.Parse("review.php", []byte("<?php "+src), syntax.Options{}))
}
