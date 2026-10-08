---
id: StrTrUsageAsStrReplace
group: Control flow
kind: semantic
needs: [names]
php: { min: "", max: "" }
---

# StrTrUsageAsStrReplace

## Summary
`strtr($subject, $from, $to)` with single-character `$from` and `$to` is really
a plain character replacement; `str_replace($from, $to, $subject)` states that intent
more clearly.

## Detection
- **D1** A function call (not a method/static call) whose name, as written in
  its last segment, is `strtr` compared case-insensitively like PHP function
  names (`StrTr`, `\STRTR` match), and that resolves to the global function:
  written with a single leading `\`, or unqualified with no function named
  `strtr` declared or imported in the current namespace. `Ns\strtr(...)` and
  shadowing user functions are not matched.
- **D2** It has exactly **3** arguments (`subject`, `from`, `to`).
- **D3** The second argument resolves to a single string literal `L`:
  - if the argument itself is a string literal, `L` is that literal;
  - otherwise compute the argument's *possible-values set* (below) and keep
    only the string-literal members; if exactly one remains, that is `L`;
    otherwise (zero or several) no report.
- **D4** The raw source contents of `L` (the characters between the quotes,
  escape sequences **not** decoded) are non-empty, at most 2 characters long,
  and:
  - for a single-quoted literal: either exactly one character (any character
    other than a line terminator), or a backslash followed by `\` or `'`
    (i.e. `'\\'`, `'\''`);
  - for a double-quoted literal (and any non-single-quoted literal): either
    exactly one character (other than a line terminator), or a backslash
    followed by one of `\`, `"`, `$`, `r`, `n`, `t`
    (i.e. `"\\"`, `"\""`, `"\$"`, `"\r"`, `"\n"`, `"\t"`).
  "One character" means one byte: a multi-byte UTF-8 character such as `é`
  does not qualify (`strtr` would map each of its bytes separately).
- **D5** The third argument (`to`) resolves to a single string literal by the
  same procedure as D3, and its raw contents satisfy the same rule as D4
  (exactly one character, or one of the listed two-character escapes).
  Otherwise no report: `strtr` only uses as many characters of `to` as `from`
  has, while `str_replace` would insert all of `to` (or delete, when `to` is
  empty).

### Possible-values set
Collect candidate value expressions of an expression, recursively, visiting
each expression at most once, with surrounding parentheses removed first:
- ternary → values of both branches (short ternary: the false branch, plus
  the true branch when present);
- `??` → values of both operands;
- variable → if inside a function-like scope (function, method, closure,
  arrow function): the default value of a same-named parameter of that scope
  (if any), plus, for every plain `=` assignment to the same variable anywhere
  in the scope body (including nested closures), the values of the right-most
  assigned value (through chains `$a = $b = v`). Outside any function scope
  the variable yields nothing. **Unstable variable** (custos refinement, see Divergences): if `$v` is also the operand of `++`/`--` or the target of a compound assignment (`+=`, `.=`, `??=`, …) anywhere in that scope body, the whole result is *unknown* → no report;
- property fetch → the resolved property's default value (unless the
  default's text ends with the property name), plus plain assignments to the
  same property expression in the current scope and in the class constructor;
- class constant → values of its declared value;
- global constant other than `true`/`false`/`null` → its defined value (if
  resolvable; nothing otherwise);
- anything else → the expression itself.

## Exceptions (no report)
- **E1** Argument counts other than 3 (`strtr($s, $map)` with an array map).
- **E2** `from` longer than one character (`'ab'`), or empty (`''`), or a
  multi-byte character (`'é'`).
- **E7** `to` that is empty (`''`), longer than one character (`'USD'`), a
  multi-byte character, or not resolvable to exactly one string literal.
- **E3** Single-quoted two-character contents that are not `\\`/`\'`, e.g.
  `'\t'` (backslash + t is two literal characters in single quotes).
- **E4** Double-quoted two-character contents with an escape outside the
  allowed set, e.g. `"\'"`, `"\0"`, `"\e"`, or interpolation like `"$x"`.
- **E5** `from` that cannot be resolved to exactly one string literal.
- **E6** Calls resolving to a user function named `strtr` (D1).

