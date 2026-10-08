---
id: UnsetConstructsCanBeMerged
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnsetConstructsCanBeMerged

## Summary

`unset()` accepts several arguments, so consecutive `unset(...)` statements can
be collapsed into one call.

## Detection

Visit every `unset(...)` statement.

- **D1** Find the previous sibling statement in the same statement list,
  skipping whitespace and all comments (line comments, block comments and doc
  comments, any number of them).
- **D2** That previous statement is itself an `unset(...)` statement → report
  the current one.
- In a run of N consecutive `unset` statements, statements 2…N are each
  reported.

## Exceptions (no report)

- **E1** The first `unset` of a run.
- **E2** Any other statement (even an empty `;` or an expression) between two
  `unset` statements breaks the run.
- **E3** `unset` statements in different blocks (e.g. last statement of an `if`
  block and first statement after it).

## Report

- Range: the whole reported `unset` statement, from the `unset` keyword through
  the terminating `;` inclusive.
- Severity: info.
- Message: `Consecutive unset() calls; merge them into one.`

## Fix

- **F1** Replace the previous `unset` statement with
  `unset(<args>);` where `<args>` is the previous statement's arguments followed
  by the current statement's arguments, each in its original source text,
  joined by `, `. Then delete the current statement (the `;` included).
  Comments that sat between the two statements stay where they were (now after
  the merged statement); whitespace around the deleted statement may remain
  (comparison is whitespace-normalised).
- **F2** Applying the fixes of a run one after another, in source order,
  accumulates everything into the first statement:
  `unset($a); unset($b); unset($c);` → `unset($a, $b, $c);`.
- Spacing quirks of the original (`unset ($x)`) are normalised to `unset(`
  because the merged statement is regenerated.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function release(array $cache, $tmp, $lock) {
    unset($cache['k1']);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($tmp);</weak_warning>
    // the lock goes too
    /** trailing note */
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset ( $lock );</weak_warning>

    $tmp = 1;
    unset($tmp, $cache);
    <weak_warning descr="Consecutive unset() calls; merge them into one.">unset($lock);</weak_warning>

    if ($tmp) {
        unset($cache);
    }
    unset($tmp);
}
```

```php
<?php
function release(array $cache, $tmp, $lock) {
    unset($cache['k1'], $tmp, $lock);
    // the lock goes too
    /** trailing note */

    $tmp = 1;
    unset($tmp, $cache, $lock);

    if ($tmp) {
        unset($cache);
    }
    unset($tmp);
}
```

## Divergences

None. (Upstream skips plain comments implicitly and doc comments explicitly;
custos skips all comments, which is the same observable behaviour.)
