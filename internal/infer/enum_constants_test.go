package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestEnumRuntimeInference(t *testing.T) {
	checkWith(t, map[string]string{"enums.php": `<?php namespace Model;
enum Flag { case Ready; }
enum Code: int { case Ready = 1; }
enum Label: string { case Ready = 'ready'; }
`}, `<?php use Model\Flag as State; use Model\Code; use Model\Label;
function run(Code $code, Label $label) {
 t('name', $code->name); t('intValue', $code->value); t('stringValue', $label->value);
 t('from', Code::from(1)); t('tryFrom', Code::tryFrom(2));
 foreach (State::cases() as $state) { t('case', $state); }
 t('element', Code::cases()[0]);
}`, map[string]string{"name": "string", "intValue": "int", "stringValue": "string", "from": `\Model\Code`, "tryFrom": `\Model\Code|null`, "case": `\Model\Flag`, "element": `\Model\Code`})
}

func TestEnumCasesNativeInference(t *testing.T) {
	f := syntax.Parse("enum.php", []byte(`<?php namespace App; enum Flag { case Ready; } Flag::cases()[0];`), syntax.Options{Version: phpver.PHP81})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !ix.IsSubtype(`App\Flag`, "UnitEnum", phpver.PHP81) || ix.IsSubtype(`App\Flag`, "BackedEnum", phpver.PHP81) {
		t.Fatal("enum runtime interfaces")
	}
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP81).Native()
	var stmt *syntax.ExprStmt
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if s, ok := n.(*syntax.ExprStmt); ok {
			stmt = s
		}
		return true
	})
	if got := env.TypeOf(stmt.Expr).String(); got != `\App\Flag` {
		t.Fatalf("native enum element: %s", got)
	}
}

func TestTypedConstantInferenceAcrossFiles(t *testing.T) {
	checkWith(t, map[string]string{"constants.php": `<?php
namespace Model; use Vendor\Item as Product;
class Base {
 public const int COUNT = UNKNOWN_COUNT;
 public const Product|null PRODUCT = null;
 public const int|string KEY = UNKNOWN_KEY;
 public const LEGACY = 'legacy';
}
class Child extends Base {}
`}, `<?php use Model\Child as Constants;
t('count', Constants::COUNT); t('product', Constants::PRODUCT);
t('key', Constants::KEY); t('legacy', Constants::LEGACY);
`, map[string]string{"count": "int", "product": `\Vendor\Item|null`, "key": "int|string", "legacy": "string"})
}

func TestTypedConstantRelativeTypes(t *testing.T) {
	checkWith(t, map[string]string{"relative.php": `<?php namespace Model;
class Root {}
class Base extends Root {
 public const self|null SELF = null;
 public const parent|null PARENT = null;
}
class Child extends Base {}
trait RelativeTrait { public const self|null VALUE = null; }
class UsesRelativeTrait { use RelativeTrait; }
class Broken { public const parent|null VALUE = null; }
`}, `<?php
t('self', Model\Child::SELF); t('parent', Model\Child::PARENT);
t('trait', Model\UsesRelativeTrait::VALUE); t('broken', Model\Broken::VALUE);
`, map[string]string{"self": `\Model\Base|null`, "parent": `\Model\Root|null`, "trait": "?unknown", "broken": "?unknown"})
}

func BenchmarkTypedConstantInference(b *testing.B) {
	f := syntax.Parse("constant.php", []byte(`<?php class C { public const int SIZE = OTHER; } C::SIZE;`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	resolver := names.New(f)
	expr := f.Stmts[1].(*syntax.ExprStmt).Expr
	b.ReportAllocs()
	for b.Loop() {
		infer.NewEnv(f, resolver, ix, phpver.PHP85).TypeOf(expr)
	}
}
