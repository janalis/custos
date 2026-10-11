---
id: UnescapedHtmlOutput
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnescapedHtmlOutput

## Summary

Detect request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping.

## Detection

- D1. Report request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping. Both `echo` and `print` are output statements.
- D1a. The context is proven from the bytes one statement emits before its first dynamic part: concatenated literals, the literal fragments of a double-quoted string or heredoc (escape sequences decoded), and the earlier operands of a comma-separated `echo`. Those bytes must close an ordinary tag and leave no tag, attribute, comment, `script` or `style` element open. Only that first dynamic part is examined. Later parts of the same operand must be free of side effects, because the whole operand is evaluated before any of it is emitted; later `echo` operands are not restricted.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Trusted literals, htmlspecialchars/htmlentities with correct charset/flags and unknown output context are excluded. Markup emitted by an earlier statement or by inline HTML outside the statement does not establish a context.
- E1a. No report when the latest dominating literal `Content-Type` header in the same scope selects an inert media type: `text/plain`, `text/csv`, `application/json`, any `+json` type or `application/octet-stream`. Every other type, including `image/svg+xml` and `application/xhtml+xml`, can execute injected markup and does not suppress the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.
- E4. A value that only selects between results is judged by those results: a `match` or ternary is request-derived only when one of its arm results is; the subject and the arm conditions are not emitted. This also applies to a variable whose single reaching assignment is such an expression.
- E5. The emitted value is numeric or boolean, and therefore not reported, when it is the result of binary `+`, `-`, `*`, `/`, `%` or `**` (an array union is emitted as the constant text `Array`), of unary `-` or `+`, of a resolved builtin returning only an integer or boolean (`strlen`, `mb_strlen`, `count`, `sizeof`, `crc32`, `ord`, the `strcmp` family, the `strpos` family, `levenshtein`, `in_array`, `array_key_exists`, `strtotime`, every `is_*` and `ctype_*`), or of `sprintf` with a literal format whose conversions are all numeric (`b d e E f F g G o u x X`, plus `%%`). A format using `%s`, `%c` or a custom padding character keeps the taint. The proof covers only the emitted value itself: the same arithmetic passed on to another call, such as `implode`, stays request-derived.
- E6. No report when a request superglobal feeding the value (or `$GLOBALS`) is rewritten earlier in the same scope: a plain assignment of a value without request text to it or to one of its elements, a reference taken from it, `++`/`--`, `unset`, a `foreach` that writes into it or iterates it by reference, or a positional argument bound to a by-reference parameter of a resolved callee that may change it. A write to an element with a literal key counts only when the value reads that same element, the whole superglobal, an element with a non-literal key, or the superglobal through a callee; a write with a non-literal or nested key always counts. These do not count: an assignment that stores a value still holding request text (`$_GET['v'] = trim($_GET['v'])`), a compound assignment to an element, a write to another literal element, writes in nested closures, writes after the output statement, reads such as `count($_GET)`, and builtins that take the array by reference only to read it or move its pointer (`extract`, `current`, `key`, `end`, `reset`, `next`, `prev`, `each`).
- E7. The narrowing of E4 and E5 through a variable's reaching assignment is not applied, and the report is kept, when the scope can rebind that variable behind the assignment: a reference alias, a by-reference closure capture or argument, a by-reference `foreach` over it, a `catch` binding or `static` declaration of the same name, `extract()`, a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Escape untrusted text for its HTML output context.`

## Fix

None. Wrapping the value in `htmlspecialchars` is not behaviour-preserving in general: the page charset cannot be proven from one output statement, so a forced `UTF-8` with `ENT_SUBSTITUTE` may rewrite bytes of a page in another encoding, and an array-valued request parameter, emitted today as `Array`, would raise a `TypeError` on PHP 8. The escaping call, its flags and its charset are left to the author.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'] . '</p>';</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<li>{$_POST['title']}</li>";</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<td>', $_GET['cell'], '</td>';</error>
<error descr="Escape untrusted text for its HTML output context.">print '<p>' . $_COOKIE['theme']</error>;
```

Not reported:

```php
<?php
$order = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
echo '<p>' . $order . '</p>';
echo '<p>' . strlen($_GET['q']) . ' characters</p>';
echo '<p>' . ($_GET['page'] + 1) . '</p>';

function export()
{
    header('Content-Type: text/csv');
    echo '<p>' . $_GET['label'];
}

function cleaned()
{
    $_GET['label'] = 'fixed text';
    echo '<p>' . $_GET['label'];
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
