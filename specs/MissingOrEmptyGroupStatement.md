---
id: MissingOrEmptyGroupStatement
group: Code style
kind: semantic
needs: [types, index]
php: { min: "", max: "" }
---

# MissingOrEmptyGroupStatement

## Summary

Control structures should always use a braced block for their body. A body
written as a bare statement is easy to break when a second line is added; a
braced block that contains nothing is usually leftover or unfinished code.

## Detection

Checked constructs: `if`, `elseif`, `else`, `foreach`, `for`, `while`,
`do … while`. Each branch of an if-chain (`if`, every `elseif`, the `else`) is
checked on its own.

"Body block" means: the construct's own body is a braced block `{ … }` (a
group statement). For an `if`, only the `if`-branch body counts — the
`elseif`/`else` branches are separate constructs and do not provide a block
to the `if`.

- D1 (missing block): the construct has no body block, i.e. its body is a
  single non-braced statement (including the empty statement `;`, a nested
  control structure, an expression statement, …). Report.
- D2 (empty block, only when `REPORT_EMPTY_BODY` is true): the construct has a
  body block that contains zero statements. When counting statements, comments
  of any kind (line, block, doc comments) are ignored, so a block holding only
  comments is empty. The empty statement `;` counts as a statement (so `{ ; }`
  is not empty — see Divergences).

## Exceptions (no report)

- E1: `else` whose body is directly an `if` statement (the `else if` form),
  even when comments sit between `else` and `if`. The nested `if` itself is
  still checked normally (its own body may be reported by D1/D2).
- E2: the file name ends with `.blade.php` — nothing is reported in such files.
- E3: `switch`, `try`/`catch`/`finally`, `declare`, functions, classes and other
  constructs are not checked.
- E4: alternative syntax bodies (`if (…): … endif;`, `foreach (…): … endforeach;`,
  `while (…): … endwhile;`, `for (…): … endfor;`, `else:` / `elseif (…):`)
  are not reported by D1 (see Divergences).
- E5: a `foreach` whose subject is known to be an iterator — typed
  `iterable`, `\Traversable`, or a class or interface extending or
  implementing it (`\Iterator`, `\Generator`, a generator function's
  result) — is not reported by D2: looping over it with an empty body runs
  the iterator for its effects (draining a stream that must throw, warming
  a lazy collection). A subject of unknown type or an array is still
  reported.
- E6: a `foreach` run for the last entry it leaves is not reported by D2:
  its key or value target is not a plain variable (`as $stub->class =>
  $stub->position`, a destructuring list), or a key or value variable is
  mentioned after the loop in the same function (or file), or a
  variable-variable is.

- E7: an empty `if` or `elseif` body that holds a comment and is followed
  by another `elseif` or an `else` of the same chain is not reported by
  D2: the commented branch is there to keep the rest of the chain from
  running (`// Intentionally empty`).

## Report

- Range: the construct's leading keyword token only — `if`, `elseif`, `else`,
  `foreach`, `for`, `while`, or `do` (for do-while, the `do` keyword, never the
  trailing `while`). For `else if`, the inner `if` keyword is the one that may
  be reported.
- Severity: info (weak warning).
- Message (D1): "Use a braced block for the body of this construct."
- Message (D2): "This construct has an empty body block."

## Fix

- F1 (D1 only): replace the body statement (exactly its text range, including
  its trailing `;`) with a braced block containing that statement:
  `{` + newline + original statement text + newline + `}`. Text before the body
  (keyword, condition, the whitespace between `)` and the body) is preserved, so
  `while ($q) ;` becomes `while ($q) {\n    ;\n}` (indentation free; the
  compared output only needs whitespace between `{`, the statement and `}`).
  For `do stmt; while (c);` only `stmt;` is wrapped; the `while (c);` tail is
  preserved.
- F1a (custos): when the construct and its body sit on one line that a
  close tag ends (`<?php if ($a) print "x" ?>` inside an HTML attribute),
  the block stays on that line: `if ($a) { print "x"; } ?>`.
- D2 has no fix.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| REPORT_EMPTY_BODY | bool | true | Reports empty braced bodies. Missing braces are always reported. |

## PHP versions

No gating.

## Examples

