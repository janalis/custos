---
id: SuspiciousReturn
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SuspiciousReturn

## Summary

A `return` inside a `finally` block discards whatever the `try` block was
doing on its way out: its own return value is replaced and any exception in
flight is swallowed. When the `try` block itself returns or throws, this is
almost always a bug.

## Detection

- **D1** A `return` statement (with or without a value).
- **D2** Walking up its ancestors, the first ancestor that is either a
  `finally` clause, a function/method/closure/arrow function, or the file is
  a `finally` clause. (A `return` inside a closure defined in a `finally`
  block belongs to the closure and is not considered.)
- **D3** The `try` block of the `try` statement owning that `finally` clause
  (only the `try { ... }` block, not its `catch` blocks) contains, at any
  depth, a `return` statement or a `throw` (statement or expression). The
  search does not enter functions, closures, arrow functions or classes
  (including anonymous classes) declared in the `try` block: their
  `return`/`throw` do not leave the block.

## Exceptions (no report)

- **E1** `return` in a `finally` whose `try` block neither returns nor
  throws (returns/throws only in `catch` blocks do not count).
- **E2** `return` in `try` or `catch` blocks.
- **E3** `return` inside a function declared in a `finally` block.
- **E4** `try` block whose only `return`/`throw` sit inside a closure, arrow
  function, nested function or anonymous class
  (`try { $f = function () { return 1; }; } finally { return 2; }`).

## Report

- Range: the whole `return` statement, from `return` through its terminating
  `;` inclusive.
- Severity: error.
- Message: `Returning from 'finally' discards the try block's return value or exception.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function load($path) {
    try {
        if (!is_file($path)) {
            throw new InvalidArgumentException($path);
        }
        $data = file($path);
    } finally {
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return null;</error>
    }
}

function fetch($id) {
    try {
        foreach ([$id] as $k) { return $k; }
    } catch (Exception $e) {
    } finally {
        $cleanup = function () { return true; };
        <error descr="Returning from 'finally' discards the try block's return value or exception.">return;</error>
    }
}

function quiet() {
    try {
        work();
    } catch (Exception $e) {
        return 1;
    } finally {
        return 0;
    }
}
```

## Divergences

- **D3 — custos diverges from upstream.** Upstream's search for
  `return`/`throw` in the `try` block also descends into closures, arrow
  functions and anonymous classes defined there, so
  `try { $f = function () { return 1; }; } finally { return 2; }` is
  reported although nothing in the `try` block itself returns or throws and
  the `finally` return discards nothing. custos stops at those boundaries
  (E4). No upstream fixture covers it.
