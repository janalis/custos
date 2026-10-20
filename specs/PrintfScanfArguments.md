---
id: PrintfScanfArguments
group: Probable bugs
kind: semantic
needs: [names, index, flow]
php: { min: "", max: "" }
---

# PrintfScanfArguments

## Summary

The `printf`/`scanf` family takes as many value arguments as the format string
has conversions. A missing or extra argument silently produces wrong output
(or an `ArgumentCountError` in PHP 8), and a stray `%` makes the format
invalid.

## Detection

- **D1** A plain function call (not a method call) that resolves to one of
  these global functions (names compared case-insensitively, as PHP does;
  an unqualified call in a namespace counts only when no function of that
  name is declared in the namespace, a `use function` import of another
  function or a namespace-qualified name does not count; custos diverges):

  | function | format argument position |
  |---|---|
  | `printf`, `sprintf` | 0 |
  | `fprintf`, `sscanf`, `fscanf` | 1 |

  Let `P` be that position. The call must have at least `P + 1` arguments.
  (`vprintf`, `vsprintf`, `vfprintf` are not checked.)

- **D2** Determine every possible format (custos diverges, see
  Divergences): collect the candidate values of `A` with a *complete* value
  discovery — parentheses stripped; full ternary → both result branches;
  short ternary `a ?: b` → `a` and `b`; `??` → both operands; local
  variable → the plain `=` assignments that may reach the call; `$this->p`,
  `self::$p` / `C::$p` → the property's default plus every plain `=`
  assignment to it in the class; class constant → its value; global
  constant → its `const`/`define()` value. The set is *incomplete* — stop,
  no report — when any source cannot be fully accounted for in the file:
  a parameter or closure import whose entry value may reach the call, a
  variable possibly undefined there, a variable written any other way
  (compound or by-reference assignment, `++`/`--`, destructuring, foreach,
  `global`, `static`, catch, imported by reference into a closure), top-level
  variables, an untyped property without default, a property with hooks or
  written other than by plain `=` in its class, unresolved properties,
  constants or classes, `static::` constants outside a final class.
  Every candidate must then be a single- or double-quoted string literal;
  any other candidate (a call, a variable that cannot be expanded, a heredoc
  or nowdoc, `null`, a concatenation) → stop, no report. Steps D3–D6 run on
  each literal; any literal stopping at D3 stops the whole check, and all
  literals must reach the same verdict (all malformed, or all valid with the
  same `expected` count), otherwise stop, no report.
- **D3** If the literal contains interpolation (a double-quoted string with
  embedded variables or expressions, e.g. `"Hi $name %s"`), stop: the format
  is only known at run time. Otherwise let `content` be the format text the
  call receives, i.e. the decoded value of the literal (escape sequences
  processed, so `"%1\$s"` reads `%1$s`, `"\x25d"` reads `%d` and
  `'%\'010d'` reads `%'010d`). Trim leading/trailing whitespace.
  Empty → stop.
- **D4** A `$` in `content` is plain text: single-quoted strings never
  interpolate, and double-quoted ones with interpolation stopped at D3.
  Positional specifications (`%2$s`) are handled by D5. `%%` is a literal
  percent sign: for the printf family every `%%` is removed first
  (left-to-right, non-overlapping); for the scanf family it is skipped
  during the scan, so that a `%` inside a character set is not mistaken
  for one.
- **D5** Scan the format from left to right; every `%` must start a
  conversion specification of the function's family (custos diverges, see
  Divergences). An *argument number* is one or more digits followed by `$`.

  **printf family** (`printf`, `sprintf`, `fprintf`), in order:
  1. `%`;
  2. optionally an argument number;
  3. any number of flags, in any order: `-`, `+`, a space, `0`, or `'`
     followed by any one padding character;
  4. optionally a width: digits or, when the target is PHP 8.0 or later,
     `*` optionally followed by an argument number (the width is read from
     an argument);
  5. optionally `.` followed by a precision: zero or more digits or, from
     PHP 8.0, `*` optionally followed by an argument number;
  6. optionally the length modifier `l`;
  7. one conversion character among `s d u c o x X b g G e E f F`, plus
     `h H` when the target is PHP 8.0 or later.

  **scanf family** (`sscanf`, `fscanf`), in order:
  1. `%`;
  2. optionally `*` (the value is read but not assigned) or an argument
     number;
  3. optionally a width: digits;
  4. optionally one size modifier among `h l L`;
  5. one conversion character among `n d D i o u x X f e E g s c`, or a
     character set: `[`, optionally `^`, optionally a `]` taken as a member,
     then everything up to the next `]`. Text inside the set, `%` included,
     is not scanned for specifications.

  Count `plain` — specifications without argument number that take a value
  (every one except a `*`-suppressed scanf conversion), plus one for each
  `*` width or precision without argument number — and `maxArg`, the largest
  argument number seen (0 if none).
