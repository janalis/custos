---
id: EfferentObjectCoupling
group: Architecture
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# EfferentObjectCoupling

## Summary

A class that refers to many distinct other classes depends on too much of the
code base; it is hard to change and to test in isolation. The rule counts the
distinct class names a class mentions and flags it once the count reaches a
limit.

## Detection

- **D1** Every named class-like declaration: `class`, `interface`, `trait`,
  `enum`. Anonymous classes are not reported themselves (they have no name),
  but their references count for the enclosing declaration (D2).
- **D2** Collect every *class-name reference* anywhere inside the declaration
  (from the `class` keyword to the closing brace, at any depth, including
  method bodies, closures and nested anonymous classes):
  - `extends` / `implements` lists, interface `extends` lists;
  - `use TraitName;` inside the body (each trait name);
  - type declarations of parameters, return types and typed properties
    (each class name of a union / intersection / nullable type);
  - `new Name`, `Name::member`, `Name::class`, `instanceof Name`,
    `catch (Name $e)` / multi-catch names;
  - attribute names `#[Name(...)]`.
  Doc comments (`@var`, `@param`, `@return`, …), strings, function calls and
  constants are **not** class-name references.
- **D3** Resolve each reference to its fully-qualified name (current
  namespace + `use` imports, as PHP does). The coupling count `N` is the
  number of distinct FQNs, compared case-insensitively (PHP class names are
  case-insensitive: `\Foo` and `\foo` are one class). The class's own name is not a reference and is not counted, unless
  the class mentions itself explicitly by name in its body.
- **D4** Report when `N >= optionCouplingLimit`.

## Exceptions (no report)

- **E1** Declarations with fewer than `optionCouplingLimit` distinct
  references.
- **E2** Anonymous classes (no name to report).
- **E3** Built-in type keywords are not class references (see Divergences):
  `int`, `float`, `string`, `bool`, `array`, `callable`, `iterable`,
  `object`, `mixed`, `void`, `never`, `null`, `false`, `true`, and the
  relative names `self`, `static`, `parent`.

## Report

- Range: the name identifier of the declaration.
- Severity: info.
- Message: `Depends on {N} distinct classes; consider splitting it up.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `optionCouplingLimit` | int | 20 | Minimum number of distinct referenced classes that triggers a report (inclusive). The upstream conformance case runs with `2`. |

## PHP versions

No gating.

## Examples

With `optionCouplingLimit = 2`:

```php
<?php
namespace Shop;

use Psr\Log\LoggerInterface;

class Money {}

class Wallet {
    public function add(Money $m): void {}
}

class <weak_warning descr="Depends on 3 distinct classes; consider splitting it up.">Checkout</weak_warning> extends Wallet {
    /** @var \Some\DocOnly */
    private $log;
    public function __construct(LoggerInterface $log, int $tries) { $this->log = $log; }
    public function pay(Money $m): self { return $this; }
}

class Basket {
    public function total(int $cents): string { return (string) $cents; }
}
```

(`Checkout` → `\Shop\Wallet`, `\Psr\Log\LoggerInterface`, `\Shop\Money`;
`Wallet` → only `\Shop\Money`; `Basket` → none.)

## Divergences

- Upstream counts every class-reference node of the syntax tree. Whether its
  tree models scalar type keywords and `self`/`static`/`parent` as class
  references (and how they would be named) is not observable from the
  fixtures. Recommendation: exclude them (E3), so the metric only counts real
  classes. No fixture is affected.
- **Case-insensitive distinctness (custos diverges):** upstream counts
  distinct resolved texts, so `\Foo` and `\foo` (or `Money` and `money`)
  count as two dependencies and inflate the metric. They name the same
  class; custos compares FQNs case-insensitively (D3).
