---
id: NestedPositiveIfStatements
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# NestedPositiveIfStatements

## Summary

An `if` that is the only statement inside another `if` (without alternative
branches) can be merged into its parent with `&&`; an `if` that is the only
statement inside an `else` block can become `else if`. Both remove a nesting
level.

## Detection

Applies to an `if` statement `C` (the inner one) whose direct parent is a
braced block `{ … }` (group statement). "Statement count" of a block ignores
comments (line, block and doc comments); the empty statement `;` counts.

### Case A — inner if inside parent if

- D1: the braced block is the body of the `if`-branch of an `if` statement `P`
  (not the body of an `elseif` or `else`).
- D2: neither `P`'s condition nor `C`'s condition is, at top level, a binary
  `||` or `or` expression. (Only the top-level node is inspected: a condition
  that is a parenthesised expression, e.g. `if (($a || $b))`, is not an
  `||` node and does not block.)
- D3: `P`'s body block contains exactly one statement (namely `C`).
- D4: neither `P` nor `C` has any `elseif` branch.
- D5: else-compatibility:
  - if `C` has no `else`, then `P` must have no `else`;
  - if `C` has an `else`, then `P` must have an `else` too, both `else` bodies
    must be braced blocks, they must have the same statement count, and they
    must be structurally equivalent (same token sequence ignoring whitespace
    and comments). (An `else if` branch has no braced block → not mergeable.)
- D5b: `C`'s own body is not an alternative-syntax (colon) block.
- When D1–D5b hold, report `C`.

### Case B — inner if inside else

- D6: the braced block is the body of an `else` branch.
- D7: that block contains exactly one statement (namely `C`).
- No condition, `elseif`, or `else` restrictions apply to `C` in this case.
- Report `C`.

## Exceptions (no report)

- E1: the parent `if`'s body is not braced, or the inner `if` sits in an
  `elseif` body.
- E2: the parent body contains other statements besides `C` (comments do not
  count).
- E3: Case A with an `||`/`or` top-level condition on either `if`.
- E4: Case A where the parent has an `else` but the inner `if` does not, or
  vice versa, or both have `else` branches that differ.
- E5: Case A where either `if` has an `elseif`.
- E6: alternative (colon) syntax — a colon-syntax block is never the parent
  body, and in Case A a colon-syntax inner `if` is not reported either (its
  body has no closing keyword of its own, so it cannot replace the parent's
  braced block). Case B still reports a colon-syntax inner `if`: the whole
  statement, `endif;` included, moves after `else`. See Divergences.

## Report

- Range: the inner `if` keyword token of `C` (just `if`).
- Severity: info (weak warning).
- Message: "Merge this if statement into its parent construct."

## Fix

Comment preservation (both cases): comments located inside the parent block
before `C` (between `{` and `C`) are moved into `C`'s body block, immediately
after its opening `{`, keeping their original order, and before `C`'s existing
body content. This only happens when `C`'s body is a braced block.

- F3 (custos) No fix is offered (the report stays) when it would move a
  directive for the line after it (a comment holding `@phpstan-ignore`,
  `@psalm-suppress`, `phpcs:ignore`, `@noinspection` or `@custos-ignore`)
  from before `C` into its body, where it would act on another statement,
  or when it would drop a comment: one before `C` whose body is not braced, one after `C` inside the
  parent block, or, in Case A, one inside `C` outside its condition and body
  (in its discarded `else`, or between its parts).

