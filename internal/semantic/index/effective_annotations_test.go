package index

import "testing"

func TestEffectiveAnnotationsExtraction(t *testing.T) {
	ix := New(nil)
	ix.Add(extract(t, "annotations.php", `<?php
namespace Docs;
/** @template T */
class Box {
 /**
  * @var int[]
  * @psalm-var bool[]
  * @phpstan-var string[]
  */
 public array $values;
 /**
  * @param int[] $items
  * @phpstan-param string[] $items
  */
 public function __construct(public array $items) {}
 /**
  * @return T
  * @phpstan-return string
  */
 public function selected() {}
 /** @psalm-return T */
 public function generic() {}
}
/**
 * @param int $first
 * @phpstan-param string $first
 * @psalm-param bool $second
 * @phpstan-return float
 * @return int
 */
function plain($first, $second) {}
/**
 * @template T
 * @param T $value
 * @return T
 * @phpstan-return bool
 */
function overridden($value) {}
/**
 * @return ($value is int ? string : float)
 * @phpstan-return bool
 */
function overriddenConditional($value) {}
/**
 * @template T
 * @phpstan-param T $value
 * @param mixed $value
 * @psalm-return (T is int ? string : bool)
 */
function selectedConditional($value) {}
/**
 * @template T
 * @param T $value
 * @phpstan-param string $value
 * @phpstan-return (T is int ? string : bool)
 */
function overriddenSubject($value) {}
/**
 * @template T
 * @phpstan-param T $value
 * @psalm-param string $value
 * @phpstan-return T
 */
function selectedTemplate($value, $other) {}
/**
 * @template T
 * @phpstan-param class-string<T> $class
 * @phpstan-assert T $value
 * @phpstan-return bool
 */
function selectedAssertion($class, $value) {}
`))
	c := ix.Class(`Docs\Box`, 0)
	if c.Props["values"].DocType != "string[]" || c.Props["items"].DocType != "string[]" {
		t.Fatalf("property selection: values=%+v items=%+v", c.Props["values"], c.Props["items"])
	}
	if m := c.Methods["selected"]; m.DocReturn != "string" || m.GenReturn != "" {
		t.Fatalf("lower-priority class template return resurrected: %+v", m)
	}
	if m := c.Methods["generic"]; m.GenReturn == "" {
		t.Fatalf("prefixed class template return missing: %+v", m)
	}
	if f := ix.Function(`Docs\plain`, 0); f.DocReturn != "float" || f.Params[0].DocType != "string" || f.Params[1].DocType != "bool" {
		t.Fatalf("ordinary prefixed extraction: %+v", f)
	}
	if f := ix.Function(`Docs\overridden`, 0); f.Tpl != nil || f.DocReturn != "bool" {
		t.Fatalf("lower-priority function template return resurrected: %+v", f)
	}
	if f := ix.Function(`Docs\overriddenConditional`, 0); f.CondReturn != "" || f.DocReturn != "bool" {
		t.Fatalf("lower-priority conditional return resurrected: %+v", f)
	}
	if f := ix.Function(`Docs\selectedConditional`, 0); f.CondReturn == "" {
		t.Fatalf("selected conditional subject missing: %+v", f)
	}
	if f := ix.Function(`Docs\overriddenSubject`, 0); f.CondReturn != "" {
		t.Fatalf("lower-priority conditional subject resurrected: %+v", f)
	}
	if f := ix.Function(`Docs\selectedTemplate`, 0); f.Tpl == nil || len(f.Tpl.Params) != 2 || f.Tpl.Params[0] == "" || f.Tpl.Params[1] != "" {
		t.Fatalf("selected template parameter missing: %+v", f)
	}
	if f := ix.Function(`Docs\selectedAssertion`, 0); f.Tpl == nil || f.Tpl.Return != "" || len(f.Asserts) != 1 {
		t.Fatalf("template assertion lost with ordinary selected return: %+v", f)
	}
}
