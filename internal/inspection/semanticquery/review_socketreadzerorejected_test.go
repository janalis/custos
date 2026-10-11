package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// reviewSocketReadProbe dispatches on calls only, as the rule package does.
type reviewSocketReadProbe struct{ nativeCheckProbe }

func (reviewSocketReadProbe) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func TestReviewSocketReadZeroRejected(t *testing.T) {
	const fail = "){throw new RuntimeException();}"
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"explicit binary mode", "if(!socket_read($s,10,PHP_BINARY_READ)" + fail, 1},
		{"omitted mode", "if(!socket_read($s,10)" + fail, 1},
		{"omitted mode assigned", "if(!($d=socket_read($s,10))){return false;}", 1},
		{"omitted mode loose false", "if(socket_read($s,10)==false){return false;}", 1},
		{"omitted mode named arguments", "if(!socket_read(length:10,socket:$s)" + fail, 1},
		{"named binary mode", "if(!socket_read($s,10,mode:PHP_BINARY_READ)" + fail, 1},
		{"named normal mode", "if(!socket_read($s,10,mode:PHP_NORMAL_READ)" + fail, 0},
		{"named mode first", "if(!socket_read(mode:PHP_NORMAL_READ,socket:$s,length:10)" + fail, 0},
		{"normal mode", "if(!socket_read($s,10,PHP_NORMAL_READ)" + fail, 0},
		{"dynamic mode", "if(!socket_read($s,10,$mode)" + fail, 0},
		{"spread arguments", "if(!socket_read(...$args)" + fail, 0},
		{"spread after positional", "if(!socket_read($s,...$rest)" + fail, 0},
		{"first-class callable", "if(!socket_read(...)" + fail, 0},
		{"omitted mode strict false", "if(socket_read($s,10)===false){return false;}", 0},
		{"omitted mode no rejection", "if(!socket_read($s,10)){echo 'empty';}", 0},
		{"user function", "namespace A;function socket_read($s,$n){return '';}if(!socket_read($s,10)" + fail, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := reviewSocketReadProbe{nativeCheckProbe{"SocketReadZeroRejected", CheckSocketReadZeroRejected}}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{"SocketReadZeroRejected"}})
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
