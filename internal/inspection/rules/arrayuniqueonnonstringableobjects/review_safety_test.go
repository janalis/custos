package arrayuniqueonnonstringableobjects

import (
	"strconv"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// wideClass declares a class implementing more interfaces than the index
// enumerates, so its hierarchy is reported incomplete.
func wideClass() string {
	var decl, names strings.Builder
	for i := range 300 {
		name := "I" + strconv.Itoa(i)
		decl.WriteString("interface " + name + " {} ")
		if i > 0 {
			names.WriteString(",")
		}
		names.WriteString(name)
	}
	return decl.String() + "class Wide implements " + names.String() + " {} array_unique([new Wide,new Wide]);"
}

// TestReviewSafety covers string-conversion proof for internal and inherited classes.
func TestReviewSafety(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		want      int
	}{
		{"gmp cast handler", "array_unique([new GMP(5),new GMP(5),new GMP(7)]);", 0},
		{"gmp explicit flag", "array_unique([new GMP(5),new GMP(5)],SORT_STRING);", 0},
		{"internal outside list", "array_unique([new SplTempFileObject(),new SplTempFileObject()]);", 0},
		{"internal with method", "array_unique([new Exception('a'),new Exception('a')]);", 0},
		{"unknown parent", "class C extends Missing {} array_unique([new C,new C]);", 0},
		{"unknown trait", "class C { use Missing; } array_unique([new C,new C]);", 0},
		{"unknown grandparent", "class B extends Missing {} class C extends B {} array_unique([new C,new C]);", 0},
		{"inherited method", "class B { function __toString(){return 'b';} } class C extends B {} array_unique([new C,new C]);", 0},
		{"trait method", "trait T { function __toString(){return 't';} } class C { use T; } array_unique([new C,new C]);", 0},
		{"incomplete hierarchy", wideClass(), 0},
		{"user class", "class C {} array_unique([new C,new C]);", 1},
		{"user interface", "interface I {} class C implements I {} array_unique([new C,new C]);", 1},
		{"extends listed builtin", "class C extends ArrayObject {} array_unique([new C,new C]);", 1},
		{"listed builtin", "array_unique([new DateTimeImmutable(),new DateTimeImmutable()]);", 1},
		{"case-insensitive builtin", "array_unique([new STDCLASS,new stdclass]);", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("review.php", []byte("<?php "+tc.src), syntax.Options{})
			if len(f.Errors) != 0 {
				t.Fatalf("invalid probe: %+v", f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.want {
				t.Fatalf("got %d findings, want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
