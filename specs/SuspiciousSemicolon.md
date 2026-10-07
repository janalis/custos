---
id: SuspiciousSemicolon
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SuspiciousSemicolon

## Summary
A lone `;` used as the body of an `if`/`elseif`/`else` or of a loop makes that
body empty; the block that follows then runs unconditionally (or only once,
after the loop). This is nearly always an accidental semicolon.

## Detection
- **D1** An empty statement: a statement made of the `;` token only (no
  expression, no keyword).
- **D2** Its direct parent is one of:
  - an `if` clause, an `elseif` clause, or an `else` clause — i.e. the `;` is
    that clause's whole body (`if ($a) ;`, `elseif ($b) ;`, `else ;`). For
    `else if ($c) ;` the `;` is the body of the inner `if` and is reported
    once;
  - a loop: `for`, `foreach`, `while`, `do … while` — the `;` is the loop body
    (`while (f()) ;`, `do ; while ($x);`).

## Exceptions (no report)
- **E1** Empty statements anywhere else: top level, inside `{ }` blocks,
  after another statement (`f();;`), after `}` (`class A {};`), in `switch`
  case lists. (Those belong to `UnnecessarySemicolon`.)
- **E2** `declare(ticks=1);` — `declare` is not in the parent list.
- **E3** Bodies consisting of `{}` or `{ ; }` (the `;` there is inside a
  block).

## Report
- Range: the single `;` character.
- Severity: error.
- Message: `This ';' is the entire body of the statement; probably unintended.`

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function poll($queue, $ready, $late) {
    if ($ready) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    elseif ($late) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    else <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    {
        notify();
    }

    while ($queue->next()) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    do <error descr="This ';' is the entire body of the statement; probably unintended.">;</error> while ($queue->busy());
    for ($i = 3; $i > 0; $i--) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    foreach ($queue as $job) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    if ($late) { ; }
    declare(ticks=1);
    notify();;
}
```

## Divergences
- Alternative syntax (`while ($x): ; endwhile;`, `if ($x): ; endif;`): whether
  the `;` is a direct child of the construct depends on the parser's tree
  shape; upstream behaviour unverified. Recommendation: do not report (the
  statement list of alternative syntax is treated like a block).
