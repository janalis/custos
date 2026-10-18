package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewSocketCallProbe dispatches on calls only, as the rule package does.
type reviewSocketCallProbe struct{ nativeCheckProbe }

func (reviewSocketCallProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func TestReviewNonblockingSocketConnectPendingRejected(t *testing.T) {
	const on = "socket_set_nonblock($s);"
	const fail = "if(!socket_connect($s,$h,80)){return false;}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"unconditional setup", on + fail, 1},
		{"unconditional setup in function", "function f($s,$h){" + on + fail + "}", 1},
		{"unconditional setup throwing", on + "if(!socket_connect($s,$h,80)){throw new RuntimeException();}", 1},
		{"assigned setup result", "$ok=socket_set_nonblock($s);" + fail, 1},
		{"silenced connect", on + "if(!@socket_connect($s,$h,80)){return false;}", 1},
		{"silenced connect without negation", on + "if(@socket_connect($s,$h,80)){return false;}", 0},
		{"loose false", on + "if(socket_connect($s,$h,80)==false){return false;}", 1},
		{"checked setup", "if(!socket_set_nonblock($s)){return false;}" + fail, 1},
		{"conditional setup", "if($x){socket_set_nonblock($s);}" + fail, 0},
		{"short-circuit setup", "$x&&socket_set_nonblock($s);" + fail, 0},
		{"blocking restored", on + "socket_set_block($s);" + fail, 0},
		{"nonblocking restored", "socket_set_block($s);" + on + fail, 1},
		{"other socket", "socket_set_nonblock($t);" + fail, 0},
		{"socket reassigned", on + "$s=socket_create(AF_INET,SOCK_STREAM,0);" + fail, 0},
		{"unknown call on socket", on + "reset_socket($s);" + fail, 0},
		{"no setup", fail, 0},
		{"last error inspected", on + "if(!socket_connect($s,$h,80)){if(socket_last_error($s)!==SOCKET_EINPROGRESS){return false;}}", 0},
		{"socket option inspected", on + "if(!socket_connect($s,$h,80)){if(socket_get_option($s,SOL_SOCKET,SO_ERROR)!==0){return false;}}", 0},
		{"socket option alias inspected", on + "if(!socket_connect($s,$h,80)){if(socket_getopt($s,SOL_SOCKET,SO_ERROR)!==0){return false;}}", 0},
		{"helper handles socket before rejection", on + "if(!socket_connect($s,$h,80)){pending($s);return false;}", 0},
		{"helper receives socket", on + "if(!socket_connect($s,$h,80)){if(pending($s)){return true;}return false;}", 0},
		{"method receives socket", on + "if(!socket_connect($s,$h,80)){if($this->pending($s)){return true;}return false;}", 0},
		{"static method receives socket", on + "if(!socket_connect($s,$h,80)){if(Net::pending($s)){return true;}return false;}", 0},
		{"dynamic callee receives socket", on + "if(!socket_connect($s,$h,80)){if($check($s)){return true;}return false;}", 0},
		{"socket is the callee", on + "if(!socket_connect($s,$h,80)){if($s()){return true;}return false;}", 0},
		{"helper without socket", on + "if(!socket_connect($s,$h,80)){log_failure($h);return false;}", 1},
		{"method without socket", on + "if(!socket_connect($s,$h,80)){$this->log($h);return false;}", 1},
		{"static method without socket", on + "if(!socket_connect($s,$h,80)){Log::write($h);return false;}", 1},
		{"builtin receives socket", on + "if(!socket_connect($s,$h,80)){socket_close($s);return false;}", 1},
		{"helper in closure", on + "if(!socket_connect($s,$h,80)){$f=function()use($s){return pending($s);};return false;}", 1},
		{"property socket receiver method", "socket_set_nonblock($this->s);if(!socket_connect($this->s,$h,80)){if($this->pending()){return true;}return false;}", 0},
		{"property socket is not tracked", "socket_set_nonblock($this->s);if(!socket_connect($this->s,$h,80)){log_failure($h);return false;}", 0},
		{"local alias of property socket", "$s=$this->s;socket_set_nonblock($s);if(!socket_connect($s,$h,80)){$this->log($h);return false;}", 1},
		{"static property socket is not tracked", "if(socket_set_nonblock(self::$s)){if(!socket_connect(self::$s,$h,80)){log_failure($h);return false;}}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewSocketCallProbe{nativeCheckProbe{"NonblockingSocketConnectPendingRejected", CheckNonblockingSocketConnectPendingRejected}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
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
