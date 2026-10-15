package phpdoc

import "testing"

func TestRound19TagsAndExamples(t *testing.T) {
	d := Parse("/**\n * <code>\n * @return false\n * </code>\n * ```php\n * @param int $bad\n * ~~~\n * @var bool\n * ```\n * @ReTuRn string\n * @TEMPLATE T\n */")
	if d.ReturnType() != "string" || d.Has("param") || len(d.All("RETURN")) != 1 || len(d.Templates()) != 1 || len(d.TemplateParams()) != 1 {
		t.Fatalf("parsed: %+v", d)
	}
	manual := &Doc{Tags: []Tag{{Name: "RETURN", Text: "int"}}}
	if !manual.Has("return") || len(manual.All("return")) != 1 {
		t.Fatal("manual tags must also ignore case")
	}
}

func TestCodeBlockDelimiters(t *testing.T) {
	for _, src := range []string{
		"/**\n * ````php\n * ```\n * @return bad\n * `````\n * @return int\n */",
		"/**\n * ~~~~php\n * ~~~\n * @return bad\n * ~~~~\n * @return int\n */",
		"/**\n * <codec>\n * @return int\n */",
		"/**\n * <code class='php'>\n * @return bad\n * </code>\n * @return int\n */",
	} {
		if got := Parse(src).ReturnType(); got != "int" {
			t.Errorf("%s: %s", src, got)
		}
	}
}

func TestCaseInsensitiveLookupKeepsAnnotationIdentity(t *testing.T) {
	d := Parse("/**\n * @Vendor\\Entity\n * @RETURN int\n * @PHPSTAN-PARAM string $value\n * @PHPSTAN-VAR bool $flag\n */")
	if d.Tags[0].Name != `Vendor\Entity` || d.EffectiveReturnType() != "int" || d.EffectiveParams()[0].Type != "string" || d.EffectiveVarType("flag") != "bool" {
		t.Fatalf("%+v", d)
	}
}
