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
  check. An early exit taken when several tests all hold (`if ($a === null
  && $b === null) { return; }`, up to eight tests joined by `&&` or `and`)
  proves the one test left false wherever later conditions prove all the
  others true, provided each of those tests a single local on its own (a
  comparison with null, its truthiness, or the negation of either) that
  has not changed since: past that exit,
  `if ($b === null) { $a->total(); }` is silent, and so is
  `if ($b === null && $c === null) { $a->total(); }` past
  `if ($a === null && $b === null && $c === null) { return; }`. A local that may change without an assignment showing it (bound
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
  A repeated access on a property (E10) survives calls the same way, so a
  dereference is never weaker than a check. Known limit: a call that sets
  the property to null between the two (`$this->reset()`) is not seen.
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
  excuse, and which is reported (no suppression comment silences it and
  no baseline entry records it, see below) or is itself such a repeated
  access, has run on every path reaching it:
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
    not run the earlier one on each of its iterations (one holding it in
    the subject of a `foreach` or the initializer of a `for` runs it once),
    the value is used only as the receiver of accesses: no assignment,
    element write, argument, comparison or other read. For a property, a
    call between them does not count, as for E6: an access that ran tells
    the property is set just as a check does.
  An access inside the arguments of an earlier access on the same value
  (`$x->m($x->n)`) is reported as well, and so is every access in a
  function or file that holds a `goto`, which may jump past the earlier
  access.
- E11. A call to `getInnerIterator()` of `IteratorIterator`,
  `FilterIterator`, `CachingIterator`, `LimitIterator` or
  `NoRewindIterator` (and their subclasses), or a local holding one: PHP
  declares it nullable, but it returns the iterator the constructor took,
  and throws on an object whose constructor did not run. The receiver's
  type, proven without PHPDoc, must name classes whose subclasses, of PHP
  or of the project, all keep that method: the one of `AppendIterator`,
  null until an iterator is appended, still counts, and so does a
  receiver typed `IteratorIterator`, which may be an `AppendIterator`, or
  one a class of the project extends overriding the method.

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
comments written for NullPointerException keep silencing the value. Likewise,
a baseline entry recorded for NullPointerException covers a finding of this
rule on the same line, and the reverse.

A baseline counts as a silencing for E10: an access after one the baseline
records (an entry of this rule with its message, or of NullPointerException,
on its line text) is reported, as after one a suppression comment silences,
since it may be code written after the baseline. Generating a baseline
therefore records every access E10 would leave out, so that analyzing with
it reports nothing new. An entry covers as many accesses as it counts,
taken in source order: when the baseline counts one and three accesses on
the same value share its line text, the first is recorded, the second is
reported and the third repeats it. An entry of NullPointerException covers
an access on its line whatever its count.
