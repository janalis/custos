---
id: DynamicCallsToScopeIntrospection
group: Language level migration
kind: semantic
needs: [names, index]
php: { min: "7.1", max: "" }
---

# DynamicCallsToScopeIntrospection

## Summary

Since PHP 7.1, functions that read or write the caller's local scope
(`compact`, `extract`, `func_get_args`, …) refuse to run when invoked
indirectly — through a variable holding their name or as a callback string.
Such calls only produce a warning and do nothing useful.

## Detection

Only active when the configured PHP level is **7.1 or higher**.

Scope-sensitive function names (set **S**, matched case-insensitively, as
PHP function names are):
`compact`, `extract`, `func_get_args`, `func_get_arg`, `func_num_args`,
`get_defined_vars`, `mb_parse_str`, `parse_str`.

Callback-taking functions and the 0-based index of their callback argument
(table **C**, function name compared case-insensitively against the last segment
of the called name; the call must resolve to the global function — written
unqualified or `\`-qualified, not a `use function` import from another
namespace, nor an unqualified call in a namespace declaring that function):

| function | callback index |
|---|---|
| `call_user_func` | 0 |
| `call_user_func_array` | 0 |
| `array_map` | 0 |
| `array_filter` | 1 |
| `array_reduce` | 1 |
| `array_walk` | 1 |
| `array_walk_recursive` | 1 |

For every function call expression determine a target expression `T`:

- **D1** Dynamic call: the call has no static name (callee is an expression,
  e.g. `$fn(...)`, `($fn)(...)`). `T` is the callee expression.
- **D2** Named call to a function in table C with at least `index + 1`
  arguments: `T` is the argument at the callback index.
- Any other call: nothing to do.

Then resolve `T` to a single string literal `L`:

- **D3** If `T` itself is a string literal, `L = T`.
- **D4** Otherwise compute the set of possible values of `T` (rules R1–R7
  below), keep only the members that are string literals; if exactly **one**
  string-literal node remains, that is `L`; otherwise (zero, or two or more —
  even two identical literals written in different places) no report.

Possible-values rules (applied recursively, each node visited at most once to
avoid cycles; parentheses around any expression are stripped first):

- **R1** Ternary `c ? a : b`: union of the values of `a` and `b` (the
  condition is not a value).
- **R2** Null coalescing `a ?? b`: union of the values of `a` and `b`.
- **R3** Variable `$v`: look up the innermost enclosing function, method,
  closure or arrow function. If there is none (top-level code), the set is
  empty. Otherwise only definitions that can **reach** this use count:
  - a plain assignment (`=`, not compound `.=` etc., not by-reference, not
    destructuring into lists) to `$v` in that function's body reaches the
    use when it is reachable code and either
    - it is written before the use and no later *dominating write* comes
      between it and the use, or
    - it is written after the use (or contains it, `$v = f($v)`) and a loop
      (`for`, `foreach`, `while`, `do-while`) encloses both, with no
      dominating write inside that loop;
  - a dominating write is an assignment statement to `$v` (any `=` form)
    sitting directly in a statement list (block or `case`) that encloses
    the use and written before it, or a `foreach` whose value or key
    variable is `$v` and whose body encloses the use;
  - assignments inside nested functions, closures and arrow functions are
    another scope and are ignored, except in a closure importing `$v` by
    reference (`use (&$v)`), whose assignments always count;
  - the parameter default of `$v` counts only when no dominating write
    precedes the use.
  Add the values of every reaching assignment's expression; for chained
  assignments `$v = $w = expr` the innermost non-assignment value `expr` is
  used.
  **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report.
- **R4** Property fetch (`$this->p`, `$obj->p`): resolve to the declared
  property; add the values of its default value (skipped when the default's
  source text ends with the property name); then apply the assignment scan of
  R3 for the property-fetch expression in the enclosing function and in the
  declaring class's constructor.
- **R5** Class constant fetch `A::K`: resolve to the constant declaration and
  add the values of its initializer.
- **R6** Global constant `NAME` other than `true`/`false`/`null`: resolve to
  its definition (`define('NAME', value)` / `const NAME = value`) and add that
  value as-is (no further recursion). Unresolvable → empty.
- **R7** Anything else: the expression itself.

Finally:

- **D5** Take `L`'s content between the quotes, unescape it according to the
  quote style (single- vs double-quoted rules), and drop one leading `\` if
  present. If the result equals one of the names in S ignoring case
  (`'Compact'` and `'EXTRACT'` match), report `T`.

## Exceptions (no report)

- **E1** PHP level below 7.1.
- **E2** Direct calls by name (`compact('a')`, `\extract($row)`) — only
  indirect invocation is a problem.
- **E3** Callback functions with too few arguments to reach the callback
  index (`array_filter($list)`).
- **E4** Callback strings naming other functions (`'trim'`) or with extra
  text (`'parse_str '`).
- **E5** Variables assigned at top level (outside any function) — no
  resolution happens there.
- **E6** Ambiguous resolution: two or more string-literal candidates (even
  equal ones), e.g. a variable assigned `'extract'` in both branches of an
  `if`, or a ternary with two string branches.
- **E8** Assignments that cannot reach the call (overwritten before it,
  written after it outside a loop, unreachable) do not contribute (R3).
- **E7** Callback given as a closure, array callable, first-class callable
  syntax, or any non-string value.

## Report

- Range: exactly the target expression `T` (the callback argument node — e.g.
  the string literal including its quotes — or the callee expression of the
  dynamic call, e.g. `$fn` without the argument list).
- Severity: warning.
- Message: `'{name}' reads the caller scope and cannot be invoked indirectly since PHP 7.1.`
  where `{name}` is the resolved function name as written in the literal
  (without leading `\`).

## Fix

None.

## Options

None.

## PHP versions

Active only for PHP ≥ 7.1 (the upstream fixture runs at 7.1). Under PhpStorm's
default test level (below 7.1) nothing is reported.

## Examples

```php
<?php
function collect($handler = 'func_get_args')
{
    $grab = <warning descr="'func_get_args' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$handler</warning>();

    $rows = array_map(<warning descr="'extract' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'extract'</warning>, $input);
    call_user_func_array(<warning descr="'compact' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">'\compact'</warning>, ['a', 'b']);

    $reader = 'mb_parse_str';
    <warning descr="'mb_parse_str' reads the caller scope and cannot be invoked indirectly since PHP 7.1.">$reader</warning>($query, $out);

    $either = $flag ? 'extract' : 'compact';
    $either($row);

    array_filter($items);
    array_filter($items, 'trim');
    compact('grab', 'rows');
}

$top = 'get_defined_vars';
$top();
```

## Divergences

- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable.
- Short ternary `a ?: b` in R1: which operands upstream treats as values is not
  covered by fixtures. Recommendation: union of `a` and `b`.
- **Reaching assignments (custos diverges):** upstream's R3 collects every
  assignment to the variable in the function regardless of order or
  reachability, so `$fn = 'extract'; $fn = 'trim'; $fn($row);` or an
  assignment written after the call (outside any loop) is taken as the
  callee and reported. custos keeps only the assignments that can reach the
  use (R3). A side effect: a variable overwritten with the same literal
  (`$f = 'extract'; $f = 'extract'; $f()`) now has a single candidate and is
  reported, which is correct.
- **Case-insensitive names (custos diverges):** upstream matches the
  callback-taking function names of table C and the resolved callback name
  case-sensitively, so `Array_Map('Extract', $rows)` or `$f = 'COMPACT'; $f()`
  is missed although PHP resolves function names case-insensitively. custos
  matches both ignoring case.
- **Callback-taking function resolved (custos diverges):** upstream takes
  table C from the written last segment, so a user `App\array_map()` (which
  may never invoke its argument) is treated as the built-in. custos only
  inspects the callback argument when the call resolves to the global
  function.
