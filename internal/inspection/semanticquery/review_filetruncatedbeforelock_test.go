package semanticquery

import "testing"

func TestReviewFileTruncatedBeforeLock(t *testing.T) {
	const open = "function f($p,$op){$h=fopen($p,'w');"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"exclusive", open + "flock($h,LOCK_EX);}", 1},
		{"qualified exclusive", open + "\\flock($h,\\LOCK_EX);}", 1},
		{"exclusive non-blocking", open + "flock($h,LOCK_EX|LOCK_NB);}", 1},
		{"parenthesized non-blocking", open + "flock($h,(LOCK_NB|LOCK_EX));}", 1},
		{"integer exclusive", open + "flock($h,2);}", 1},
		{"integer exclusive non-blocking", open + "flock($h,6);}", 1},
		{"named operation", open + "flock(operation:LOCK_EX,stream:$h);}", 1},
		{"local operation", open + "$mode=LOCK_EX|LOCK_NB;flock($h,$mode);}", 1},
		{"exclusive after unlock", open + "flock($h,LOCK_UN);flock($h,LOCK_EX);}", 1},
		{"unlock only", open + "fwrite($h,'x');flock($h,LOCK_UN);}", 0},
		{"unlock non-blocking", open + "flock($h,LOCK_UN|LOCK_NB);}", 0},
		{"shared only", open + "flock($h,LOCK_SH);}", 0},
		{"shared non-blocking", open + "flock($h,LOCK_SH|LOCK_NB);}", 0},
		{"integer shared", open + "flock($h,1);}", 0},
		{"integer unlock", open + "flock($h,3);}", 0},
		{"dynamic operation", open + "flock($h,$op);}", 0},
		{"partly dynamic operation", open + "flock($h,LOCK_EX|$op);}", 0},
		{"other operator", open + "flock($h,LOCK_EX&$op);}", 0},
		{"missing operation", open + "flock($h);}", 0},
		{"other handle", open + "$g=fopen($p,'c');flock($g,LOCK_EX);}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := auditNative(t, "FileTruncatedBeforeLock", CheckFileTruncatedBeforeLock, tc.src)
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
