---
id: InstanceofCanBeUsed
group: Language level migration
kind: semantic
needs: [types, index, hierarchy]
php: { min: "", max: "" }
---

# InstanceofCanBeUsed

## Summary

Testing an object's class through string comparisons or reflection-like
helpers (`get_class($o) === 'Foo'`, `is_a($o, 'Foo')`,
`in_array('Foo', class_parents($o))`, …) is slower and less readable than the
`instanceof` operator, and hides the class name from refactoring tools.
Suggest `instanceof` when the subject is known to be a non-string value and
the named class exists.

## Detection

All function names below are matched on the name part, case-insensitively
(as PHP compares function names: `IS_A(...)` matches), and the call must resolve to the global PHP function of
that name with PHP's runtime rules: `is_a(...)` and `\is_a(...)` match;
`App\is_a(...)`, a call imported with `use function App\is_a;`, or an
unqualified call in namespace `App` where `App\is_a()` is a known function
do not. The same holds for the inner `class_implements`/`class_parents`
call of D3. Only plain function calls are considered (no method calls).

### Class-name literal (shared)

A **class literal** is a string literal (single- or double-quoted) without
any interpolation whose content:

- has more than 3 characters, and
- is not exactly `__PHP_Incomplete_Class`.
Its **FQN** is `\` followed by the content with every doubled backslash
`\\` collapsed to a single `\` (so `'Acme\\Mail'` gives `\Acme\Mail`). The
content is not otherwise normalised: a leading backslash in the content
yields an FQN starting with two backslashes, which never matches a class
(see E6).

### Non-string subject (shared)

A subject expression `S` qualifies when it is not a string literal, its
inferred type is fully known (no unknown component), and none of the type's
components is `string`. Untyped variables or expressions with unknown type
never qualify.

### Class existence (shared)

The FQN must name at least one **class** in the project index/stubs
(case-insensitive lookup). Interfaces and traits do not count.

### Patterns

- **D1** `get_class(S)` or `get_parent_class(S)` with exactly one argument
  `S`, whose **direct** parent (no parentheses in between) is a comparison
  with `==`, `!=`, `<>`, `===` or `!==`; the other operand (left or right) is
  a class literal; `S` is non-string; the class exists. Context node = the
  comparison. For `get_class` only, additionally the class must have **no
  direct subclasses** in the index; `get_parent_class` has no such
  restriction.
- **D2** `is_a(S, L)` or `is_subclass_of(S, L)` with exactly two arguments,
  or exactly three arguments where the third is the constant `false` (any
  case); `L` is a class literal; `S` is non-string; the class exists.
  Context node = the call.
- **D3** `in_array(L, I, ...)` with at least two arguments (any third
  argument is ignored), where the second argument `I` is directly a call to
  `class_implements(...)` or `class_parents(...)` with at least one argument;
  that call's first argument `S` is non-string; the first `in_array`
  argument `L` is a class literal; the class exists — or, for
  `class_implements` only, the literal names an **interface** (the values
  `class_implements()` actually returns). Context node = the `in_array`
  call.
- **D4** Negation: only for D1 with operator `!=`, `<>` or `!==` the
  replacement is negated. D2/D3 are never negated.

## Exceptions (no report)

- **E1** `get_class($o) == 'Base'` when `Base` has at least one direct
  subclass.
- **E2** Subject typed as `string` (or a union containing `string`), a
  string literal subject, or a subject with unknown/partially unknown type.
- **E3** `is_a`/`is_subclass_of` with a third argument other than literal
  `false` (e.g. `true`, a variable), or with one or four-plus arguments.
- **E4** Class literal of 3 characters or fewer, `__PHP_Incomplete_Class`,
  interpolated strings, non-literal class arguments (`Foo::class`, a
  variable, concatenation).
- **E5** The named class is unknown, or names an interface/trait (except
  interfaces with `class_implements`, D3).
- **E6** Class literal content starting with a backslash
  (`'\Acme\Mail'`).
- **E7** D1 where the call is wrapped in parentheses before the comparison,
  or compared with an operator other than equality/identity.

## Report

- Range: the context node — for D1 the whole comparison (left operand start
  to right operand end), for D2/D3 the whole call including any leading
  namespace qualifier.
- Severity: warning.
- Message, when the rewrite is exact (F2): `Prefer '{replacement}'.` where
  `{replacement}` is the F1 text.
- Message otherwise: `Consider '{replacement}' (not an exact equivalent).`

## Fix

- **F2** A fix is offered only when the rewrite is an exact equivalent:
  - D2 with `is_a` (`is_a($o, 'X')` on a non-string subject is exactly
    `$o instanceof X`);
  - D1 with `get_class` when the named class is declared `final` and the
    literal spells its FQN with exactly the declared case (`get_class()`
    compares the precise class name, case-sensitively; `instanceof` also
    accepts subclasses and ignores case).
  - D3 with `class_implements` when the literal names an interface spelled
    with exactly its declared case (`in_array` compares names
    case-sensitively) and every component of the subject's type is a class
    (on other values `class_implements()` fails instead of returning false).
  Every other match — `get_parent_class` (D1), `is_subclass_of` (D2),
  `class_parents` and class-named `class_implements` (D3), interface checks
  with a differently-cased literal or a nullable/scalar subject, `get_class`
  with a non-final class or a differently-cased literal — is reported with
  the second message and no fix.
- **F1** Replace the context node with:
  - `{S} instanceof {FQN}` normally;
  - `!{S} instanceof {FQN}` for negated D1 (no space after `!`; `!` binds
    looser than `instanceof`, so this is the negation of the whole check).
  `{S}` is the verbatim source text of the subject; `{FQN}` is the computed
  fully-qualified name with its leading `\` (spelled as in the literal, not
  as declared). No parentheses are added around the result.

## Options

None.

## PHP versions

No gating (upstream fixture runs at the harness default, below 7.1).

## Examples

```php
<?php