- **D6** Validity: if some `%` does not start a specification of the family
  (unknown conversion character, misplaced flag, unterminated set, end of
  the format) → report **invalid format** on `A` and stop.
- **D7** Arity: `expected = P + 1 + max(plain, maxArg)`. If the call's
  argument count (a spread `...$x` counts as one) differs from `expected`,
  report **argument count**, unless one of:
  - **D7a** the function is `sscanf` or `fscanf`, the call has exactly 2
    arguments, and the call is used as a value: its parent or grandparent is
    an assignment (including `list(...) =` / `[...] =` destructuring, e.g.
    through one pair of parentheses), or it is directly an argument of a
    function or method call;
  - **D7b** the last argument is a spread `...$x`;
  - **D7c** `A` is a plain variable and somewhere in the enclosing function
    body that same variable is the left side of a **compound** assignment
    (`.=`, `+=`, `??=`, …; not plain `=`).

## Exceptions (no report)

- **E1** Possible formats not all known string literals (parameters, even
  with a default; `$message ?: 'fmt %s'`; heredoc/nowdoc alternatives;
  concatenations; calls), or several literals with different verdicts
  (`$verbose ? 'Field %s: %s' : 'Field %s'`).
- **E2** Double-quoted formats with interpolation (`"Hi $name %s"`).
- **E3** Empty/whitespace-only format.
- **E4** `sscanf`/`fscanf` with only subject + format whose result is
  assigned or passed on.
- **E5** Spread as last argument.
- **E6** Format variable later extended with a compound assignment.
- **E7** `%*…` scanf suppressions and `%[...]` character sets are handled by
  D5 (a `%*s` takes no argument; `'%[^%]%%%s'` holds two conversions).
- **E8** Everything PHP itself accepts in a format: flags in any order
  (`'%0+5d'`, `'%-+6d|'`), `'%h'`, `'%H'`, `'%*d'` and `'%.*f'` on PHP 8.0+,
  and the scanf conversions `%i`, `%n`, `%D`, `%hd`, `%Lf`.

## Report

- Invalid format: range = the format argument expression `A` as written in
  the call (the literal, or the variable/constant/property expression that
  resolved to it). Severity: error. Message: `Malformed format string.`
- Argument count: range = the function **name token only** (e.g. `sprintf`),
  not the parentheses or arguments. Severity: error. Message:
  `This call needs {expected} argument(s) in total.`

## Fix

None.

## Options

None.

## PHP versions

The rule runs on every version. The printf conversions `h`/`H` and the `*`
width and precision are accepted only when the target is PHP 8.0 or later;
on older targets they are reported as malformed (D5).

## Examples

```php
<?php
class Report
{
    const ROW = '%s: %d%%';
    public static $line = '%-10s|%5.1f';
    private $head = '%%title%% %s';

    public function render($fh, $name, $total, $raw)
    {
        $fmt = '%%%s=%x';
        echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>($fmt, $name);
        <error descr="This call needs 4 argument(s) in total.">fprintf</error>($fh, self::ROW, $name, $total, $raw);
        echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>(self::$line, $name);
        echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>($this->head);
        echo sprintf(<error descr="Malformed format string.">'50 %, total %s'</error>, $name);
        <error descr="This call needs 4 argument(s) in total.">sscanf</error>($raw, '%d-%d', $a);

        echo sprintf('%05.2f %s', $total, $name);
        echo sprintf("%2\$'#8s %1\$s", $name, $total);
        echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>('%2$s %1$s', $name);
        [$x, $y] = sscanf($raw, '%d-%d');
        $parts = fscanf($fh, '%s %s');
        store(sscanf($raw, '%s'));
        sscanf($raw, '%*d %s', $only);
        sscanf($raw, '%[a-z]', $word);
        printf('%s %s', ...$pair);
        echo sprintf($unknown, $name);
        echo sprintf("Hi $name %s");
    }

    public function grow()
    {
        $tpl = 'id=%d';
        $tpl .= ' name=%s';
        return sprintf($tpl, 1, 'x');
    }
}
```

