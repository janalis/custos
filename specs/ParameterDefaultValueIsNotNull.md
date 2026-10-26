---
id: ParameterDefaultValueIsNotNull
group: Code style
kind: semantic
needs: [names, hierarchy, index]
php: { min: "", max: "" }
---

# ParameterDefaultValueIsNotNull

## Summary

Optional parameters that may be "absent" are best expressed with a `null`
default (and a nullable type) rather than a sentinel such as `0`, `''` or
`[]`. This opt-in style rule flags non-null defaults where `null` could be used.

## Detection

Applies to every function-like declaration: named functions, methods
(including abstract and interface methods), closures and arrow functions.

- D1: a parameter has a default value that is not the `null` constant (`null`
  matched case-insensitively: `null`, `NULL`, `Null`; a fully-qualified
  `\null` also counts as null).
- D2: the parameter's native (declared) type is either absent, or includes
  `null` (`?T`, a union containing `null`, or standalone `null`). Docblock
  `@param` types are ignored.
- D3: a parameter satisfying D1 and D2 is a violation; every violation of the
  function is reported separately.

## Exceptions (no report)

- E1: the parameter has a native type that does not include `null`
  (`string $s = ''`, `int|float $n = 0`, `array $a = []`, `mixed $m = 1`):
  `null` cannot be used without changing the signature.
- E2: override context: the function is a method of a class that has a parent
  class (`extends`), and that parent class — including anything it inherits
  from its own ancestors or traits — has a method with the same name
  (case-insensitive) that is not private. Then nothing is reported for this
  method at all (the signature is dictated by the parent). Interfaces
  implemented by the class do not trigger this exception.
- E2a (custos, see Divergences): the same holds when the parent class, or a
  parent class or trait further up its hierarchy, cannot be resolved: the
  unknown class may declare the method and dictate its signature. Private
  methods and constructors are still reported: no parent dictates their
  parameters.
- E3: a private method in the parent does not trigger E2.
- E4 (custos, see Divergences): the default value is a `new` expression
  (`$clock = new SystemClock()`, PHP 8.1): an object default is the modern
  replacement for `= null` followed by `$clock ??= new SystemClock();`.

## Report

- Range: the whole parameter declaration: from its first token (attributes,
  promotion modifiers, type, `&`, as present) to the end of its default value.
- Severity: info (weak warning).
- Message: "Prefer null as the default value for this parameter."

## Fix

None.

## Options

None.

## PHP versions

No gating (the rule is disabled by default).

## Examples

```php
<?php
class Shape {
    public function scale(
        $factor,
        <weak_warning descr="Prefer null as the default value for this parameter.">$min = 1</weak_warning>,
        $max = null,
        <weak_warning descr="Prefer null as the default value for this parameter.">$tags = ['a']</weak_warning>
    ) {}

    private function tint(<weak_warning descr="Prefer null as the default value for this parameter.">$hue = 'red'</weak_warning>) {}
}

class Circle extends Shape {
    public function scale($factor, $min = 1, $max = null, $tags = ['a']) {}   // overrides parent

    private function tint(<weak_warning descr="Prefer null as the default value for this parameter.">$hue = 'blue'</weak_warning>) {}  // parent's is private
}

function pad(<weak_warning descr="Prefer null as the default value for this parameter.">?int $width = 8</weak_warning>, int $fill = 0) {}
function wrap(<weak_warning descr="Prefer null as the default value for this parameter.">string|null $glue = ','</weak_warning>, string $edge = '') {}
/** @param int|null $n */
function count_up(<weak_warning descr="Prefer null as the default value for this parameter.">$n = 10</weak_warning>, $m = NULL) {}
$cb = function (<weak_warning descr="Prefer null as the default value for this parameter.">$limit = 3</weak_warning>) {};
```

## Divergences

- E4: upstream also asks for `null` instead of a `new` default value.
- **Unresolved ancestors (custos diverges, E2a).** Upstream reports a
  method whose class extends a parent it cannot resolve (a framework base
  class outside the analyzed tree, such as `setUp($importViews = TRUE)` on
  a Drupal test base), although that parent most likely dictates the
  signature.
- `mixed` is treated as a type that does not include `null` (matching
  upstream, whose declared type for `mixed` does not list `null`).
