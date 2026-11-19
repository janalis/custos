package index

import (
	"testing"

	phpversion "custos/internal/php/version"
)

const attrHeader = `<?php
namespace App;
use JetBrains\PhpStorm\ArrayShape;
use JetBrains\PhpStorm\ObjectShape;
use JetBrains\PhpStorm\ExpectedValues;
use JetBrains\PhpStorm\NoReturn;
`

func TestShapeAttributesOnReturns(t *testing.T) {
	fs := extract(t, "shapes.php", attrHeader+`
#[ArrayShape(['id' => 'int', 'tags' => 'string[]', 'owner' => User::class, 'x' => FOO])]
function plain() {}
#[ArrayShape(shape: ['id' => 'int'])]
function native(): ?array {}
#[ArrayShape(['id' => 'int'])]
function falsy(): array|false {}
/** @return array{id: string} */
#[ArrayShape(['id' => 'int'])]
function tagged(): array {}
/** @return array */
#[ArrayShape(['id' => 'int'])]
function vague(): array {}
#[ArrayShape(['id' => 'int'])]
function scalar(): int {}
#[ArrayShape(['id' => ''])]
function emptyValue() {}
#[ArrayShape([$k => 'int'])]
function badKey() {}
#[ArrayShape(['int'])]
function listForm() {}
#[ArrayShape(FOO)]
function notArray() {}
#[ArrayShape]
function noArgs() {}
#[ArrayShape(...)]
function placeholder() {}
#[ObjectShape(['a' => 'int'])]
function obj(): object {}
#[ObjectShape([0 => 'int'])]
function objInt(): object {}
#[\Other\ArrayShape(['id' => 'int'])]
function lookalike() {}
`)
	want := map[string]string{
		`App\plain`:       `array{id?:int,tags?:string[],owner?:\App\User,x?:mixed,...}`,
		`App\native`:      `array{id?:int,...}|null`,
		`App\falsy`:       `array{id?:int,...}|false`,
		`App\tagged`:      `array{id:string}`,
		`App\vague`:       `array{id?:int,...}`,
		`App\scalar`:      ``,
		`App\emptyValue`:  `array{id?:mixed,...}`,
		`App\badKey`:      ``,
		`App\listForm`:    ``,
		`App\notArray`:    ``,
		`App\noArgs`:      ``,
		`App\placeholder`: ``,
		`App\obj`:         `object{a?:int,...}`,
		`App\objInt`:      ``,
		`App\lookalike`:   ``,
	}
	for _, fn := range fs.Functions {
		if got := fn.DocReturn; got != want[fn.FQN] {
			t.Errorf("%s: DocReturn = %q, want %q", fn.FQN, got, want[fn.FQN])
		}
	}
}

func TestNoReturnAttribute(t *testing.T) {
	fs := extract(t, "noreturn.php", attrHeader+`
#[NoReturn] function a() {}
#[NoReturn] function b(): void {}
#[NoReturn] function c(): never {}
#[NoReturn] function d(): int {}
#[NoReturn(true)] function conditional(bool $exit = false) {}
#[NoReturn()] function emptyArgs() {}
/** @return string */
#[NoReturn] function e() {}
function f() {}
class K { #[NoReturn] function m() {} }
`)
	want := map[string]string{`App\a`: "never", `App\b`: "never", `App\c`: "never", `App\d`: "", `App\conditional`: "", `App\emptyArgs`: "never", `App\e`: "string", `App\f`: ""}
	for _, fn := range fs.Functions {
		if fn.DocReturn != want[fn.FQN] {
			t.Errorf("%s: DocReturn = %q, want %q", fn.FQN, fn.DocReturn, want[fn.FQN])
		}
	}
	if got := fs.Classes[0].Methods["m"].DocReturn; got != "never" {
		t.Errorf("method: DocReturn = %q", got)
	}
}

