---
id: PossiblyNullAccess
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# PossiblyNullAccess

## Summary

Report a property, method, class-member or array-element access on a value
that the code itself shows may be null. Calling a method or writing a
property on null throws an `Error`; reading a property or an element of
null raises a warning and yields null. Null that comes only from PHPDoc is
not reported.

## Detection

- D1. An access through `->` (property read or write, method call), `::`
  on an expression (`$x::CONST`, `$x::$prop`, `$x::m()`, `$x::class`), or
  an element read `$x[k]`.
- D2. The type of the accessed value comes from native declarations only
  (typed parameters, properties and returns, literals, operations), narrowed
  by the conditions guarding the access: `!== null`, `isset()`,
  `instanceof`, truthiness, early exits (`return`, `throw`, `continue`,
  `break`), assignments (destructuring included: it replaces what each
  variable held, and an element taken from a list is typed as the list's
  element, as a read of `$list[1]` is), and calls whose argument PHP's own functions
  check. An early exit taken when two tests both hold (`if ($a === null
  && $b === null) { return; }`) proves the first test false wherever a
  later condition proves the second one true, provided the second tests a
  single local on its own (a comparison with null, its truthiness, or the
  negation of either) that has not changed since: past that exit,
  `if ($b === null) { $a->total(); }` is silent. Exits on three tests
  or more are not combined that way. A local that may change without an assignment showing it (bound
  by reference, a `global` or `static` local, passed where it may be
  written by reference) is unknown.
- D3. That type includes null.

## Exceptions (no report)

- E1. The access is nullsafe (`?->`), or follows a nullsafe access in the
  same chain (`$a?->b()->c()`), whose short-circuit may be where null
  comes from.
- E2. A property, static property or element read inside `isset()` or
  `empty()`, or on the left of `??` or `??=`, along a chain of such reads
  (`$a->b['c'] ?? null`): null raises nothing there. A method call ends the
  chain.
- E3. An element written, appended, destructured into (as a value: a key
  of the pattern, `[$map['k'] => $v] = $row`, is read), iterated into,
  incremented, unset, taken or bound by reference (`$r = &$x['k']`), or
  passed to a call whose parameters are unknown: PHP turns null into an
  array there.
  Element reads before PHP 7.4, which read null silently.
- E4. The value's type is unknown or includes `mixed`.
- E5. The code rules null out in a way D2 does not count: a PHPDoc type
  or assertion (`@var`, `@phpstan-assert`, `@psalm-assert`), or an
  `assert()` call, whatever the configuration says about assertions.
- E6. A property checked earlier in the function, read again after a call
  that may have written it: the check still stands for the code's author.
- E7. A call whose identical call earlier in the function was checked by
  the conditions guarding the access (`if ($p->date() !== null) {
  $p->date()->format('c'); }`), or a local holding such a call's result.
- E8. `$this` inside a closure or an arrow function, which may run bound
  to another object.
- E9. Class names (`Foo::bar()`), which are never null.
- E10. A repeated access: only the first access on a value is reported
  when later ones only run after it. The value is a local or a property of
  `$this` (`$this->last`); a later access on it is silent when an earlier
  access on the same value, which D1 to D3 cover and E1 to E9 do not
  excuse, has run on every path reaching it:
  - the earlier access is in a statement before the later one's, or in an
    earlier part of the same expression that is evaluated first (an
    operand, an argument, an array item, the receiver of a call whose
    arguments hold the later one), or in the condition of an `if`,
    `elseif`, `while`, `switch` or `match`, the subject of a `foreach` or
    the initializer of a `for` that holds the later one;
  - on its way to that point, the earlier access is not in a branch, an
    operand that `&&`, `||`, `??` or `??=` may skip, a ternary or `match`
    arm, `isset()`, a loop body or step, a `try`, `do` or `case` block, an
    argument or key after a nullsafe access, or a nested function;
  - between the two, and in any loop around the later access that does
    not hold the earlier one, the value is used only as the receiver of
    accesses: no assignment, element write, argument, comparison or other
    read.
  An access inside the arguments of an earlier access on the same value
  (`$x->m($x->n)`) is reported as well.
- E11. A call to `getInnerIterator()` of `IteratorIterator`,
  `FilterIterator`, `CachingIterator`, `LimitIterator` or
  `NoRewindIterator` (and their subclasses), or a local holding one: PHP
  declares it nullable, but it returns the iterator the constructor took,
  and throws on an object whose constructor did not run. The one of
  `AppendIterator`, null until an iterator is appended, still counts.

## Report

- Range: the accessed value.
- Severity: warning.
- Message: `Check this value for null before the access: it is Order|null.`
  with the value's type, without value details.

## Fix

None: whether to check, default or throw is the author's choice.

## Options

None.

## PHP versions

All versions; element reads from PHP 7.4.

## Examples

```php
<?php
class Order { public function total(): int { return 0; } }
class Cart {
    public function last(): ?Order { return null; }
    public function sum(?Order $order, ?array $lines): int {
        echo <warning descr="Check this value for null before the access: it is array|null.">$lines</warning>['a'];
        $last = $this->last();
        return <warning descr="Check this value for null before the access: it is Order|null.">$last</warning>->total();
    }
}
```

Guards and null-tolerant forms are silent:

```php
<?php
class Order { public function total(): int { return 0; } }
function sum(?Order $order): int {
    echo $order?->total();
    echo $order->total ?? 0;
    if ($order === null) {
        return 0;
    }
    return $order->total();
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies.
These conditions are independently specified and examples are original.
NullPointerException, when it runs together with this rule, leaves to it
the accesses this rule reports. A suppression comment naming NullPointerException
(or its PhpStorm name) suppresses this rule's findings too, so that
comments written for NullPointerException keep silencing the value.
