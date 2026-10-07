---
id: DirectoryConstantCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "", max: "" }
---

# DirectoryConstantCanBeUsed

## Summary
`dirname(__FILE__)` computes at runtime what the magic constant `__DIR__`
(PHP 5.3+) already provides. Use the constant.

## Detection
- **D1** A function call (not a method or static call) whose function name —
  the last segment of the called name — is `dirname`, compared
  case-insensitively as PHP does (`Dirname`, `DIRNAME` match). A leading
  `\` or any namespace qualifier is ignored for this
  test, so `\dirname(...)` matches. The call must also resolve to the global
  `dirname()`: an unqualified call inside a namespace that declares its own
  `dirname` function, or imports one with `use function`, is not reported.
- **D2** The call has exactly one argument.
- **D3** That argument is, directly (no parentheses around it), the magic
  constant `__FILE__` in any case (`__file__` is the same constant).

## Exceptions (no report)
- **E1** Any other argument shape: an expression built from `__FILE__`
  (`dirname(__FILE__ . '/x')`), a parenthesised `(__FILE__)`, a variable, a
  string, etc.
- **E2** Argument count other than one: `dirname()`, `dirname(__FILE__, 2)`.
- **E3** (removed: function and constant names match in any case, see
  D1/D3.)
- **E4** Dynamic calls (`$fn(__FILE__)`), method calls (`$p->dirname(__FILE__)`)
  and static calls.

## Report
- Range: the whole function call expression, from the start of the function
  name (including a leading `\` if present) to the closing `)`.
- Severity: warning (rendered as deprecated in IDEs).
- Message: `Replace dirname(__FILE__) with __DIR__.`

## Fix
- **F1** Replace the whole reported call with `__DIR__`. Surrounding code is
  untouched (`echo dirname(__FILE__);` → `echo __DIR__;`,
  `require dirname(__FILE__) . '/boot.php';` → `require __DIR__ . '/boot.php';`).

## Options
None.

## PHP versions
Upstream applies no gating (`__DIR__` exists since 5.3, the minimum custos
supports).

## Examples

```php
<?php
function bootstrap()
{
    $root = <warning descr="Replace dirname(__FILE__) with __DIR__.">dirname(__FILE__)</warning>;
    require <warning descr="Replace dirname(__FILE__) with __DIR__.">\dirname(__FILE__)</warning> . '/config.php';

    $up    = dirname(__FILE__, 2);
    $other = dirname(__FILE__ . '/../lib');
    $paren = dirname((__FILE__));
    $none  = dirname();
    $dyn   = $resolver->dirname(__FILE__);
}
```

```php
<?php
function bootstrap()
{
    $root = __DIR__;
    require __DIR__ . '/config.php';

    $up    = dirname(__FILE__, 2);
    $other = dirname(__FILE__ . '/../lib');
    $paren = dirname((__FILE__));
    $none  = dirname();
    $dyn   = $resolver->dirname(__FILE__);
}
```

## Divergences
- Upstream matches only the bare last segment of the called name, so a
  namespaced user function call like `Tools\dirname(__FILE__)` is reported as
  well. Recommendation: report only unqualified or `\`-qualified `dirname`
  calls (and unqualified calls that resolve to the global function). No
  fixture covers it.
- **Name case (custos diverges):** upstream compares the function name and
  the magic constant case-sensitively, missing `Dirname(__FILE__)` and
  `dirname(__file__)`. Both names are case-insensitive in PHP, so custos
  matches any case (D1, D3); the fix always writes `__DIR__`.
- **Shadowing namespaced function (custos diverges):** an unqualified
  `dirname(__FILE__)` inside a namespace that declares (or imports) its own
  `dirname` calls that user function, not the built-in, so custos does not
  report it (D1).
