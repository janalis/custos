---
id: TypeUnsafeComparison
group: Type compatibility
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# TypeUnsafeComparison

## Summary
Loose equality (`==`, `!=`, `<>`) silently juggles types (`'abc' == 0` was
true before PHP 8, `'1e1' == '10'` is true). This rule asks for strict
comparison, offers a one-click switch when comparing against a non-numeric
string literal, and flags objects compared to strings that have no
`__toString()`.

## Detection
Visit every binary expression whose operator is `==`, `!=` or `<>`. The
*strict counterpart* is `===` for `==`, and `!==` for `!=` and `<>`.
Let `L` and `R` be the operands as written.

### Step 1 — comparison with a string literal
Applies when `R` or `L` is itself a string literal (single- or double-quoted,
heredoc/nowdoc; an operand wrapped in parentheses such as `('x')` is **not** a
string literal here).
- **D1** Pick the literal: if `R` is a string literal, the literal is `R` and
  the other operand `O` is `L`; otherwise the literal is `L` and `O` is `R`.
  (When both are literals, `R` is the literal and `O` is `L`.) Strip
  parentheses from `O`. The literal's *contents* are the raw text between the
  delimiters (escapes not decoded).
- **D2** *Object in string context*: infer the type of `O`, drop unknown
  parts, and normalise each part: any array type (`X[]`) → `array`; scalar /
  pseudo types (`int`, `integer`, `string`, `bool`, `boolean`, `true`,
  `false`, `float`, `null`, `void`, `mixed`, `callable`, `resource`,
  `iterable`, `object`, `static`, `self`, `$this`, matched
  case-insensitively, with or without leading `\`) → their keyword form; any
  other type, `\Closure` included, stays a fully-qualified class name
  (`\Foo`). If, ignoring `null`,
  at least one part remains and **every** non-`null` part is a class name
  (starts with `\`), then:
  - resolve each class name to its class/interface declarations (project and
    stubs); for the first one found that has no
    `__toString` method — own or inherited from parent classes, traits or
    interfaces, abstract declarations count — report kind M (D7) and stop
    checking further classes; **except** when the type contains `null` and
    the literal's static value is the empty string (`?Invoice $n; $n == ''`):
    that comparison is a null test (`null == ''` holds) and is not reported;
  - in all cases (reported or not), **end** analysis of this expression
    (no kind S or H).
  Notably `callable`, `object` and `static`/`$this` are not class names, so
  they never trigger D2; a `\Closure` operand is a class name and, having no
  `__toString()`, is reported as kind M.
- **D3** *Non-numeric literal*: take the literal's runtime value — escapes
  decoded for single/double-quoted literals; the raw body for a nowdoc or a
  heredoc without backslashes. A literal with interpolation (`"$v"`,
  `"{$o->x}"`) or a heredoc containing escapes has no static value and skips
  D3. If the value is non-empty and is **not** a PHP numeric string, report
  kind S and end analysis. A numeric string is, over the whole value:
  optional leading and trailing whitespace (space, `\t`, `\n`, `\r`, `\v`,
  `\f`), an optional `+`/`-`, digits with an optional `.` (at least one digit
  overall: `1`, `1.`, `.5`, `1.5`), then an optional exponent `e`/`E`,
  optional sign, one or more digits.
  - numeric (no kind S): `'0'`, `'42'`, `'-3'`, `'+7'`, `'1.25'`, `'.5'`,
    `'1.'`, `'1e3'`, `'-2.5E-3'`, `' 1'`, `"7\n"`, `"\x31"`;
  - non-numeric (kind S): `'abc'`, `'1e'`, `'.'`, `'0x1A'`, `'1_000'`.
- Otherwise (empty, numeric or statically unknown literal) continue with
  step 2.

### Step 2 — general hardening
- **D4** Both operands exist (no parse error), and
- **D5** neither operand is a *comparable object*: an operand whose inferred
  type (unknown parts dropped) contains a class-name part (starting with `\`)
  that resolves to a class or interface which is, or extends / implements /
  uses (transitively, through parents, interfaces and traits) one of:
  `\DateTime`, `\DateTimeImmutable`, `\IntlBreakIterator`, `\IntlTimeZone`,
  `\PDO`, `\PDOStatement`, `\ArrayObject`, `\SplObjectStorage`, `\Closure`. Operands are used as written (a parenthesised
  operand has the type of its content). Types such as `\DateTime|false`
  (e.g. return of `date_create()` from stubs) count as comparable.
- **D6** Then report kind H.

At most one report per expression (M, S or H).

## Exceptions (no report)
- **E1** `===`, `!==` and all other operators.
- **E2** An object-typed operand compared with a string literal (D2) when
  every resolved class has `__toString()` — no report at all.
- **E3** Either side is a comparable object (D5) and step 1 did not report.
- **E2a** A nullable object (`?Invoice`, `Invoice|null`) compared with the
  empty string `''`/`""`: no report at all (a null test). Compared with a
  non-empty literal it is still kind M (never true for either member).
- **E4** Union types mixing classes with scalars (`Foo|int`) are not D2; they
  go on to D3 / step 2.

## Report
- Range: the whole binary expression (left operand start to right operand
  end); parentheses around the whole expression are not included.
- Severities and messages (D7):
  - **M** — severity **error**: `{class} has no __toString(), so it cannot be compared to a string.`
    where `{class}` is the class FQN with leading `\` (e.g. `\Invoice`).
  - **S** — severity **warning** (rule default): `Use '{op}' here; the string is not numeric, so strict comparison is safe.`
  - **H** — severity **info** (weak warning): `Prefer '{op}' to avoid implicit type juggling.`
  `{op}` is the strict counterpart.

## Fix
- **F1** Kind S only: replace the operator token with its strict counterpart
  (`==` → `===`, `!=` → `!==`, `<>` → `!==`). Operands and the whitespace
  around the operator are kept unchanged: `$a <> 'x'` → `$a !== 'x'`,
  `$a=='x'` → `$a==='x'`.
- Kinds M and H have no fix.

## Options
None.

## PHP versions
None. Upstream fixtures run at the PhpStorm test default level (5.6–7.0).

## Examples

```php
<?php
interface Labelled            { public function __toString(); }
interface Plain               {}
abstract class BaseTag        { abstract public function __toString(); }
class Tag extends BaseTag     { function __toString(): string { return "t"; } }
class ChildTag extends Tag    {}
class Invoice                 {}
class Stamp extends DateTimeImmutable {}

function compare(Labelled $l, Plain $p, ChildTag $t, Invoice $i, ?Invoice $n, $v, $w) {
    return [
        $l == 'x',
        <error descr="\Plain has no __toString(), so it cannot be compared to a string.">$p != 'x'</error>,
        'name' == $t,
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">'0' == $i</error>,
        ($n) <> '',
        <error descr="\Invoice has no __toString(), so it cannot be compared to a string.">$n == 'paid'</error>,

        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == 'ready'</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">"done" != $v</warning>,
        <warning descr="Use '!==' here; the string is not numeric, so strict comparison is safe.">$v<>'1x3'</warning>,

        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == '1e3'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == "{$w}"</weak_warning>,

        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == '12'</weak_warning>,
        <weak_warning descr="Prefer '!==' to avoid implicit type juggling.">$v != '-.5'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == ""</weak_warning>,
        <weak_warning descr="Prefer '!==' to avoid implicit type juggling.">$v <> $w</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == ('text')</weak_warning>,

        new Stamp() == new DateTime(),
        date_create('now') != $w,
        $v === 'ready',
    ];
}
```

```php
<?php
interface Labelled            { public function __toString(); }
interface Plain               {}
abstract class BaseTag        { abstract public function __toString(); }
class Tag extends BaseTag     { function __toString(): string { return "t"; } }
class ChildTag extends Tag    {}
class Invoice                 {}
class Stamp extends DateTimeImmutable {}

function compare(Labelled $l, Plain $p, ChildTag $t, Invoice $i, ?Invoice $n, $v, $w) {
    return [
        $l == 'x',
        $p != 'x',
        'name' == $t,
        '0' == $i,
        ($n) <> '',
        $n == 'paid',

        $v === 'ready',
        "done" !== $v,
        $v!=='1x3',

        $v == '1e3',
        $v == "{$w}",

        $v == '12',
        $v != '-.5',
        $v == "",
        $v <> $w,
        $v == ('text'),

        new Stamp() == new DateTime(),
        date_create('now') != $w,
        $v === 'ready',
    ];
}
```

## Divergences
- **Numeric-string grammar (custos diverges):** upstream's numeric test only
  knows plain decimals, so exponent forms (`'1e3'`), a trailing dot (`'1.'`)
  and whitespace-padded numbers (`' 1'`) get the strict fix although PHP
  compares them numerically (`1000 == '1e3'` is true, `1000 === '1e3'` is
  false). Upstream also tests the raw source, so escapes and interpolation
  are misjudged. custos decodes the literal and applies PHP's full
  numeric-string grammar, and skips the fix when the value is not static
  (D3). These comparisons get the info-level report instead.
- **`\Closure` (custos diverges).** Upstream lists `\Closure` among the
  comparable objects but first normalises that type to `callable`, so the
  entry never matches: `$onDone == $handler` with two closures gets the
  info report, and `$onDone == 'x'` is treated as a scalar comparison
  (kind S fix). custos keeps `\Closure` as a class name, so closure
  comparisons are exempt from kind H (D5) and a closure compared with a
  string literal gets kind M (D2), since `Closure` has no `__toString()`.
- **E2a — custos diverges from upstream.** Upstream reports kind M for a
  nullable object without `__toString()` compared with `''`
  (`?Invoice $n; $n == ''`). That comparison is true exactly when `$n` is
  null, so it is a (loose) null test rather than an attempt to compare an
  object with a string, and the "cannot be compared to a string" message is
  misleading. custos stays silent for that case only; nullable objects
  compared with a non-empty literal, and non-nullable ones, are still
  reported.
- Heredoc/nowdoc literals: treated as string literals per PhpStorm's model;
  no upstream fixture covers them.
- **Unresolved ancestors (custos).** A class whose parent, interfaces or
  traits (transitively) cannot be resolved may inherit `__toString()` from
  the missing declaration, so D2 skips it instead of reporting kind M
  (`class Ghostly extends Missing {}`, `new Ghostly() == 'x'`: no report;
  analysis still ends, as for any object-only operand).