final class Invoice        {}
class Document             {}
class Receipt extends Document {}
interface Printable        {}

function check(Invoice $doc, Document $base, string $name, $loose) {
    return [
        <warning descr="Prefer '$doc instanceof \Invoice'.">'Invoice' === get_class($doc)</warning>,
        <warning descr="Consider '!$doc instanceof \Receipt' (not an exact equivalent).">get_class($doc) <> 'Receipt'</warning>,
        <warning descr="Consider '$base instanceof \Document' (not an exact equivalent).">get_parent_class($base) == "Document"</warning>,
        <warning descr="Prefer '$doc instanceof \Invoice'.">\is_a($doc, 'Invoice', FALSE)</warning>,
        <warning descr="Consider '$base instanceof \Receipt' (not an exact equivalent).">is_subclass_of($base, 'Receipt')</warning>,
        <warning descr="Consider '$base instanceof \Document' (not an exact equivalent).">in_array('Document', class_parents($base), true)</warning>,

        get_class($base) === 'Document',
        (get_class($doc)) === 'Invoice',
        get_class($name) == 'Invoice',
        get_class($loose) == 'Invoice',
        is_a($doc, 'Invoice', true),
        is_a($doc, 'Foo'),
        is_a($doc, '\\Invoice'),
        is_a($doc, 'Printable'),
        in_array('Invoice', class_implements($name)),
    ];
}
```

```php
<?php

final class Invoice        {}
class Document             {}
class Receipt extends Document {}
interface Printable        {}

function check(Invoice $doc, Document $base, string $name, $loose) {
    return [
        $doc instanceof \Invoice,
        get_class($doc) <> 'Receipt',
        get_parent_class($base) == "Document",
        $doc instanceof \Invoice,
        is_subclass_of($base, 'Receipt'),
        in_array('Document', class_parents($base), true),

        get_class($base) === 'Document',
        (get_class($doc)) === 'Invoice',
        get_class($name) == 'Invoice',
        get_class($loose) == 'Invoice',
        is_a($doc, 'Invoice', true),
        is_a($doc, 'Foo'),
        is_a($doc, '\\Invoice'),
        is_a($doc, 'Printable'),
        in_array('Invoice', class_implements($name)),
    ];
}
```

## Divergences

- **Interfaces with `class_implements` (custos diverges):** upstream requires
  the literal to name a class (E5), but `class_implements()` only ever
  returns interface names, so the realistic check
  `in_array('Countable', class_implements($o))` is never reported. custos
  accepts interfaces there (D3) and offers the `instanceof` fix when it is
  exact (F2).
- custos diverges from upstream on which matches get a fix (F2). Upstream
  rewrites every pattern to `instanceof`, but only `is_a()` and `get_class()`
  on a final, exactly-spelled class are equivalent: `get_class()` on a
  non-final class rejects subclasses, `get_parent_class($o) == 'X'` is true
  only when `X` is the direct parent, `is_subclass_of()` and `class_parents()`
  exclude `X` itself, and `class_implements()` never lists a class. Applying
  those rewrites silently changes which objects pass the check, so custos
  still reports them as suggestions but offers no fix.
- Namespaced functions (custos diverges from upstream). Upstream matches the
  helper functions by name only, so a user-defined `App\is_a()` or
  `App\get_class()` — whose semantics are unknown — is reported and
  rewritten to `instanceof`. custos only matches calls that resolve to the
  global PHP functions.
- A subject with lower precedence than `instanceof` (e.g. an assignment or
  ternary passed directly as argument) would need parentheses in the
  replacement; upstream never adds them. Recommendation: wrap such subjects
  in parentheses (not covered by fixtures).
- **Function-name case (custos diverges):** upstream matches `get_class`,
  `is_a`, `in_array`, `class_implements` & co. case-sensitively, missing
  `Get_Class($o) === 'Final'` or `IS_A($o, 'X')`. custos matches any case.
