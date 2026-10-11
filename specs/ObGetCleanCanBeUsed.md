---
id: ObGetCleanCanBeUsed
group: Control flow
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# ObGetCleanCanBeUsed

## Summary

Reading the output buffer with `ob_get_contents()` and then discarding it with
`ob_end_clean()` in the next statement is exactly what the single built-in
`ob_get_clean()` does. Merge the two calls.

## Detection

- **D1** A plain function call (not a method/static call) whose name part is
  `ob_end_clean` (case-insensitive, as PHP compares function names), which is by itself an expression statement
  (`ob_end_clean();` — the call is the statement's whole expression; not
  `@ob_end_clean();`, not `$ok = ob_end_clean();`, not inside a condition).
- **D2** The previous sibling statement in the same statement list (skipping
  whitespace and ordinary `//`, `#`, `/* */` comments) is itself a plain
  expression statement (`<expr>;`). If the previous sibling is a doc comment
  `/** */`, an `echo`, `return`, `if`, block or any other statement kind, or
  there is none, nothing is reported.
- **D3** Search the previous statement for plain function calls (not method or
  static calls) in source order, at any depth (call arguments, nested
  expressions), but **not** inside closures, arrow functions or anonymous
  classes (D5). Take the **first** call whose name part is
  `ob_get_contents` (case-insensitive). If there is none, stop. A first-class
  callable (`ob_get_contents(...)`, likewise `ob_end_clean(...)` for D1) only
  creates a closure; it is not a call and never matches.
- **D4** Report only if both the `ob_end_clean` call (D1) and that
  `ob_get_contents` call (D3) resolve to the global built-in functions: an
  unqualified call (in or outside a namespace) whose namespaced counterpart is
  not user-defined, or a fully-qualified `\ob_end_clean` / `\ob_get_contents`.
  A call that resolves to a user function of that name in a namespace
  (`App\ob_get_contents()` or an unqualified call in `namespace App` where
  `App\ob_get_contents` is declared) is not reported. Only the first matching
  `ob_get_contents` is considered; if it fails D4, later ones are not tried.
- **D5** Both calls must run in the same scope, as part of the two
  statements: an `ob_get_contents()` inside a closure, arrow function or
  anonymous class body written in the previous statement only runs when that
  code is called later, so it is ignored by D3. In addition, the previous
  statement must contain exactly **one** such `ob_get_contents()` call
  (same scope rule): with two, the second would read a buffer that
  `ob_get_clean()` has already closed.

- **D6** The `ob_get_contents()` call must be evaluated every time the
  previous statement runs. Nothing is reported when it sits in a part that
  may be skipped: a branch of a ternary (`?:` included; the condition is
  fine), the right operand of `&&`, `||`, `and`, `or`, `??` or `??=`, a
  `match` arm (its conditions or its result), or an argument, subscript or
  dynamic member name of a chain that a nullsafe `?->` step can cut short.
  Merging there would leave the buffer open whenever that part is skipped.

## Exceptions (no report)

- **E1** `ob_end_clean()` used as a value or with a prefix operator.
- **E2** The previous statement is not an expression statement (e.g.
  `echo ob_get_contents();`, `return …;`) or is separated by a doc comment.
- **E3** Statements not adjacent (any other statement in between).
- **E4** Calls resolving to namespaced user functions (D4).
- **E5** The only `ob_get_contents()` of the previous statement is inside a
  closure, arrow function or anonymous class
  (`$read = fn() => ob_get_contents();`), or the statement calls
  `ob_get_contents()` more than once (`send(ob_get_contents(), ob_get_contents());`).

- **E6** `ob_get_contents(...)` or `ob_end_clean(...)` written as a
  first-class callable (D3).
- **E7** The `ob_get_contents()` call is only conditionally evaluated (D6):
  `$out = $keep ? ob_get_contents() : '';`, `$buf?->store(ob_get_contents());`.

## Report

- Range: the `ob_get_contents(...)` call expression (name through closing
  parenthesis, including any namespace qualifier written before the name).
- Severity: warning.
- Message: `Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().`

## Fix

