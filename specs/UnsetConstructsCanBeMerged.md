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
  its terminator inclusive (`;`, or the closing tag when the statement is ended
  by `?>` in template code).
- Severity: info.
- Message: `Consecutive unset() calls; merge them into one.`

## Fix

- **F1** Replace the previous `unset` statement with
  `unset(<args>);` where `<args>` is the previous statement's arguments followed
  by the current statement's arguments, each in its original source text,
  joined by `, `. Then delete the current statement (the `;` included).
  Comments that sat between the two statements stay where they were (now after
  the merged statement); whitespace around the deleted statement may remain
  (comparison is whitespace-normalized).
- **F1a** When the merged statements were written on several lines and
  the single merged line, with its indentation, would be longer than 120
  characters, the merged statement keeps one line per original statement
  instead: `unset(`, then each original statement's arguments joined by
  `, ` on its own line, indented one level deeper than the statement (a tab
  when the statement is indented with tabs, four spaces otherwise) and
  followed by `,` except the last, then `);` at the statement's
  indentation. No trailing comma is added (PHP before 7.3 rejects it).
- **F2** Applying the fixes of a run one after another, in source order,
  accumulates everything into the first statement:
  `unset($a); unset($b); unset($c);` → `unset($a, $b, $c);`.
- Spacing quirks of the original (`unset ($x)`) are normalized to `unset(`
  because the merged statement is regenerated.

- **F3** The fix is withheld (the report stays) for every statement of a run
  when any statement of that run
  - is not terminated by `;` — a statement ended by a closing tag owns that
    tag, so regenerating or deleting it would drop the tag and leave the
    following opening tag or inline HTML inside PHP code; or
  - contains a comment anywhere inside it (typically within the parentheses):
    regeneration keeps only the argument texts and would lose the comment.
  - other than the last of the run, is followed by a comment on the same line
    (`unset($a); // why`): deleting the statement would append that comment
    to the merged line, away from the argument it describes. A comment on a
    line of its own between the statements stays where it is, unless a
    blank line also separates the two statements: the comment then opens a
    paragraph of its own (`// Unset old attributes.` above the statements
    it describes), which the merge would leave describing nothing.

  The decision is taken for the run as a whole so that a partial merge never
  reorders the arguments around a statement left in place.

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

- Upstream skips plain comments implicitly and doc comments explicitly;
  custos skips all comments, which is the same observable behavior.
- **F3** is custos-only: upstream offers the merge regardless of how the
  statements are terminated and of comments inside or after them, which can produce
  unparsable template code or silently drop comments. custos keeps the report
  and offers no fix in those runs.
