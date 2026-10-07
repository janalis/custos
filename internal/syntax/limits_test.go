package syntax

import (
	"strings"
	"testing"

	"custos/internal/phpver"
)

func TestParseLimits(t *testing.T) {
	// garbage: an error per token, capped
	f := ParseBest("g.php", []byte("<?php "+strings.Repeat(")]}", 100000)), Options{Version: phpver.PHP84})
	if len(f.Errors) != MaxErrors+1 || !strings.Contains(f.Errors[MaxErrors].Msg, "more than") {
		t.Fatalf("garbage: %d errors, last %q", len(f.Errors), f.Errors[len(f.Errors)-1].Msg)
	}
	// too large: one error, no tree, no tokens
	big := make([]byte, MaxFileSize+1)
	copy(big, "<?php ")
	for i := 6; i < len(big); i++ {
		big[i] = ';'
	}
	f = ParseBest("big.php", big, Options{Version: phpver.PHP84})
	if len(f.Errors) != 1 || !strings.Contains(f.Errors[0].Msg, "larger than") || len(f.Stmts) != 0 || len(f.Tokens) != 0 {
		t.Fatalf("too large: %d errors, %d stmts", len(f.Errors), len(f.Stmts))
	}
	// the nesting report survives the error cap
	deep := "<?php " + strings.Repeat(")", 2000) + strings.Repeat("[", 200000)
	f = ParseBest("deep.php", []byte(deep), Options{Version: phpver.PHP84})
	if last := f.Errors[len(f.Errors)-1].Msg; !strings.Contains(last, "nesting deeper") {
		t.Fatalf("deep: last error %q", last)
	}
}

// TestTreeDepthWide checks the depth walk on wide and deep trees.
func TestTreeDepthWide(t *testing.T) {
	f := Parse("w.php", []byte("<?php "+strings.Repeat("$a;", 200000)), Options{Version: phpver.PHP84})
	if d := TreeDepth(f.Stmts); d != 2 {
		t.Fatalf("wide depth %d", d)
	}
	f = Parse("d.php", []byte("<?php $a = "+strings.Repeat("[", 1000)+strings.Repeat("]", 1000)+";"), Options{Version: phpver.PHP84})
	if d := TreeDepth(f.Stmts); d < 1000 || d > MaxDepth {
		t.Fatalf("deep depth %d", d)
	}
}
