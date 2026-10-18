package unescapedhtmloutput

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// A replacement of a request table clears only what is read from it later:
// text copied before keeps the request text.
func TestRequestReadTime(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source   string
		findings int
	}{
		{`$v=$_GET['v'];$_GET['v']='safe';echo '<p>'.$v;`, 1},
		{`$v=$_GET['v'];unset($_GET['v']);echo '<p>'.$v;`, 1},
		{`$v=$_GET['v'];$_GET=[];echo '<p>'.$v;`, 1},
		{`$v=$_GET['v'];$_GET['v']=htmlspecialchars($_GET['v']);echo '<p>'.$v;`, 1},
		{`$v=$_GET['v'];$copy=$v;$_GET['v']='safe';echo '<p>'.$copy;`, 1},
		{`$v=$_GET['v'];function clear(&$v){$v='safe';}clear($_GET['v']);echo '<p>'.$v;`, 1},
		{`$v=$_GET['v'];$_GET['v']='safe';echo '<p>'.strtoupper($v);`, 1},
		{`$v=$_GET['v'];$_GET['v']='safe';echo '<p title="'.$v.'">';`, 1},
		{`$v=$_GET;$_GET=[];echo '<p>'.$v['v'];`, 1},
		{`if($flag){$v=$_GET['v'];}else{$v=$_POST['v'];}$_GET=[];$_POST=[];echo '<p>'.$v;`, 1},
		{`$_GET['v']='safe';$v=$_GET['v'];echo '<p>'.$v;`, 0},
		{`$_GET['v']=htmlspecialchars($_GET['v']);$v=$_GET['v'];echo '<p>'.$v;`, 0},
		{`$v=htmlspecialchars($_GET['v']);$_GET=[];echo '<p>'.$v;`, 0},
		{`$v=$_GET['v'];$_GET['v']='safe';echo '<p>'.htmlspecialchars($v);`, 0},
		{`$_GET['v']='safe';echo '<p>'.$_GET['v'];`, 0},
		// The value is stored once computed: a read inside it is earlier.
		{`$_GET['v']=$_GET['v'].'';echo '<p>'.$_GET['v'];`, 1},
		// In a loop, a replacement after the read leaves the first turn.
		{`while($c){$v=$_GET['v'];$_GET['v']='safe';echo '<p>'.$v;}`, 1},
		{`function g(){return $_GET['v'];}$_GET['v']='safe';echo '<p>'.g();`, 0},
		// A call of the output statement runs before what follows it.
		{`function refill(){$_GET['v']=$_POST['v'];}$_GET['v']='safe';echo '<p>'.(refill().$_GET['v']);`, 1},
		{`function refill(){$_GET['v']=$_POST['v'];}$_GET['v']='safe';echo '<p>'.($_GET['v'].refill());`, 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			f := syntax.Parse("output.php", []byte("<?php "+tc.source), syntax.Options{})
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("got %+v, want %d", got, tc.findings)
			}
		})
	}
}
