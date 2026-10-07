---
id: ArgumentUnpackingCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "5.6", max: "" }
---

# ArgumentUnpackingCanBeUsed

## Summary
Since PHP 5.6, `call_user_func_array('fn', $args)` with a literal function
name can be written as a direct call with argument unpacking, `fn(...$args)`,
which is faster and easier to read and analyse.

## Detection
- **D1** Project PHP level ≥ 5.6.
- **D2** A plain function call (not a method/static call) whose last name
  segment, compared case-insensitively (`Call_User_Func_Array` matches), is
  `call_user_func_array`, and
  that resolves to the global function: written fully qualified
  (`\call_user_func_array(...)`), or unqualified where PHP falls back to the
  global function (not imported from elsewhere with `use function`, and no
  function of that name declared in the current namespace). A qualified
  name such as `Lib\call_user_func_array(...)`, or an unqualified call that
  resolves to a user function `Lib\call_user_func_array`, is not reported.
- **D3** Exactly two arguments.
- **D4** The first argument is a string literal (single- or double-quoted)
  that contains **no** interpolation (no embedded variables/expressions).
  Heredoc/nowdoc follow the same rule if the parser treats them as string
  literals; a non-literal first argument (variable, array callable
  `[$obj, 'm']`, closure, concatenation) is never reported.
- **D5** The second argument is one of:
  - a variable (`$args`, including `$$name`);
  - a property fetch (`$this->items`, `$o->p`, `$o?->p`, and static
    property `Cls::$p`);
  - an array literal (`[]`, `[1, 2]`, `array()`, `array($a)`);
  - a call: function call, method call or static method call (`args()`,
    `$this->args()`, `Cls::args()`).
  Anything else (array element access `$a['k']`, ternary, cast,
  `new`, constants, literals other than arrays, parenthesised expressions,
  …) is not reported.
