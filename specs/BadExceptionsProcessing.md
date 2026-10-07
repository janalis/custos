---
id: BadExceptionsProcessing
group: Architecture
kind: syntax
needs: []
php: { min: "", max: "" }
---

# BadExceptionsProcessing

## Summary
Two smells around `try`/`catch`: a `try` block that wraps too many statements
(hard to tell which one is expected to throw), and a `catch` clause that
binds the exception to a variable but never looks at it (the error is
swallowed or its cause is lost).

## Detection

### Oversized try block
- **D1** A `try` statement whose own block (`try { … }`, not the catch or
  finally blocks) contains **more than 3** statements.
- **D2** Statement counting: count the direct children of the block that are
  statements — every top-level statement counts as one, whatever it contains
  (an `if` with a large body counts 1, a nested `try` counts 1). Comments and
  doc comments do not count. Nested blocks are not descended into.

### Unused caught exception
- **D3** A `catch` clause that names a variable (`catch (Foo $err)`,
  `catch (A | B $err)`). PHP 8 catch clauses without a variable are ignored.
- **D4** The variable name does not occur anywhere inside the catch body:
  search every variable node in the body at any depth, including nested
  closures, their `use (...)` lists and arrow functions; a match is any
  variable whose name equals the caught variable's name (case-sensitive).
  Only real variable nodes count: `compact('err')` is not a use, while an
  interpolated `"… $err …"` contains a variable node and is a use.
- **D4b** (custos refinement, see Divergences) The variable name does not
  occur after the catch clause in the same scope either. The scope is the
  body of the innermost enclosing function, method or closure (or the file's
  top-level code outside any function). Search the variable nodes located
  after the end of this catch clause: the remaining catch clauses and the
  `finally` block of the same `try`, then everything after the `try`
  statement up to the end of the scope. Bodies of nested named functions,
  classes and closures are skipped, except closure `use (...)` lists and
  arrow-function bodies (which capture the enclosing variables). Occurrences
  inside another catch clause that binds the **same** name (its variable and
  its body) do not count. Any remaining occurrence (`throw $err ?? …`,
  `return $err;`, `if ($err !== null)`, `"… $err"`) → no report.
- **D5** Two flavours (same range, different message):
  - **D5a** the catch body contains **no** statements (empty, or only
    comments) → "swallowed" message;
  - **D5b** the body contains at least one statement (counted as in D2) →
    "cause lost" message.

## Exceptions (no report)
- **E1** A `try` block with 0–3 statements.
- **E2** A catch variable that is referenced at least once in its body
  (read, written, passed on, used inside a closure).
- **E2b** A catch variable that is referenced later in the same scope
  (D4b), e.g. stored for a retry loop or rethrown after cleanup.
- **E3** Catch clauses without a variable, or incomplete/broken catch clauses
  whose variable has an empty name.

## Report
- D1: range = the `try` keyword only (3 characters). Severity: info.
  Message: `Too many statements in this try block; extract some of them so the
  failing call is obvious.`
- D5a: range = the caught variable including `$` (e.g. `$err`). Severity: info.
  Message: `Caught exception is silently discarded; at least log it.`
- D5b: same range and severity. Message: `Caught exception is dropped; log it
  or pass it on as the previous exception.`

## Fix
None.

## Options
None.

## PHP versions
No gating. The variable-less catch form (PHP 8.0) is parsed at every level
and simply never matches D3.

## Examples

```php
<?php
function load(string $path) {
    <weak_warning descr="Too many statements in this try block; extract some of them so the failing call is obvious.">try</weak_warning> {
        $h = fopen($path, 'r');
        $line = fgets($h);
        fclose($h);
        return trim($line);
    } catch (\ErrorException $problem) {
        error_log($problem->getMessage());
    }

    try {
        $a = 1;
        // comments are not statements
        $b = 2;
        if ($a) { $b++; $b++; $b++; }
    } catch (\LogicException <weak_warning descr="Caught exception is silently discarded; at least log it.">$ignored</weak_warning>) {
        // nothing to do
    } catch (\DomainException | \RangeException <weak_warning descr="Caught exception is dropped; log it or pass it on as the previous exception.">$lost</weak_warning>) {
        throw new \RuntimeException('cannot load');
    } catch (\TypeError $kept) {
        $log = function () use ($kept) { return $kept; };
    } catch (\Error) {
    }
}
```

```php
<?php
function fetchWithRetry(Client $client, string $url) {
    $failure = null;
    for ($attempt = 0; $attempt < 3; $attempt++) {
        try {
            return $client->get($url);
        } catch (\RuntimeException $failure) {      // E2b: read after the loop
            usleep(100);
        }
    }
    throw new \LogicException('gave up', 0, $failure);
}

function cleanup($handle) {
    try {
        flush_buffers($handle);
    } catch (\Exception $pending) {               // E2b: used in finally
    } finally {
        fclose($handle);
        if (isset($pending)) { report($pending); }
    }
    try {
        rewind_all($handle);
    } catch (\Exception <weak_warning descr="Caught exception is silently discarded; at least log it.">$oops</weak_warning>) {
    }
    try {
        close_all($handle);
    } catch (\Exception $oops) {                  // same name rebound: does not save the first one
        report($oops);
    }
    $later = function () { return $oops; };       // separate scope: ignored
}
```

## Divergences
- **Uses after the catch (D4b/E2b) — custos refinement, not upstream.**
  Upstream looks only inside the catch body, so a caught exception kept for
  later (`throw $e ?? new …` after a retry loop, inspected in `finally`) is
  reported as discarded although it is not (found on real code). custos also
  searches the rest of the enclosing scope after the catch clause. The
  upstream fixture's caught variables are never used afterwards, so
  conformance is unaffected. Recorded in `docs/decisions.md`
  ("Spec-level false positives").
- Whether an empty statement `;` inside the block counts as a statement is
  unverified upstream. Recommendation: do not count empty statements.
