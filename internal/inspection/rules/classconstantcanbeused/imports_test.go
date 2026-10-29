package classconstantcanbeused

import (
	"testing"

	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// Imports that cannot be inserted next to the ones of the same pass wait for
// the next one: the iterative fix loop imports every class.
func TestImportsAcrossPasses(t *testing.T) {
	src := "<?php\nnamespace Lib\\Tools;\nclass Hammer {}\nclass Saw {}\nclass Drill {}\nclass Level {}\n\n" +
		"namespace App;\n\n$list = ['Lib\\Tools\\Hammer', 'Lib\\Tools\\Saw', 'Lib\\Tools\\Drill', 'Lib\\Tools\\Level', 'Lib\\Tools\\Drill'];\n"
	want := "<?php\nnamespace Lib\\Tools;\nclass Hammer {}\nclass Saw {}\nclass Drill {}\nclass Level {}\n\n" +
		"namespace App;\n\nuse Lib\\Tools\\Hammer;\nuse Lib\\Tools\\Saw;\nuse Lib\\Tools\\Drill;\nuse Lib\\Tools\\Level;\n\n" +
		"$list = [Hammer::class, Saw::class, Drill::class, Level::class, Drill::class];\n"
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{
		PHP:       phpversion.PHP84,
		EnableAll: true,
		Rules:     map[string]analysis.RuleConfig{"ClassConstantCanBeUsed": {Options: map[string]any{"IMPORT_CLASSES_ON_QF": true}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	res := fixing.FixSource(e, "imports.php", []byte(src), fixing.Options{Parse: syntax.Options{Version: phpversion.PHP84}})
	if string(res.Source) != want {
		t.Fatalf("fixed source:\n%s\nwant:\n%s", res.Source, want)
	}
}