Notes on the examples: `'%%%s=%x'` → `%s=%x` → 2 conversions → expected
1 + 2. `"%2\$'#8s %1\$s"` decodes to `%2$'#8s %1$s` → highest argnum 2 →
expected 3, given 3. `'%2$s %1$s'` needs 3 arguments but has 2. In `grow()`
the literal found is `'id=%d'` (expected 2, given 3) but `$tpl .= …`
suppresses the report (D7c).
A text like `'99% sure'` is *valid* for D5 (`% s` = space padding + `s`).

## Divergences

- **Every possible format must be known (custos diverges).** Upstream
  keeps only the string-literal results of value discovery and checks the
  format when exactly one is found, ignoring every other candidate. So
  `sprintf($message ?: 'Expected a value. Got: %s', $v, $type)` (a user
  message that may use more placeholders) or a ternary between a literal
  and a nowdoc is checked against the literal alone — false positives on
  real code (infection, webmozart/assert). Parameter defaults are treated
  as the only value although callers can pass another format. custos
  reports only when the complete set of possible formats is known
  literals that all agree (D2): with several agreeing literals it reports
  too (`$c ? 'Field %s: %s' : '%s => %s'` with one value argument).
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analyzed with its initial value only. custos makes the result unknown (stop: neither the format-validity nor the arity check runs, which subsumes D7c for plain variables), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable.
- **Positional formats and `$` text (custos diverges from upstream).**
  Upstream stops whenever the raw format contains `$` followed by a letter
  or digit, to avoid interpolated strings. That also skips every format
  using `%n$` followed directly by a conversion letter (`'%2$s %1$s'`, the
  most common positional form) and single-quoted text such as
  `'Price $amount: %s'`, so wrong argument counts there go unnoticed. custos
  stops only on real interpolation and reads double-quoted formats with
  their escapes decoded (D3/D4).
- **Function names (custos diverges).** Upstream matches the written name
  case-sensitively without resolving it, so `Sprintf('%s %s', $a)` is
  missed while a namespaced user function called `sprintf` (declared in
  the namespace or imported with `use function`) is checked as if it were
  the builtin. custos compares the name case-insensitively and requires the
  call to reach the global function (D1).
- **Redeclarable format properties (custos diverges).** `$this->p` is
  resolved only when the property is private or its class is final (or
  anonymous): a subclass may redeclare it with another format (adodb's
  `$dropIndex = 'DROP INDEX %s'`, redeclared `'DROP INDEX %s ON %s'` by the
  MySQL dictionary), which `$this->p` then reads. One upstream case is
  listed in `testdata/ea-divergences.json`.
- **Two-argument scanf used as a value (custos diverges from D7a).** Any
  use of the returned array counts (`sscanf($t, 'PT%dH%dM%dS') ?? []`,
  `return sscanf(…)`, an array element), not only assignments and call
  arguments; a discarded call or one used as a truth value (`if`/`while`
  condition, `!`, `&&`/`||` operand, ternary condition) is still reported.
- **`l` length modifier (custos diverges).** PHP accepts and ignores an
  `l` before the conversion character (`%ld`, `%.0lf`, common in code
  ported from C: nusoap). D5 accepts an optional `l` after the precision;
  upstream reports such formats as malformed.
- **Per-family grammar (custos diverges).** Upstream recognizes one
  specification shape for both families: a fixed flag order, no `h`/`H`
  conversion, no `*` width or precision, and none of the scanf-only
  conversions or size modifiers; it also drops every `%*` before counting
  and counts a `%` inside a scanf character set. Formats PHP runs without
  complaint (`'%0+5d'`, `'%*d'` and `'%h'` on PHP 8, `'%i %s%n'`, `'%hd'`,
  `'%[^%]%%%s'`) are then reported as malformed or with a wrong count.
  custos follows the grammar PHP applies to each family (D5), gating the
  PHP 8.0 printf additions on the target version. The reverse also holds:
  a specification only the other family knows (`sprintf('%[a-z]', …)`,
  `sscanf($s, '%.2f', …)`, `sscanf($s, '%b', …)`) or an unterminated set
  is rejected by PHP and reported as malformed, where upstream accepts it.
