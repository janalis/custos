package unescapedhtmloutput

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
)

func TestAttributeOutput(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		version      phpversion.Version
		want         int
	}{
		{"double quoted raw", `echo '<p title="'.$_GET['v'].'">';`, phpversion.PHP85, 1},
		{"single quoted raw", `echo "<p title='".$_GET['v']."'>";`, phpversion.PHP85, 1},
		// Before PHP 7 the escape is text: the value is unquoted.
		{"unicode quote before PHP 7", `echo "<p title=\u{22}{$_GET['v']}>";`, phpversion.PHP56, 1},
		{"unicode quote from PHP 7", `echo "<p title=\u{22}{$_GET['v']}>";`, phpversion.PHP70, 1},
		{"interpolation", `echo "<p title=\"{$_GET['v']}\">";`, phpversion.PHP85, 1},
		{"unfinished literal value", `echo '<p title="unfinished >'.$_GET['v'].'">';`, phpversion.PHP85, 1},
		{"double compat", `echo '<p title="'.htmlspecialchars($_GET['v'],ENT_COMPAT).'">';`, phpversion.PHP85, 0},
		{"single compat", `echo "<p title='".htmlspecialchars($_GET['v'],ENT_COMPAT)."'>";`, phpversion.PHP85, 1},
		{"no quotes", `echo '<p title="'.htmlspecialchars($_GET['v'],ENT_NOQUOTES).'">';`, phpversion.PHP85, 1},
		{"default before 8.1", `echo "<p title='".htmlspecialchars($_GET['v'])."'>";`, phpversion.PHP74, 1},
		{"default from 8.1", `echo "<p title='".htmlspecialchars($_GET['v'])."'>";`, phpversion.PHP81, 0},
		{"named flag", `echo "<p title='".htmlspecialchars(flags:ENT_QUOTES,string:$_GET['v'])."'>";`, phpversion.PHP85, 0},
		{"escaped variable", `$v=htmlspecialchars($_GET['v'],ENT_QUOTES);echo "<p title='$v'>";`, phpversion.PHP85, 0},
		{"escaped wrapper", `function esc($x){return htmlspecialchars($x,ENT_QUOTES);}echo "<p title='".esc($_GET['v'])."'>";`, phpversion.PHP85, 0},
		{"conditional flags", `$flags=$c?ENT_QUOTES:ENT_COMPAT;echo "<p title='".htmlspecialchars($_GET['v'],$flags)."'>";`, phpversion.PHP85, 1},
		{"request element rewrite", `$_GET['v']=htmlspecialchars($_GET['v'],ENT_COMPAT);echo "<p title='".$_GET['v']."'>";`, phpversion.PHP85, 1},
		{"request element escaped", `$_GET['v']=htmlspecialchars($_GET['v'],ENT_QUOTES);echo "<p title='".$_GET['v']."'>";`, phpversion.PHP85, 0},
		{"escaped request then decoded", `$_GET['v']=htmlspecialchars($_GET['v'],ENT_QUOTES);echo "<p title='".htmlspecialchars_decode($_GET['v'],ENT_QUOTES)."'>";`, phpversion.PHP85, 1},
		{"encoded request then decoded", `$_GET['v']=rawurlencode($_GET['v']);echo "<p title='".urldecode($_GET['v'])."'>";`, phpversion.PHP85, 1},
		{"literal request then decoded", `$_GET['v']='safe';echo "<p title='".urldecode($_GET['v'])."'>";`, phpversion.PHP85, 0},
		{"reference escaped request then decoded", `function escRef(&$v){$v=htmlspecialchars($v,ENT_QUOTES);}escRef($_GET['v']);echo "<p title='".htmlspecialchars_decode($_GET['v'],ENT_QUOTES)."'>";`, phpversion.PHP85, 1},
		{"escaped request copied", `$_GET['v']=htmlspecialchars($_GET['v'],ENT_QUOTES);$v=$_GET['v'];echo "<p title='$v'>";`, phpversion.PHP85, 0},
		{"deep request read", `echo "<p title='" . $_GET` + strings.Repeat(`[0]`, 9) + ` . "'>";`, phpversion.PHP85, 1},
		{"arithmetic", `echo '<p title="'.($_GET['v']+1).'">';`, phpversion.PHP85, 0},
		{"URL encoding alphabet", `echo '<p title="'.rawurlencode($_GET['v']).'">';`, phpversion.PHP85, 0},
		{"hexadecimal", `echo '<p title="'.md5($_GET['v']).'">';`, phpversion.PHP85, 0},
		{"preserving transform", `echo "<p title='".strtoupper(htmlspecialchars($_GET['v'],ENT_QUOTES))."'>";`, phpversion.PHP85, 0},
		{"decoding transform", `echo "<p title='".htmlspecialchars_decode(htmlspecialchars($_GET['v'],ENT_QUOTES),ENT_QUOTES)."'>";`, phpversion.PHP85, 1},
		{"unknown transform", `echo "<p title='".str_replace('&#039;',"'",htmlspecialchars($_GET['v'],ENT_QUOTES))."'>";`, phpversion.PHP85, 1},
		{"comma echo", `echo '<p title="',$_GET['v'],'">';`, phpversion.PHP85, 1},
		{"print", `print '<p title="'.$_GET['v'].'">';`, phpversion.PHP85, 1},
		{"unknown prefix", `echo $prefix.'<p title="'.$_GET['v'].'">';`, phpversion.PHP85, 0},
		{"URL start", `echo '<a href="'.$_GET['v'].'">';`, phpversion.PHP85, 1},
		{"URL start escaped", `echo '<a href="'.htmlspecialchars($_GET['v'],ENT_QUOTES).'">';`, phpversion.PHP85, 1},
		{"URL start encoded", `echo '<a href="'.rawurlencode($_GET['v']).'">';`, phpversion.PHP85, 0},
		{"URL query escaped", `echo '<a href="/search?q='.htmlspecialchars($_GET['v'],ENT_QUOTES).'">';`, phpversion.PHP85, 0},
		{"event", `echo '<a onclick="'.$_GET['v'].'">';`, phpversion.PHP85, 1},
		{"event escaped", `echo '<a onclick="f(\''.htmlspecialchars($_GET['v'],ENT_QUOTES).'\')">';`, phpversion.PHP85, 1},
		{"event encoded as JSON then escaped", `echo '<a onclick="f('.htmlspecialchars(json_encode($_GET['v']),ENT_QUOTES).')">';`, phpversion.PHP85, 0},
		{"event number", `echo '<a onclick="f('.(int)$_GET['v'].')">';`, phpversion.PHP85, 0},
		{"event JSON in a string", `echo '<a onclick="f(&quot;'.htmlspecialchars(json_encode($_GET['v']),ENT_QUOTES).'&quot;)">';`, phpversion.PHP85, 1},
		{"script JSON in a string", `echo '<script>let a = "'.json_encode($_GET['v']).'";</script>';`, phpversion.PHP85, 1},
		{"script JSON", `echo '<script>let a = '.json_encode($_GET['v']).';</script>';`, phpversion.PHP85, 0},
		{"script URL escaped", `echo '<a href="javascript:'.htmlspecialchars($_GET['v'],ENT_QUOTES).'">';`, phpversion.PHP85, 1},
		{"style", `echo '<p style="'.htmlspecialchars($_GET['v'],ENT_QUOTES).'">';`, phpversion.PHP85, 1},
		{"unquoted", `echo '<p title='.htmlspecialchars($_GET['v'],ENT_QUOTES).'>';`, phpversion.PHP85, 1},
		{"unquoted encoded", `echo '<p title='.rawurlencode($_GET['v']).'>';`, phpversion.PHP85, 0},
		{"later side effect", `echo '<p title="'.$_GET['v'].audit().'">';`, phpversion.PHP85, 0},
		{"later echo operand", `echo '<p title="',$_GET['v'],audit(),'">';`, phpversion.PHP85, 1},
		{"plain response", `header('Content-Type: text/plain');echo '<p title="'.$_GET['v'].'">';`, phpversion.PHP85, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: tc.version, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("attribute.php", []byte("<?php "+tc.source), syntax.Options{Version: tc.version})
			if len(f.Errors) > 0 {
				t.Fatal(f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}

func TestHTMLContext(t *testing.T) {
	for _, tc := range []struct {
		prefix string
		want   flow.Context
	}{
		{`<p title="`, flow.HTMLDoubleQuoted},
		{`<p title='`, flow.HTMLSingleQuoted},
		{"<p TITLE \t=\n'", flow.HTMLSingleQuoted},
		{`<p data-value="`, flow.HTMLDoubleQuoted},
		{`<p aria-label="`, flow.HTMLDoubleQuoted},
		{`<p id="x">`, flow.HTML},
		{`<p title='>`, flow.HTMLSingleQuoted},
		{`</p title="`, 0},
		{`<p '`, 0},
		{`<p title"`, 0},
		{`<p ="`, 0},
		{`<title="`, 0},
		{`<p custom="`, 0},
		{`<p style="`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="`, flow.HTMLURL | flow.HTMLDoubleQuoted},
		{`<a href="/search?q=`, flow.HTMLDoubleQuoted},
		{`<a href='java`, flow.HTMLURL | flow.HTMLSingleQuoted},
		{`<a onclick="`, flow.HTMLScript | flow.HTMLDoubleQuoted},
		{`<iframe srcdoc='`, flow.HTMLRaw | flow.HTMLSingleQuoted},
		{`<p title=`, flow.HTMLUnquoted},
		{`<p title= `, flow.HTMLUnquoted},
		{`<p title=ab`, flow.HTMLUnquoted},
		{`<a href=`, flow.HTMLURL | flow.HTMLUnquoted},
		{`<p custom=`, 0},
		{`<p title=ab `, 0},
		{`<p =`, 0},
		{`<script>let a = `, flow.HTMLScript},
		{`<script type="module">`, flow.HTMLScript},
		{`<style>p { color: `, flow.HTMLRaw},
		{`<script src="`, 0},
		{`<script>x</script><p>`, 0},
		{`<script>let a = "`, flow.HTMLRaw},
		{`<script>let a = '\'' + `, flow.HTMLScript},
		{"<script>let a = `", flow.HTMLRaw},
		{`<script>// note`, flow.HTMLRaw},
		{"<script>// note\nlet a = ", flow.HTMLScript},
		{`<script>/* note */ let a = `, flow.HTMLScript},
		{`<script>let a = 1 / 2 + `, flow.HTMLScript},
		{`İİİ<script>let a = `, flow.HTMLScript},
		{`<a onclick="f(&#39;`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="javascript:`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href=" JavaScript:go(`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="java&#9;script:`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="javascript&#58;`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="javascript&#5`, flow.HTMLURL | flow.HTMLDoubleQuoted},
		{`<a href="data:text/html,`, flow.HTMLRaw | flow.HTMLDoubleQuoted},
		{`<a href="https://example.com/?q=`, flow.HTMLDoubleQuoted},
		{`<a href="mailto:`, flow.HTMLDoubleQuoted},
		{`<a href="#`, flow.HTMLDoubleQuoted},
		{`<p title='x'`, 0},
	} {
		if got := htmlContext(tc.prefix); got != tc.want {
			t.Errorf("%q: got %v want %v", tc.prefix, got, tc.want)
		}
	}
}

func BenchmarkHTMLAttributeContext(b *testing.B) {
	prefix := strings.Repeat(`<span title="static">text</span>`, 32) + `<input aria-label="`
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		htmlContext(prefix)
	}
}

func TestMalformedAttributeFragment(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	f := syntax.ParseBest("fragment.php", []byte(`<?php echo "<p title=\u{110000}{$_GET['v']}>";`), syntax.Options{Version: phpversion.PHP85})
	if got := e.Analyze(f); len(got) != 0 {
		t.Fatalf("invented a context from an invalid escape: %+v", got)
	}
}
