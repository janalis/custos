package unpackbuffertooshort

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestVersion155(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnpackBufferTooShort"}, PHP: phpversion.MustParse("5.3")})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("test.php", []byte("<?php unpack('Jvalue','x');"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("unsupported contract: %+v", got)
	}
}

func TestVersion508(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"UnpackBufferTooShort"}, PHP: phpversion.MustParse("5.3")})
	if err != nil {
		t.Fatal(err)
	}
	got := e.Analyze(syntax.Parse("test.php", []byte("<?php unpack('Nvalue','xxxx',1);"), syntax.Options{}))
	if len(got) != 0 {
		t.Fatalf("unsupported contract: %+v", got)
	}
}
