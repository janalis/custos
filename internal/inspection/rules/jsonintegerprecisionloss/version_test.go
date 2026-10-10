package jsonintegerprecisionloss

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestFixRequiresDecodeFlags(t *testing.T) {
	for _, version := range []phpversion.Version{phpversion.PHP53, phpversion.PHP54} {
		t.Run(version.String(), func(t *testing.T) {
			engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: version, Only: []string{"JsonIntegerPrecisionLoss"}})
			if err != nil {
				t.Fatal(err)
			}
			file := syntax.Parse("test.php", []byte(`<?php json_decode('{"id":9223372036854775808}');`), syntax.Options{Version: version})
			findings := engine.Analyze(file)
			if len(findings) != 1 {
				t.Fatalf("expected precision diagnostic: %+v", findings)
			}
			wantFix := version >= phpversion.PHP54
			if (len(findings[0].Fixes) != 0) != wantFix {
				t.Fatalf("fix availability for PHP %s: %+v", version, findings)
			}
		})
	}
}
