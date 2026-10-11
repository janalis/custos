package semanticquery

import "testing"

func TestReviewContentLengthUsesCharacterCount(t *testing.T) {
	const head = "function f($body){header('Content-Length: '."
	for _, tc := range []struct {
		name, src string
		want      int
		fixes     int
	}{
		{"8bit", head + "mb_strlen($body,'8bit'));echo $body;}", 0, 0},
		{"binary mixed case", head + "mb_strlen($body,'Binary'));echo $body;}", 0, 0},
		{"ascii", head + "mb_strlen($body,'ASCII'));echo $body;}", 0, 0},
		{"us-ascii", head + "mb_strlen($body,'us-ascii'));echo $body;}", 0, 0},
		{"latin1", head + "mb_strlen($body,'latin1'));echo $body;}", 0, 0},
		{"iso-8859-15", head + "\\mb_strlen($body,'ISO-8859-15'));echo $body;}", 0, 0},
		{"windows-1252", head + "mb_strlen($body,'Windows-1252'));echo $body;}", 0, 0},
		{"cp1251", head + "mb_strlen($body,'CP1251'));echo $body;}", 0, 0},
		{"named byte encoding", head + "mb_strlen(encoding:'8bit',string:$body));echo $body;}", 0, 0},
		{"constant byte encoding", "const ENC='8bit';" + head + "mb_strlen($body,ENC));echo $body;}", 0, 0},
		{"multibyte encoding stays reported", head + "mb_strlen($body,'UTF-16'));echo $body;}", 1, 1},
		{"sjis stays reported without fix", head + "mb_strlen($body,'SJIS'));echo $body;}", 1, 0},
		{"unknown iso part stays reported", head + "mb_strlen($body,'ISO-8859-12'));echo $body;}", 1, 0},
		{"dynamic encoding stays reported", "function f($body,$e){header('Content-Length: '.mb_strlen($body,$e));echo $body;}", 1, 0},
		{"byte strlen", head + "strlen($body));echo $body;}", 0, 0},
		{"omitted encoding", head + "mb_strlen($body));echo $body;}", 1, 1},
		{"grapheme length", head + "grapheme_strlen($body));echo $body;}", 1, 1},
		{"inline html after echo", head + "mb_strlen($body));echo $body; ?>\n<!-- served -->\n<?php }", 0, 0},
		{"inline html before echo", head + "mb_strlen($body)); ?> lead <?php echo $body;}", 0, 0},
		{"bare tag pair emits nothing", head + "mb_strlen($body));echo $body; ?>\n<?php }", 1, 1},
		{"foreach repeats echo", head + "mb_strlen($body));foreach([1,2] as $_){echo $body;}}", 0, 0},
		{"for repeats echo", head + "mb_strlen($body));for($i=0;$i<2;$i++){echo $body;}}", 0, 0},
		{"while repeats echo", head + "mb_strlen($body));while($body){echo $body;}}", 0, 0},
		{"do while repeats echo", head + "mb_strlen($body));do{echo $body;}while($body);}", 0, 0},
		{"enclosing loop outside function", "foreach([1] as $_){$g=function($body){header('Content-Length: '.mb_strlen($body));echo $body;};}", 1, 1},
		{"conditional echo", "function f($body,$c){header('Content-Length: '.mb_strlen($body));if($c){echo $body;}}", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := auditNative(t, "ContentLengthUsesCharacterCount", CheckContentLengthUsesCharacterCount, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
			if tc.want == 1 && len(got[0].Fixes) != tc.fixes {
				t.Fatalf("got %d fixes, want %d", len(got[0].Fixes), tc.fixes)
			}
		})
	}
}