## Report
- Range: the whole call expression, from the start of the function name
  (including any leading `\` or namespace qualifier) to the closing `)`.
- Severity: info (weak warning).
- Message: `Use '{replacement}' instead.` where `{replacement}` is the F1
  text.

## Fix
- **F1** Replace the call with
  `{qualifier}str_replace({from}, {to}, {subject})`, where:
  - `{qualifier}` is `\` for `\strtr`, and also for a plain `strtr` when an
    unqualified `str_replace` at that position would not reach the global
    function (a `use function` import under that name, or a same-named function declared in the current namespace); otherwise empty (D1 already
    requires the call to reach the global `strtr`);
  - `{from}`, `{to}`, `{subject}` are the **original source texts** of the
    2nd, 3rd and 1st arguments respectively (verbatim, e.g. a variable name
    stays a variable even when it was resolved through D3);
  - arguments are joined with `, ` (comma + one space).
  Examples:
  - `strtr($name, '-', '_')` → `str_replace('-', '_', $name)`
  - `\strtr($row, "\t", ';')` → `\str_replace("\t", ';', $row)`

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function slug($title, $glue = '+') {
    $a = <weak_warning descr="Use 'str_replace('-', '_', $title)' instead.">strtr($title, '-', '_')</weak_warning>;
    $b = <weak_warning descr="Use '\str_replace(&quot;\t&quot;, ';', $title)' instead.">\strtr($title, "\t", ';')</weak_warning>;
    $c = <weak_warning descr="Use 'str_replace('\'', '`', $title)' instead.">strtr($title, '\'', '`')</weak_warning>;
    $d = <weak_warning descr="Use 'str_replace(&quot;\$&quot;, 'S', $title)' instead.">strtr($title, "\$", 'S')</weak_warning>;
    $e = <weak_warning descr="Use 'str_replace($glue, ' ', $title)' instead.">strtr($title, $glue, ' ')</weak_warning>;

    $f = strtr($title, '--', '_');
    $k = strtr($title, '-', '');
    $l = strtr($title, '$', 'USD');
    $g = strtr($title, '\r', '_');
    $h = strtr($title, "\0", '_');
    $i = strtr($title, ['-' => '_']);
    $j = strtr($title, '', '_');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l];
}
```

```php
<?php
function slug($title, $glue = '+') {
    $a = str_replace('-', '_', $title);
    $b = \str_replace("\t", ';', $title);
    $c = str_replace('\'', '`', $title);
    $d = str_replace("\$", 'S', $title);
    $e = str_replace($glue, ' ', $title);

    $f = strtr($title, '--', '_');
    $k = strtr($title, '-', '');
    $l = strtr($title, '$', 'USD');
    $g = strtr($title, '\r', '_');
    $h = strtr($title, "\0", '_');
    $i = strtr($title, ['-' => '_']);
    $j = strtr($title, '', '_');
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l];
}
```

## Divergences
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- **`to` length — custos diverges from upstream** (D5). Upstream reports and
  rewrites whatever `to` is. With a one-character `from`, `strtr` uses only
  the first character of a longer `to`, and an empty `to` turns the call
  into a no-op, whereas `str_replace` inserts the whole string or deletes the
  match. custos only reports when `to` is also exactly one character.
- **Multi-byte characters — custos diverges from upstream** (D4). Upstream
  counts a UTF-8 character such as `é` as one character. `strtr` translates
  bytes, so such a call maps each byte on its own (affecting other characters
  that share those bytes) and is not a substring replacement; custos
  requires single-byte `from`/`to`.
- **Heredoc/nowdoc literals — custos diverges from upstream** (D4/D5).
  Upstream treats anything not single-quoted with the double-quoted rule.
  custos applies PHP's own escape rules: a heredoc body is read like a
  double-quoted string except that `\"` is two characters (no escape in a
  heredoc), and a nowdoc has no escapes at all, so `\'`, `\\` or `\n` in a
  nowdoc are two characters (`strtr` would then map only the backslash,
  not replace the pair) and are not reported. One plain character is
  accepted in both.
- **Case of the name (custos diverges from upstream).** Upstream matches
  `strtr` case-sensitively, so `STRTR($s, '-', '_')` is not reported
  although PHP calls the same function. custos compares the name
  case-insensitively (D1).
- **Namespace resolution (custos diverges from upstream).** Upstream matches
  the last name segment only, so `Ns\strtr(...)` or an unqualified call to a
  `strtr` declared in the current namespace was reported and rewritten to
  the global `str_replace`. custos only matches calls resolving to the global
  `strtr` (D1).
- **Builtin spelling (custos diverges).** Upstream copies the written
  qualifier and inserts `str_replace` bare, so a namespaced or imported
  `str_replace` captures the rewritten call. custos writes `\str_replace(`
  in that case (F1).
