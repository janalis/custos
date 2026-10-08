---
id: UselessUnset
group: Unused
kind: semantic
needs: [flow]
php: { min: "", max: "" }
---

# UselessUnset

## Summary
Calling `unset()` on a function parameter only destroys the function's local
binding: the caller's value (whether passed by value or by reference) is
unaffected, and the local goes away at return anyway. Such an `unset()` is
almost always pointless.

## Detection
Visit every function declaration, method and closure that declares at least
one parameter.

- **D1** For each parameter with a non-empty name `p` (by value or by
  reference, variadic or not, promoted or not):
- **D2** Collect every access to variable `$p` inside that scope's body that
  is reachable by control flow from the scope's entry (code after an
  unconditional `return`/`throw`/`exit` is unreachable; bodies of nested
  closures/functions are separate scopes and are not searched; a closure's
  `use ($p)` list belongs to the outer scope).
- **D3** An access qualifies when the variable node `$p` is **directly** one
  of the arguments of an `unset(...)` statement (the argument is exactly
  `$p`, not `$p[0]`, `$p->x`, `${'p'}`, etc.).
- **D4** Report each qualifying access (see Report for the range). Each
  argument node is reported at most once.

- **D5** The name has not been rebound before the `unset()`: no
  `global $p;` or `static $p;` statement of the same scope (nested scopes
  excluded) starts before the `unset(...)` argument in source order. After
  such a statement `$p` is a global/static alias, not the parameter (E2).

No other condition applies: the parameter may have been reassigned or
modified earlier — the `unset()` is still reported.

## Exceptions (no report)
- **E1** Unsetting an element or property of a parameter (`unset($p['k'])`,
  `unset($p->cache)`).
- **E2** Unsetting non-parameter variables: locals, `static` or `global`
  variables, `foreach` value variables, `$this`.
- **E3** Unreachable `unset()` statements.
- **E4** Code outside any function (top-level `unset($x)`).
- **E5** Parameters of an outer function unset inside a nested closure that
  imported them (`use ($p)`) — inside the closure `$p` is not a parameter.
- **E6** A parameter name rebound by `global $p;` or `static $p;` earlier in
  the scope (`function f(&$p) { global $p; unset($p); }`) — D5.

## Report
- Range:
  - if the `unset(...)` statement has exactly one argument: the whole
    statement, from `unset` through the closing `)` and the terminating `;`
    inclusive;
  - otherwise: only that argument variable (`$p`).
  With several parameter arguments in one multi-argument `unset`, each is
  reported separately.
- Severity: info (rendered as an "unused" highlight; fixture markup
  `weak_warning`).
- Message: `Unsetting a parameter only drops the local variable; this unset() is pointless.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
function consume(array &$queue, $item, ...$rest)
{
    static $seen;
    global $registry;

    unset($queue['head']);
    unset($seen, $registry);
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($item);</weak_warning>
    unset($seen, <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$rest</weak_warning>);
    unset(
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$queue</weak_warning>,
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$item</weak_warning>
    );

    foreach ($registry as $entry) {
        unset($entry);
    }

    $later = function () use ($item) {
        unset($item);
    };

    return;
    unset($item);
}

class Pool
{
    public function release($handle)
    {
        $handle = null;
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($handle);</weak_warning>
    }
}
```

## Divergences
- **D5 — custos diverges from upstream.** Upstream matches on the variable
  name only, so `function f(&$p) { global $p; unset($p); }` is reported as
  "unsetting a parameter" although `$p` has become a global (or static)
  alias at that point, which E2 never reports. custos skips `unset()`
  arguments placed after a `global`/`static` rebinding of the name. No
  upstream fixture covers it.
- **Exposed scope (custos diverges).** When the function body includes a
  file, calls `eval`, `get_defined_vars()` or `compact()` (outside nested
  functions and classes), unsetting a parameter keeps it out of the scope
  that code sees (Joomla module dispatchers: `extract($displayData);
  unset($displayData); include $path;`): nothing is reported in that
  function.
- **Observed unsets (custos diverges from D5's "no other condition").**
  An `unset($p)` followed (in source order, or anywhere in a loop enclosing
  it) by a read of `$p` — including `isset`/`empty`, element and property
  writes, compound assignments — that the parameter's value or an assignment
  made before the unset may reach changes what that code sees (null instead
  of the value): Moodle's `blog_get_headers()` unsets `$userid` so a later
  `!empty($userid)` is false; `unset($config)` in a loop restarts the array
  built by `$config[$k] = …`. Such unsets are not reported; rebinding writes
  (`=`, `global`, `static`, another `unset`) do not count as reads.