```php
<?php
function demo($flag, $rows, $n) {
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($flag) log_it();
    <weak_warning descr="Use a braced block for the body of this construct.">elseif</weak_warning> ($n > 3) $n--;
    <weak_warning descr="Use a braced block for the body of this construct.">else</weak_warning> $n = 0;

    if ($flag) { tick(); }
    else <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($n) tock();

    <weak_warning descr="Use a braced block for the body of this construct.">foreach</weak_warning> ($rows as $row) emit($row);
    <weak_warning descr="Use a braced block for the body of this construct.">for</weak_warning> ($i = 0; $i < $n; $i++) ;
    <weak_warning descr="Use a braced block for the body of this construct.">while</weak_warning> (more()) step();
    <weak_warning descr="Use a braced block for the body of this construct.">do</weak_warning> pull(); while (pending());
}
```

```php
<?php
function demo($flag, $rows, $n) {
    if ($flag) {
        log_it();
    }
    elseif ($n > 3) {
        $n--;
    }
    else {
        $n = 0;
    }

    if ($flag) { tick(); }
    else if ($n) {
        tock();
    }

    foreach ($rows as $row) {
        emit($row);
    }
    for ($i = 0; $i < $n; $i++) {
        ;
    }
    while (more()) {
        step();
    }
    do {
        pull();
    } while (pending());
}
```

With `REPORT_EMPTY_BODY` = true (default):

```php
<?php
<weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($queue as $job) {
    // nothing yet
}
<weak_warning descr="This construct has an empty body block.">while</weak_warning> (poll()) {}
if ($ready) { start(); } <weak_warning descr="This construct has an empty body block.">else</weak_warning> { }
if ($ready) { ; }            // not empty: contains an empty statement
```

In `views/page.blade.php` nothing is reported (E2).

## Divergences

- **Draining iterators (custos diverges, E5).** Upstream reports every
  empty body. An empty `foreach` over an iterator is how a test drains a
  response stream that must fail (Symfony's HTTP client test case), so
  custos does not report it when the subject is known to be one.

- **Last-entry loops (custos diverges, E6).** Upstream reports every
  empty body. An empty `foreach` whose target is a property, or whose
  variable is used afterwards, exists to leave the last key or value
  behind (Symfony's VarDumper data cloner).

- **Commented skip branches (custos diverges, E7).** Upstream reports
  every empty body. A commented empty `if` before an `elseif`/`else` is a
  deliberate case that stops the chain (Drupal's field group, Symfony's
  translation bundle).

- Alternative syntax (E4): upstream behavior depends on how the IDE models the
  colon-syntax body; it is not covered by upstream fixtures. custos never
  reports D1 for colon syntax (adding braces there is not the intended
  style); for D2, a colon-syntax body with zero statements counts as empty.
- Whether `{ ; }` counts as empty is not covered by upstream fixtures; we count
  `;` as a statement (not empty).
- **Formatting (custos).** Upstream relies on the IDE formatter; the fix now
  produces formatted code itself: the opening brace joins the header line
  (replacing the whitespace before the body), the body is indented one level
  under the construct's line (a tab when that line is indented with tabs,
  else the step from the nearest enclosing statement indented less, else the
  file's own space indentation step, four spaces when it has none; a body
  already on a line of its own, indented deeper than the construct, keeps
  its indentation) and the closing brace aligned with it
  (`if ($a)\n    stmt;` → `if ($a) {\n    stmt;\n}`).
- **Line comment after the header (custos diverges, F1).** In `for (…) //
  note` followed by the body on the next line, a brace placed after the
  comment is commented out (the fix produced unbalanced braces and
  re-fired on every pass; Magento's Redis session vendor code). The block
  is opened right after the header (`for (…) { // note`), the body stays
  in place and the closing brace follows it.
- **Statements ended by a close tag (custos).** `if ($h) echo "x" ?>` has
  no `;`: the close tag ends the statement. The braced body gets a `;` and
  the closing brace goes before the tag (`{ echo "x"; } ?>`); placed after
  it, the brace landed in the HTML (phpxmlrpc debugger, Moodle vendor).
  Comments between the statement and the tag stay after the closing brace
  (`echo "x" // why ?>` → `echo "x"; } // why ?>`): a `;` or brace written
  after a line comment would be commented out. The same holds when a line
  comment follows the header (`if ($h) // note` + `echo "x" ?>` →
  `if ($h) { // note` + `echo "x"; } ?>`).
- **Close tag as the body, after a line comment (custos, F1).** In
  `for (…) # note ?>` the close tag is the empty body. The empty block is
  written right after the header (`for (…) {} # note ?>`); placed before
  the tag it sat inside the comment, so the report never cleared and each
  fix pass added another block.
