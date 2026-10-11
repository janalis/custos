package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewTLSHandshakeProbe dispatches on calls only, as the rule package does.
type reviewTLSHandshakeProbe struct{ nativeCheckProbe }

func (reviewTLSHandshakeProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func TestReviewTLSHandshakePendingAccepted(t *testing.T) {
	const (
		open  = "function f($s,$o){if(!stream_set_blocking($s,false)){return;}if(stream_socket_enable_crypto($s,true)!==false)"
		again = "stream_socket_enable_crypto($s,true)"
	)
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"direct write", "{fwrite($s,'x');}", 1},
		{"unbraced write", "fwrite($s,'x');", 1},
		{"blocking alone", "{stream_set_blocking($s,true);fwrite($s,'x');}", 1},
		{"write before guard", "{fwrite($s,'x');if(" + again + "!==true){return;}}", 1},
		{"guard does not leave", "{if(" + again + "!==true){echo 'pending';}fwrite($s,'x');}", 1},
		{"guard accepts pending", "{if(" + again + "===false){return;}fwrite($s,'x');}", 1},
		{"guard compares to false", "{if(" + again + "!==false){return;}fwrite($s,'x');}", 1},
		{"guard on other stream", "{if(stream_socket_enable_crypto($o,true)!==true){return;}fwrite($s,'x');}", 1},
		{"guard disables crypto", "{if(stream_socket_enable_crypto($s,false)!==true){return;}fwrite($s,'x');}", 1},
		{"guard with unknown toggle", "{if(stream_socket_enable_crypto($s,$o)!==true){return;}fwrite($s,'x');}", 1},
		{"guard tests other call", "{if(feof($s)!==true){return;}fwrite($s,'x');}", 1},
		{"guard tests variable", "{if(!$o){return;}fwrite($s,'x');}", 1},
		{"guard negates arithmetic", "{if(-" + again + "){return;}fwrite($s,'x');}", 1},
		{"guard with other condition", "{if($o){return;}fwrite($s,'x');}", 1},
		{"guard compound assignment", "{if(($o.=" + again + ")!==true){return;}fwrite($s,'x');}", 1},
		{"strict guard", "{stream_set_blocking($s,true);if(" + again + "!==true){return;}fwrite($s,'x');}", 0},
		{"loose guard", "{if(" + again + "!=true){return;}fwrite($s,'x');}", 0},
		{"negated guard", "{if(!" + again + "){throw new RuntimeException();}fwrite($s,'x');}", 0},
		{"assigned guard", "{if(($r=" + again + ")!==true){return;}fwrite($s,'x');}", 0},
		{"negated assigned guard", "{if(!($r=" + again + ")){return;}fwrite($s,'x');}", 0},
		{"unbraced guard", "{if(" + again + "!==true)return;fwrite($s,'x');}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewTLSHandshakeProbe{nativeCheckProbe{"TlsHandshakePendingAccepted", CheckTLSHandshakePendingAccepted}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{"TlsHandshakePendingAccepted"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("review.php", []byte("<?php "+open+tc.body+"}"), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %+v, want %d", got, tc.want)
			}
		})
	}
}