- **F1** Rename the reported call's function name to `ob_get_clean` (only the
  name identifier changes; a leading `\` and the argument list are kept).
  The new name is written `\ob_get_clean` when an unqualified call to it at
  that position would not reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace).
- **F2** Delete the `ob_end_clean();` statement (the statement node including
  its `;`). Surrounding whitespace is not touched (the line it was on becomes
  blank / the whitespace before and after remains).

- **F3** The fix is offered only when nothing else of the statement can
  write to or look at the buffer once the call has returned, because that
  code would then run with the buffer already closed. Walking outwards from
  the call up to the statement, only these wrappers are accepted:
  parentheses, a cast, a call to the global `strlen`, `trim`, `ltrim` or
  `rtrim` with the value as its single (not unpacked) argument, and a plain
  `=` assignment to a simple `$name` variable. Anything else — `print`, an
  operator with another operand, a user or other function call, a method
  call, a compound assignment, a property, element, variable-variable or
  destructuring target — keeps the report but withholds the fix.

Examples:

- `print ob_get_contents();⏎ob_end_clean();` → reported, not rewritten
  (the printed text would no longer land in the discarded buffer).
- `$html = ob_get_contents();⏎ob_end_clean();` → `$html = ob_get_clean();⏎`
- `$len = strlen(\ob_get_contents());⏎ob_end_clean();` →
  `$len = strlen(\ob_get_clean());⏎`

## Options

None.

## PHP versions

None (`ob_get_clean()` exists since PHP 4.3). The upstream fixture runs at the
IDE test default level (below 7.1); nothing in it is version-sensitive.

## Examples

```php
<?php
function renderPanel($tpl) {
    ob_start();
    include $tpl;
    $markup = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    ob_start();
    show_footer();
    $size = strlen(<warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">\ob_get_contents()</warning>);
    // drop the buffer
    ob_end_clean();

    ob_start();
    echo ob_get_contents();
    ob_end_clean();

    ob_start();
    $copy = ob_get_contents();
    flush_logs();
    ob_end_clean();

    ob_start();
    $last = ob_get_contents();
    $ok = ob_end_clean();

    return [$markup, $size, $copy, $last, $ok];
}
```

```php
<?php
function renderPanel($tpl) {
    ob_start();
    include $tpl;
    $markup = ob_get_clean();

    ob_start();
    show_footer();
    $size = strlen(\ob_get_clean());
    // drop the buffer

    ob_start();
    echo ob_get_contents();
    ob_end_clean();

    ob_start();
    $copy = ob_get_contents();
    flush_logs();
    ob_end_clean();

    ob_start();
    $last = ob_get_contents();
    $ok = ob_end_clean();

    return [$markup, $size, $copy, $last, $ok];
}
```

## Divergences

- Resolution (D4) needs function name resolution with the built-in function
  list. Without an index, an acceptable approximation is: accept unqualified
  and `\`-qualified names, reject any other qualified name (`Foo\ob_get_contents`)
  and reject unqualified names inside a namespace that declares a function of
  the same name in the current file.
- Scope (custos diverges from upstream). Upstream also matches an
  `ob_get_contents()` written inside a closure (or other deferred code) in
  the previous statement, and renames the first of several calls. Renaming a
  deferred call makes the later invocation close a buffer that
  `ob_end_clean()` was meant to close now, and with several calls the later
  ones read a buffer that no longer exists. custos only considers calls that
  run directly in the previous statement and requires exactly one (D5).
- Function-name case (custos diverges from upstream). Upstream matches
  `ob_end_clean` / `ob_get_contents` case-sensitively, so
  `OB_GET_CONTENTS(); Ob_End_Clean();` was not reported although PHP calls
  the same built-ins. custos compares the names case-insensitively (D1, D3).
- **Builtin spelling (custos diverges).** Upstream renames only the
  identifier, so a namespaced or imported `ob_get_clean` captures the
  rewritten call. custos writes `\ob_get_clean` in that case (F1).
- **Conditional evaluation (custos diverges from upstream).** Upstream
  merges an `ob_get_contents()` that only runs on some paths of the previous
  statement, which leaves the buffer open on the others. custos does not
  report those (D6).
- **First-class callables (custos diverges).** `ob_get_contents(...)` and
  `ob_end_clean(...)` are not calls; custos ignores them (D3).
- **Fix safety (custos diverges).** Upstream rewrites whatever the previous
  statement does with the value. custos keeps those reports but only offers
  the rewrite when the rest of the statement is inert (F3).
- **Formatting (custos).** The removed `ob_end_clean();` statement takes
  the whitespace before it along, so no blank line with trailing spaces is
  left.