func TestExpectedValuesAttribute(t *testing.T) {
	fs := extract(t, "expected.php", attrHeader+`
class Mode { const A = 1; const B = 2; }
function f(
	#[ExpectedValues([1, 2])] int $ints,
	#[ExpectedValues(values: ['a', 'b'])] string $strs,
	#[ExpectedValues([1])] ?int $nullable,
	#[ExpectedValues(flags: [1, 4])] int $flags,
	#[ExpectedValues(valuesFromClass: Mode::class)] int $fromClass,
	#[ExpectedValues(flagsFromClass: Mode::class)] int $maskFromClass,
	#[ExpectedValues([Mode::A, 4])] int $constant,
	#[ExpectedValues(flags: [Mode::A, Mode::B])] int $maskOfConstants,
	#[ExpectedValues([Mode::A])] ?int $nullableConstant,
	#[ExpectedValues(flags: [Mode::A])] string $maskOnString,
	#[ExpectedValues([self::A])] int $relative,
	#[ExpectedValues([Mode::class])] int $className,
	#[ExpectedValues([Mode::A | Mode::B])] int $expression,
	#[ExpectedValues(valuesFromClass: 'Mode')] int $notClassConstant,
	#[ExpectedValues([SOME_GLOBAL])] int $global,
	#[ExpectedValues(['a'])] int $wrongKind,
	#[ExpectedValues([1])] array $notScalar,
	#[ExpectedValues([])] int $none,
	#[ExpectedValues([1, , 2])] int $hole,
	#[ExpectedValues(flags: [])] int $noFlags,
	#[ExpectedValues] int $bare,
	/** @param-ignored */ int $plain,
) {}
class P {
	#[ExpectedValues([1, 2])]
	public int $mode;
	#[ArrayShape(['a' => 'int'])]
	public array $row;
}
`)
	want := map[string]string{
		"ints": "int<1, 2>", "strs": "'a'|'b'", "nullable": "int<1, 1>|null", "flags": "int<0, 1>|int<4, 5>",
		"fromClass": `\App\Mode::*`, "maskFromClass": `int-mask-of<\App\Mode::*>`, "constant": `\App\Mode::A|4`,
		"maskOfConstants": `int-mask<\App\Mode::A,\App\Mode::B>`, "nullableConstant": `\App\Mode::A|null`,
	}
	for _, p := range fs.Functions[0].Params {
		if p.DocValues != want[p.Name] {
			t.Errorf("$%s: DocValues = %q, want %q", p.Name, p.DocValues, want[p.Name])
		}
	}
	c := fs.Classes[1]
	if got := c.Props["mode"].DocValues; got != "int<1, 2>" {
		t.Errorf("property values = %q", got)
	}
	if got := c.Props["row"].DocType; got != "array{a?:int,...}" {
		t.Errorf("property shape = %q", got)
	}
}

func TestLanguageLevelParamType(t *testing.T) {
	fs := extract(t, "lang.php", `<?php
use JetBrains\PhpStorm\Internal\LanguageLevelTypeAware;
function f(#[LanguageLevelTypeAware(['8.0' => 'int'], default: 'string')] $x, $y) {}
`)
	fn := fs.Functions[0]
	if p := fn.Params[0]; p.Type != "int" || p.TypeVer == "" {
		t.Fatalf("param = %+v", p)
	}
	if got := paramsAt(fn.Params, phpversion.PHP74)[0]; got.Type != "string" || got.TypeVer == "" {
		t.Errorf("at 7.4 = %+v", got)
	}
	if got := paramsAt(fn.Params, phpversion.PHP80)[0]; got.Type != "int" {
		t.Errorf("at 8.0 = %+v", got)
	}
	if got := paramsAt(fn.Params, 0)[0]; got.Type != "int" {
		t.Errorf("latest = %+v", got)
	}
	if paramsUnchanged(fn.Params, phpversion.PHP80) {
		t.Error("version-typed parameters reported unchanged")
	}
}