- **D6** Compute `F` = the string literal's contents with escape sequences
  decoded according to its quote style (single quotes: `\\` → `\`,
  `\'` → `'`; double quotes: the PHP double-quoted escapes). E.g. the
  literal `'\\Lib\\fmt'` gives `\Lib\fmt`.
- **D7** The suggested replacement is `C + "(..." + <second argument source
  text verbatim> + ")"`, where the callee `C` spells `F` so that a direct
  call at the reported position reaches the same function (a string
  callable is always an absolute name, while a written call resolves
  against the namespace and imports):
  - `F` starting with `\`: `C = F`;
  - unqualified `F`: `C = F` when an unqualified call there reaches the
    global function `F`, otherwise `\F` (a `use function` import under that
    name, or a function `F` declared in the current namespace);
  - qualified `F` (`Lib\fmt`): `C = F` when it resolves to itself at that
    position (global code, no import of its first segment), otherwise
    `\F`.
- **D8** Key safety of the second argument. From PHP 8.0 both
  `call_user_func_array` and `...` unpacking pass string keys as named
  arguments, so every D5 argument is *safe*. Below 8.0
  `call_user_func_array` ignores keys while unpacking an array with a string
  key throws, so:
  - an array literal whose entries all have no key or an integer-literal key
    is *safe*;
  - an array literal with a string-literal key is *string-keyed*: no report;
  - anything else (a variable, property, call, or an array literal with a
    non-literal key such as `[$k => 1]`) *may have string keys*: report
    without a fix (F2).

## Exceptions (no report)
- **E1** PHP level below 5.6.
- **E2** Argument count other than 2.
- **E3** First argument not a non-interpolated string literal (e.g.
  `[$svc, 'run']`, `"{$prefix}_run"`, `$fn`).
- **E4** Second argument not of a D5 kind.
- **E5** Method named `call_user_func_array` (`$x->call_user_func_array(...)`).
- **E7** Calls that do not resolve to the global `call_user_func_array`
  (D2): `Lib\call_user_func_array(...)`, a `use function` import of another
  function under that name, or an unqualified call in a namespace declaring
  its own `call_user_func_array`.
- **E6** PHP level below 8.0 and the second argument is an array literal with
  a string-literal key (D8).

No check is made that `F` is a valid function name; `'Cls::m'` would yield
`Cls::m(...$a)` (see Divergences).

## Report
- Range: the whole `call_user_func_array(...)` call expression, from its first
  character (including a leading `\` qualifier if present) to the closing
  `)`.
- Severity: warning.
- Message: `Call '{replacement}' directly using argument unpacking (wrap with array_values() when keys are not sequential).`

## Fix
- **F1** Replace the whole reported call with the D7 replacement text. The
  qualifier of `call_user_func_array` itself disappears; a leading `\` that
  was inside the string literal is preserved:
  - `\call_user_func_array('max', $nums)` → `max(...$nums)`
  - `call_user_func_array('\\Lib\\fmt', $this->parts)` → `\Lib\fmt(...$this->parts)`
  - `call_user_func_array("printf", array($f, $v))` → `printf(...array($f, $v))`
  - in `namespace Shop` declaring `function max()`:
    `call_user_func_array('max', $a)` → `\max(...$a)`; and
    `call_user_func_array('Util\fmt', $a)` → `\Util\fmt(...$a)`
  The second argument's text is copied verbatim (inner whitespace kept).
  Offered only when D8 classifies the argument as *safe*.
- **F2** Below PHP 8.0, when the argument *may have string keys* (D8), the
  report carries no fix: the rewrite would throw at run time if the array
  turned out to be string-keyed.

## Options
None.

## PHP versions
- Requires ≥ 5.6. Below 8.0 the fix is restricted to integer-keyed array literals (D8); the examples assume 8.0+.
- Conformance note: the upstream fixture runs at the harness default level
  (no explicit level; below 7.1 but at least 5.6) and expects reports, so the
  default level must satisfy ≥ 5.6.

## Examples

```php
<?php
class Report {
    public function build(array $cols) {
        $a = <warning descr="Call 'sprintf(...$cols)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('sprintf', $cols)</warning>;
        $b = <warning descr="Call 'max(...$this->totals)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">\call_user_func_array("max", $this->totals)</warning>;
        $c = <warning descr="Call 'min(...[3, 9])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('min', [3, 9])</warning>;
        $d = <warning descr="Call '\Lib\fmt(...self::defaults())' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('\\Lib\\fmt', self::defaults())</warning>;

        $e = call_user_func_array([$this, 'render'], $cols);
        $f = call_user_func_array("{$this->kind}_fmt", $cols);
        $g = call_user_func_array('sprintf', $cols['main']);
        $h = call_user_func_array('sprintf');
    }
}
```

```php
<?php
class Report {
    public function build(array $cols) {
        $a = sprintf(...$cols);
        $b = max(...$this->totals);
        $c = min(...[3, 9]);
        $d = \Lib\fmt(...self::defaults());

        $e = call_user_func_array([$this, 'render'], $cols);
        $f = call_user_func_array("{$this->kind}_fmt", $cols);
        $g = call_user_func_array('sprintf', $cols['main']);
        $h = call_user_func_array('sprintf');
    }
}
```

## Divergences
- Upstream does not validate the decoded string: `'Cls::method'`, `''`, or a
  name with spaces would produce invalid or different code. Recommendation:
  only report when `F` matches an optionally `\`-qualified, `\`-separated
  identifier path (no `::`). No upstream fixture covers this.
- custos diverges from upstream on string keys (D8, E6, F2). Upstream offers
  the rewrite for any second argument at every level. Below PHP 8.0 that fix
  can turn a working call into a fatal "cannot unpack array with string
  keys" error, so custos keeps the fix only for arrays known to be
  integer-keyed, reports unknown arrays without a fix, and does not report
  string-keyed literals. From 8.0 both forms treat string keys as named
  arguments, so the fix is always offered there. EA cases run at the 5.6
  default level whose expected output rewrites variable arguments are listed
  in `testdata/ea-divergences.json`.
- **Name resolution (custos diverges):** upstream matches the call name
  textually and ignores its namespace, so calls to a user function such as
  `Lib\call_user_func_array` (qualified, or unqualified inside namespace
  `Lib` where it is declared) are reported and rewritten as if they were
  the built-in. Those calls have unrelated semantics; custos only reports
  calls that resolve to the global function (D2).
- **Function-name case (custos diverges):** upstream matches
  `call_user_func_array` case-sensitively, missing
  `CALL_USER_FUNC_ARRAY('f', $args)`, which calls the same built-in. custos
  matches any case (D2).
- **Callee spelling (custos diverges):** upstream inserts the decoded
  string as written. Inside a namespace that turns the absolute name of a
  string callable into a relative one: `'Util\fmt'` becomes a call to
  `Shop\Util\fmt`, and `'max'` is captured by a `Shop\max` function or a
  `use function` import. custos adds a leading `\` in those cases (D7).
