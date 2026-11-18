package jsondecodeknowninvalidutf8

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{"<?php\njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n", 1},
		{"<?php\njson_decode('{}', true, 512, JSON_THROW_ON_ERROR);\n", 0},
		{"<?php other();", 0},
		{"<?php json_decode(\"\\xFF\",true,512,$flags);", 0},
		{"<?php json_decode(\"\\xFF\",true,512,JSON_INVALID_UTF8_SUBSTITUTE);", 0},
		{"<?php json_decode(\"\\xFF\");", 1},
		{"<?php function unreachable(){return;\njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n}", 0},
		{"<?php \njson_decode(\"\\\"\\xFF\\\"\", true, 512, JSON_THROW_ON_ERROR);\n function broken(", 0},
	} {
		e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"JsonDecodeKnownInvalidUtf8"}})
		if err != nil {
			t.Fatal(err)
		}
		got := e.Analyze(syntax.Parse("test.php", []byte(tc.src), syntax.Options{}))
		if len(got) != tc.want {
			t.Errorf("%s: got %d findings, want %d: %+v", tc.src, len(got), tc.want, got)
		}
	}
}
