package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
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
	f := syntax.Parse("enum.php", []byte(`<?php namespace App; enum Flag { case Ready; } Flag::cases()[0];`), syntax.Options{Version: phpversion.PHP81})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	if !ix.IsSubtype(`App\Flag`, "UnitEnum", phpversion.PHP81) || ix.IsSubtype(`App\Flag`, "BackedEnum", phpversion.PHP81) {
		t.Fatal("enum runtime interfaces")
	}
	env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP81).Native()
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
`, map[string]string{"self": `\Model\Base|null`, "parent": `\Model\Root|null`, "trait": `\Model\UsesRelativeTrait|null`, "broken": "?unknown"})
}

func TestTypedTraitConstantOwnersAcrossFiles(t *testing.T) {
	checkWith(t, map[string]string{"traits.php": `<?php namespace Model;
trait Relative {
 public const self|null SELF = null;
 public const parent|null PARENT = null;
 public const int SIZE = UNKNOWN_SIZE;
}
trait Nested { use Relative; }
class Root {}
class OtherRoot {}
class First extends Root { use Nested; }
class Second extends OtherRoot { use Relative; }
class Child extends First {}
class Orphan { use Relative; }
`}, `<?php
t('first', Model\First::SELF); t('firstParent', Model\First::PARENT);
t('second', Model\Second::SELF); t('secondParent', Model\Second::PARENT);
t('child', Model\Child::SELF); t('childParent', Model\Child::PARENT);
t('orphan', Model\Orphan::SELF); t('orphanParent', Model\Orphan::PARENT);
t('rawTrait', Model\Relative::SELF); t('size', Model\Child::SIZE);
`, map[string]string{
		"first": `\Model\First|null`, "firstParent": `\Model\Root|null`,
		"second": `\Model\Second|null`, "secondParent": `\Model\OtherRoot|null`,
		"child": `\Model\First|null`, "childParent": `\Model\Root|null`,
		"orphan": `\Model\Orphan|null`, "orphanParent": "?unknown",
		"rawTrait": "?unknown", "size": "int",
	})
}

func BenchmarkTypedConstantInference(b *testing.B) {
	f := syntax.Parse("constant.php", []byte(`<?php class C { public const int SIZE = OTHER; } C::SIZE;`), syntax.Options{Version: phpversion.PHP85})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	resolver := names.New(f)
	expr := f.Stmts[1].(*syntax.ExprStmt).Expr
	b.ReportAllocs()
	for b.Loop() {
		infer.NewEnv(f, resolver, ix, phpversion.PHP85).TypeOf(expr)
	}
}
