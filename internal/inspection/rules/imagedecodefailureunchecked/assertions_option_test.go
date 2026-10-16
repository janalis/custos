package imagedecodefailureunchecked

import (
	"fmt"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

func TestAssertionAssumption(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, description := range []string{"", ", $missing", ", rewrite()", ", 'message'"} {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ImageDecodeFailureUnchecked"}, Flow: flow.Options{Assertions: enabled}})
			if err != nil {
				t.Fatal(err)
			}
			src := fmt.Sprintf(`<?php function f($bytes){$im=imagecreatefromstring($bytes);assert($im!==false%s);imagepng($im);}`, description)
			got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
			want := 1
			if enabled && (description == "" || description == ", 'message'") {
				want = 0
			}
			if len(got) != want {
				t.Fatalf("enabled=%v description=%s findings=%v", enabled, description, got)
			}
		}
	}
}

func TestAssertionCompositeDecodeGuard(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Flow: flow.Options{Assertions: enabled}})
		if err != nil {
			t.Fatal(err)
		}
		src := `<?php function f($bytes,$ready){$image=imagecreatefromstring($bytes);assert(($image||is_object($image))&&$ready);imagepng($image);}`
		got := e.Analyze(syntax.Parse("test.php", []byte(src), syntax.Options{}))
		want := 1
		if enabled {
			want = 0
		}
		if len(got) != want {
			t.Fatalf("enabled=%v findings=%v", enabled, got)
		}
	}
}