- F1 (Case A): with `PC` = parent condition text and `CC` = child condition
  text (both the expression inside the `if ( … )` parentheses):
  - `PC` is wrapped as `(PC)` whenever its right edge would absorb the
    appended `&& CC`: the parent condition is an assignment, a ternary
    (including the short `?:` form), an `instanceof`, a binary expression
    whose operator binds as loosely as `&&` or looser (`||`, `??`, `and`,
    `or`, `xor`, `|>`), a `print`, `include`/`require`, `throw`,
    `yield`/`yield from` or arrow function, or it ends in one of those: a
    prefix operator (`!`, `-`, `@`, a cast, legacy `clone`) applied to such
    an operand (`!$row = load()`), or a binary expression of any other
    operator (`&&`, `==`, `.`, …) whose right operand is one
    (`$a && $row = load()`, `$a == $row = load()`). Otherwise (plain `&&`
    chain, comparison, variable, call, `!$a`, parenthesised expression, …)
    it is used as is: `if ($a == $b) { if ($c) {} }` → `if ($a == $b && $c) {}`.
  - `CC` is wrapped as `(CC)` when it binds looser than `&&`: a ternary
    (either form) or a binary `??` or `xor`. An `and` child keeps its meaning
    unwrapped and an assignment child is already the whole right operand, so
    both are used as is.
  - The parent condition is replaced by `PC && CC` (single spaces around
    `&&`).
  - The parent's body block is replaced by `C`'s body (the statement after
    `C`'s condition: its braced block, with preserved comments, or its single
    statement). `C`'s `else` (if any) is discarded; `P`'s `else` (if any) is
    kept unchanged.
  - Example: `if ($u) { if ($v) { go(); } }` → `if ($u && $v) { go(); }`;
    `if ($u = load()) { if ($v) {} }` → `if (($u = load()) && $v) {}`;
    `if (!$r = load()) { if ($v) {} }` → `if ((!$r = load()) && $v) {}`;
    `if ($u) { if ($a ?? $b) {} }` → `if ($u && ($a ?? $b)) {}`.
- F2 (Case B): the `else` branch's braced block (from `{` to `}`) is replaced
  by the full text of `C` (including `C`'s own `elseif`/`else` branches), with
  preserved comments inserted into `C`'s body. The text between `else` and
  `{` is kept, giving `else if (…) …`.
- Indentation (both cases): the moved text (`C`'s body in Case A, `C` in
  Case B) loses one level: each of its lines after the first that starts
  with the indentation of `C`'s line is moved to the indentation of the
  parent block's opening line instead. The text is kept as is when a line
  break lies inside a string, heredoc or inline HTML within it.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function merge_cases($u, $v, $w) {
    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) {
            go();
        }
    }

    if ($u && $v) {
        // explain
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w > 2) {
            run();
        }
    }

    if ($u ?? $v) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($w) { run(); }
    }

    if ($u) {
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) { run(); } else { stop(); }
    } else { stop(); }

    if ($u) {
        if ($v) {}
    } else {
        stop();
    }

    if ($u || $v) {
        if ($w) {}
    }

    if ($u) {
        if ($v) {}
        else { stop(); }
    }
}

function else_cases($u, $v) {
    if ($u) {
        run();
    } else {
        // fallback
        <weak_warning descr="Merge this if statement into its parent construct.">if</weak_warning> ($v) {
            stop();
        } else {
            halt();
        }
    }
}
```

```php
<?php
function merge_cases($u, $v, $w) {
    if ($u && $v) {
            go();
        }

    if ($u && $v && $w > 2) {
        // explain
            run();
        }

    if (($u ?? $v) && $w) { run(); }

    if ($u && $v) { run(); } else { stop(); }

    if ($u) {
        if ($v) {}
    } else {
        stop();
    }

    if ($u || $v) {
        if ($w) {}
    }

    if ($u) {
        if ($v) {}
        else { stop(); }
    }
}

function else_cases($u, $v) {
    if ($u) {
        run();
    } else if ($v) {
        // fallback
            stop();
        } else {
            halt();
        }
}
```

## Divergences

- **Comments (custos diverges, F3).** Upstream's fix drops the comments it
  cannot place (after the inner `if`, in its discarded `else`). custos
  reports those cases without a fix rather than delete them (composer/semver,
  monolog, php-cs-fixer).
- Upstream never wraps the child condition. When the child condition binds
  looser than `&&` (ternary, `??`, `xor`), the merged condition changes
  meaning (`$u && $a ? $b : $c`), so the fix parenthesises such a child.
  Not covered by upstream fixtures.
- Upstream only parenthesises a parent condition whose top-level node is an
  assignment, ternary or non-`&&` binary. A parent that merely ends in such
  an expression (`!$row = load()`, `@$u = $v`, `$a && $row = load()`,
  `print "p"`) would swallow the appended `&& CC`; the fix parenthesises
  those parents as well. Not covered by upstream fixtures.
- Alternative (colon) syntax: not covered by upstream fixtures. Only braced
  blocks qualify as the parent body (colon-syntax `if`/`else` bodies are
  never reported), and a colon-syntax inner `if` directly inside a parent
  `if` is not reported because merging it would drop its `endif`.
