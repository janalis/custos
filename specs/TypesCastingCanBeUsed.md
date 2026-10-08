---
id: TypesCastingCanBeUsed
group: Language level migration
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# TypesCastingCanBeUsed

## Summary

Conversion helpers (`intval()`, `floatval()`, `strval()`, `boolval()`,
`settype()`), a string made of a single interpolated expression (`"$x"`), and
explicit `->__toString()` calls all perform a plain type conversion. A cast
expresses the same thing directly and avoids a function call.

## Detection

### Conversion functions

- **D1** A function call whose name (compared case-insensitively) is
  `intval`, `floatval`, `strval` or `boolval` and which resolves to that
  global function (unqualified or `\`-qualified; not `Foo\intval(...)`, a
  `use function` import from another namespace, or an unqualified call in a
  namespace declaring that function), mapped to
  the cast types `int`, `float`, `string`, `bool` respectively.
- **D2** It has exactly one argument `A`; or the name is `intval` with exactly
  two arguments and the second is an integer literal whose source text is
  exactly `10` (`010`, `0xA`, `16`, a variable → no).
- **D3** Replacement `R` = `(type)` + `a`, where `a` = `(` + A text + `)` when
  `A` is a binary expression (any operator incl. `??`) or a ternary (incl.
  `?:`), otherwise A's text.

### settype()

- **D4** A function call named `settype` (last segment, case-insensitive) that
  resolves to the built-in global function (unqualified calls in a namespace
  fall back to the global one unless a same-named function is declared in, or
  imported into, that namespace; a call resolving to a user function is
  skipped; the name inside a `use function` declaration is not a call).
- **D5** Exactly two arguments, the second being a string literal (single or
  double quoted, no interpolation) whose content, compared
  case-insensitively as PHP's `settype()` does (`'INT'`, `"Array"`), is one
  of: `boolean`/`bool` → `bool`, `integer`/`int` → `int`,
  `float`/`double` → `float`, `string` → `string`, `array` → `array`.
- **D6** The call is used as a statement on its own (it is the whole
  expression of an expression statement; its return value is unused).
- **D7** Replacement `R` = `V = (type)V` where `V` is the first argument's
  source text.

### Single-interpolation strings (option REPORT_INLINES)

- **D8** A double-quoted string literal (not heredoc, nowdoc or single-quoted)
  whose content is exactly one interpolation and nothing else — no literal
  characters, no escape sequences, before or after it.
- **D9** If the interpolation is a braced `{$expr}` form, `R` =
  `(string)(` + text of `$expr` (inside the braces) + `)`; for a simple
  interpolation (`$var`, `$var[key]`, `$var->prop`) `R` = `(string)` + its
  text.

### __toString calls (option REPORT_TO_STRING_METHOD_CALLS)

- **D10** A method call (instance `->` or static `::`) whose method name is
  `__toString` (case-insensitive, as PHP compares method names), unless the left side is the class
  reference `parent` (`parent::__toString()`).
- **D11** `R` = `(string)` + text of the left side (the part before `->`/`::`).
  Arguments, if any, are dropped.

## Exceptions (no report)

- **E1** Conversion functions with other argument counts; `intval` with a base
  other than literal `10`; `doubleval()` and other aliases are not covered.
- **E2** `settype` with a non-literal type, an unmapped type (`null`,
  `object`, `INT`, `whatever`), not used as a statement
  (`if (settype($v, 'int'))`), or not resolving to the built-in.
- **E3** Strings with surrounding text (`" $x"`, `"$x "`), two
  interpolations (`"{$a}{$b}"`), heredocs, single-quoted strings.
- **E4** `parent::__toString()`; calls to other methods; option off.

## Report

- Range: the whole function call (D1–D7), the whole string literal including
  quotes (D8–D9), the whole method call including arguments (D10–D11).
- Severity: info (weak warning). Upstream additionally renders the
  conversion-function reports as "deprecated-like" (strikethrough); severity
  stays info.
- Messages:
  - D1–D7: `Use '{R}' instead (a cast is clearer and faster).`
  - D8–D9: `Use '{R}' to make the string conversion explicit.`
  - D10–D11: `Use '{R}' instead of calling __toString() directly.`

## Fix

- **F1** Replace the reported node with `R`. Casts are emitted **without**
  a space between `)` and the operand:
  - `intval($n)` → `(int)$n`; `boolval($a && $b)` → `(bool)($a && $b)`;
    `intval($s, 10)` → `(int)$s`
  - `settype($v, 'double');` → `$v = (float)$v;` (the statement's `;` is
    kept)
  - `"$name"` → `(string)$name`; `"{$o->title}"` → `(string)($o->title)`
  - `$money->__toString()` → `(string)$money`

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| REPORT_INLINES | bool | true | Enables the single-interpolation string check (D8–D9). |
| REPORT_TO_STRING_METHOD_CALLS | bool | false | Enables the `__toString()` call check (D10–D11). |

The upstream fixture enables both.

## PHP versions

No gating. The upstream fixture declares no level (test default, below 7.1).

## Examples

Options: REPORT_INLINES = true, REPORT_TO_STRING_METHOD_CALLS = true.

```php
<?php
namespace Store;

function normalize($qty, $ratio, $label, $item, $price) {
    $a = <weak_warning descr="Use '(int)$qty' instead (a cast is clearer and faster).">intval($qty)</weak_warning>;
    $b = <weak_warning descr="Use '(float)($ratio * 2)' instead (a cast is clearer and faster).">floatval($ratio * 2)</weak_warning>;
    $c = <weak_warning descr="Use '(bool)($qty ?: $ratio)' instead (a cast is clearer and faster).">\boolval($qty ?: $ratio)</weak_warning>;
    $d = <weak_warning descr="Use '(int)$label' instead (a cast is clearer and faster).">intval($label, 10)</weak_warning>;
    $e = intval($label, 8);
    $f = <weak_warning descr="Use '(string)$price' instead (a cast is clearer and faster).">strval($price)</weak_warning>;

    <weak_warning descr="Use '$qty = (int)$qty' instead (a cast is clearer and faster).">settype($qty, 'integer')</weak_warning>;
    <weak_warning descr="Use '$item['tags'] = (array)$item['tags']' instead (a cast is clearer and faster).">settype($item['tags'], "array")</weak_warning>;
    settype($ratio, 'null');
    $ok = settype($ratio, 'int');

    $g = <weak_warning descr="Use '(string)$label' to make the string conversion explicit.">"$label"</weak_warning>;
    $h = <weak_warning descr="Use '(string)($item->name)' to make the string conversion explicit.">"{$item->name}"</weak_warning>;
    $i = "#$label";
    $j = "{$label}{$qty}";
    $k = <weak_warning descr="Use '(string)$price' instead of calling __toString() directly.">$price->__toString()</weak_warning>;
    return compact('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'ok');
}

class Tag extends Base {
    public function __toString()
    {
        return 'tag:' . parent::__toString();
    }
}
```

```php
<?php
namespace Store;

function normalize($qty, $ratio, $label, $item, $price) {
    $a = (int)$qty;
    $b = (float)($ratio * 2);
    $c = (bool)($qty ?: $ratio);
    $d = (int)$label;
    $e = intval($label, 8);
    $f = (string)$price;

    $qty = (int)$qty;
    $item['tags'] = (array)$item['tags'];
    settype($ratio, 'null');
    $ok = settype($ratio, 'int');

    $g = (string)$label;
    $h = (string)($item->name);
    $i = "#$label";
    $j = "{$label}{$qty}";
    $k = (string)$price;
    return compact('a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'ok');
}

class Tag extends Base {
    public function __toString()
    {
        return 'tag:' . parent::__toString();
    }
}
```

## Divergences

- A cast binds more loosely than `**`, `[]` and `->`; replacing a call that is
  the operand of such an operator changes meaning (`intval($x) ** 2` →
  `(int)$x ** 2`, `strval($n)[0]` → `(string)$n[0]`). Upstream does not guard
  this. Recommendation: wrap `R` in parentheses when the call is the base of an
  array access / member access or an operand of `**`. Not covered by upstream
  fixtures.
- Spread arguments (`intval(...$xs)`) count as one argument upstream and give
  broken code. Recommendation: skip calls with a spread argument.
- D10 with a static call on `self`/`static`/a class name yields
  `(string)self`, which is invalid; nullsafe `$o?->__toString()` loses the
  null-safety. Recommendation: only report instance calls with `->`.
- `"${name}"` (dollar-brace interpolation): upstream behaviour unverified.
  Recommendation: do not report it.
- **Name case and resolution (custos diverges):** upstream matches
  `intval`/`floatval`/`strval`/`boolval`, `settype` and `__toString`
  case-sensitively, and the conversion functions on the written last segment
  only. custos matches all three in any case (`FloatVal($x)`,
  `$o->__TOSTRING()`), and skips conversion calls bound to a user function
  (`Money\intval()` declared in the namespace, or a qualified/imported
  non-global name), which a cast would not replace faithfully (D1, D4, D10).
- **`settype` type case (custos diverges):** upstream compares the type
  string case-sensitively, missing `settype($v, 'INT')` although PHP accepts
  type names in any case. custos compares it case-insensitively (D5); the
  cast is written in lower case.
