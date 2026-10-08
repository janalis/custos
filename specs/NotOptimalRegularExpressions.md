---
id: NotOptimalRegularExpressions
group: Performance
kind: semantic
needs: [names, index, types]
php: { min: "", max: "" }
---

# NotOptimalRegularExpressions

## Summary
A family of checks on PCRE patterns passed to the `preg_*` functions: broken
or missing delimiters, invalid / pointless / missing modifiers, class
spellings that have a shorter escape, redundant or backtracking-prone
constructs, and calls where a plain string function (`strpos`,
`str_replace`, `trim`, `explode`, a `===` comparison) does the same job
faster. Only a few findings carry a fix (the plain-function replacements).

## Detection

### Entry point and pattern extraction
- **D1** A plain function call (not a method or static call) that resolves to
  the global function (names compared case-insensitively, as PHP does; `\` and
  a global `use function` import are fine, but a same-named function declared
  in the current namespace, one imported from another namespace, or a
  qualified non-global name such as `Foo\preg_match` does not count): one of
  `preg_filter`, `preg_grep`, `preg_match_all`, `preg_match`,
  `preg_replace_callback`, `preg_replace`, `preg_split`, `preg_quote`
  (`\preg_match` and `PREG_MATCH` qualify). Call it `C`, its (lower-case) name
  `fn`. `C` must have at least one argument; let `A` be the first one.
- **D2** Candidate literals:
  - if `A` is an array literal (`[...]` or `array(...)`), every element's
    value (the value part of `k => v` elements) is resolved with *literal
    resolution* (below); each success is a candidate. `C` is then in
    **array mode**;
  - otherwise `A` itself is resolved with literal resolution; a success is
    the single candidate.
  *Literal resolution* of an expression `e`: if `e` is a string literal, use
  it. Otherwise run *value discovery* on `e` and keep only the string-literal
  results; exactly one string literal → use it (non-literal results are
  ignored), zero or several → no candidate.
  *Value discovery* (visited-set against cycles; a revisited node yields
  nothing): strip parentheses; ternary `c ? a : b` → union of discovery on
  `a` and `b`; `a ?? b` → union of discovery on both operands; local variable
  `$v` → nothing at top level, otherwise (nearest enclosing function, method
  or closure) the discovery of the default value of the same-named parameter
  (if any) plus the discovery of the right-hand side of every plain `$v = …`
  assignment anywhere in that body (chained `$v = $w = X` → `X`; a variable
  that is also the operand of `++`/`--` or the target of a compound
  assignment anywhere in that body makes the whole result *unknown* → no
  candidate, custos refinement, see Divergences); property fetch → the property's default value
  (unless its source text ends with the property name) plus plain
  assignments to an equivalent fetch in the current function and in the
  declaring class's constructor; class constant → discovery of its value;
  global constant other than `true`/`false`/`null` → the value expression of
  its `define()` (not discovered further), unresolved → nothing; anything
  else → the expression itself.
- **D3** A candidate literal `L` is processed only if it is located in the
  same file as `C`, contains no interpolation (single-quoted, or
  double-quoted without embedded variables/expressions), and its raw
  contents `R` (the source text between the quotes, escape sequences **not**
  decoded — `'\\d'` gives the three characters `\\d`) are non-empty. Empty
  contents (`''`) → nothing at all for that literal.
- **D4** Delimiter parsing of `R`, as PHP does it. Leading whitespace
  bytes (space, tab, `\n`, `\v`, `\f`, `\r` as they appear in the raw
  contents) are skipped first. If the first remaining character is an ASCII
  letter, an ASCII digit, `\` or NUL — delimiters PHP rejects — or nothing
  remains, parsing fails. Otherwise the first rule that matches wins:
  1. first character `δ` is anything except `{`, `<`, `(`, `[`; `R` ends with a later
     occurrence of `δ` followed by zero or more ASCII letters only. The
     closing `δ` is the **last** occurrence of `δ` such that only ASCII
     letters follow it (`/a/b/c` → body `a/b`, modifiers `c`;
     `/x/i/` → body `x/i`, no modifiers; `  #a#` → body `a`);
  2. `{` … `}` + optional ASCII letters (closing `}` chosen the same way);
  3. `<` … `>` + optional letters;
  4. `(` … `)` + optional letters;
  5. `[` … `]` + optional letters.
  Line breaks inside `R` are allowed. On success: `body` = text between the
  delimiters (may be empty), `mods` = the trailing letters (possibly none).
  Then run the pattern checks D6–D19 on (`fn`, `body`, `mods`, `L`) and, if
  `C` is **not** in array mode, the call checks D20–D23.
- **D5** In array mode the call checks D20–D23 never run, even for elements
  that parse fine. Every candidate of the array is checked independently
  (see Divergences for upstream's handling of several elements).

Findings on the literal use `L`'s own location — when `L` was reached via
value discovery that is the definition site (e.g. the assignment). Each
literal finding (same range, severity and message) is reported once per
file: when several calls reach the same literal, a later call does not
repeat a finding an earlier call (in source order) already made on it; a
finding only the later call produces (e.g. D6/D8/D13 after an earlier
`preg_quote`) is still reported. Call findings (D20–D23) are per call.

Notation for the checks: "contains" = substring test; counts of
substrings are non-overlapping left-to-right occurrence counts; "line
terminator" = `\n`, `\r`, U+0085, U+2028, U+2029.

### Missing delimiters
- **D6** D4 fails (invalid delimiter, or no rule matches) and
  `fn` ≠ `preg_quote` → report `L` (warning). No other check runs on that
  literal: `'abca'` and `'1x1'` get only this report (PHP rejects them before
  compiling anything).

### Modifier checks (run whenever D4 succeeded)
- **D7** `mods` contains `e` → report `L` (error) — the eval modifier is
  gone. Applies to every function in D1 (including `preg_quote`) and at
  every PHP level.
- **D8** `fn` ≠ `preg_quote` and `mods` non-empty: for **each character** of
  `mods` not in the allowed set → one report on `L` per offending character
  (error; `/k/QQ` gives two). Allowed set: `eimsuxADJSUX`; plus `n` when the
  PHP level is ≥ 8.2; plus `n` and `r` when ≥ 8.4.
- **D9** `mods` contains `D`:
  - **D9a** `mods` also contains `m` → report `L` (info: `D` is ignored
    under `m`);
  - **D9b** independently, `body` non-empty and
    count(`$`) − count(`\$`) = 0 → report `L` (info: `D` has nothing to
    act on). `[\\$]` counts as escaped and is reported.
- **D10** `mods` contains `s`, `body` non-empty: let `N` = `body` with every
  `\.`, `\[`, `\]` removed (scan left to right), then with every bracket
  run removed (a `[`, one or more characters other than `]`, then `]`;
  leftmost first, non-overlapping). If count(`.` in `N`) − count(`\.` in
  `N`) = 0 → report `L` (info: no `.` for `s` to affect).
- **D11** `mods` contains `i`, `body` non-empty: let `N` = `body` with
  every two-character sequence `\` + one of `\ d D w W s S` removed (left to
  right, non-overlapping; so `\\d` loses `\\` and keeps the letter `d`). If
  `N` contains no Unicode letter (any character of general category L,
  including non-ASCII letters; line breaks irrelevant) → report `L`
  (info: nothing case-sensitive).
- **D12** `mods` contains `r` (no PHP-level condition):
  - **D12a** `mods` lacks `i` → report `L` (error);
  - **D12b** `mods` lacks `u` → report `L` (error).
  Both can fire on the same literal (two findings).
- **D13** `mods` lacks `u`, `body` non-empty, `fn` ≠ `preg_quote`:
  - **D13a** `body` contains a non-ASCII character (code point > U+007F) and
    no line terminator → report `L` (error: non-ASCII text needs `u`).
    custos only counts a non-ASCII character that changes meaning without
    `u`: inside a character class, directly followed by a quantifier
    (`*`, `+`, `?`, `{`), or a letter when `mods` contains `i` (see
    Divergences);
  - **D13b** otherwise: remove every `\\` pair from `body` (left to right);
    if the result contains `\p`, `\P` or `\X` and no line terminator →
    report `L` (error: Unicode escapes need `u`). `\\p` (escaped backslash)
    does not count.
  At most one of D13a/D13b per literal.

### Body checks (run whenever D4 succeeded)
- **D14** Short class spellings. `body` non-empty. Let `B'` = `body` with all
  occurrences of `a-zA-Z` replaced by `A-Za-z`, then all `0-9A-Za-z` by
  `A-Za-z0-9`. For **each** key below that `B'` contains, one report on `L`
  (info), naming the key and its shorthand:

  | key | shorthand |
  |---|---|
  | `[0-9]`, `[:digit:]` | `\d` |
  | `[^0-9]`, `[^\d]` | `\D` |
  | `[:word:]`, `[A-Za-z0-9_]` | `\w` |
  | `[^\w]`, `[^A-Za-z0-9_]` | `\W` |
  | `[^\s]` | `\S` |

  The message carries a hint: "equivalent" when `mods` lacks `u`, "broader
  under /u" when it contains `u`. (`[0-9,]` matches no key; `[[:digit:]]`
  matches `[:digit:]`.)
- **D15** Repeated identical classes. `body` contains `[`. Find the first
  (leftmost) run of the shape: `[X]` `q?` `[X]` `q?`, repeated one or more
  times greedily, where `X` is one or more characters other than `]`, both
  bracket groups in one repetition are textually identical, and `q` is one
  of `*`, `+`, `?`, `{…}` (`{`, one or more non-`}`, `}`). Report `L` once
  (info), naming the matched run and the class `[X]` of its last repetition.
  `[ab][ab][ab][ab]` → run is all four; `[ab][ab][ab]` → run is the first
  two; `[ab]{2}[ab]+` → whole; `[ab][ba]` → no.
- **D16** Leading/trailing anything. `fn` starts with `preg_match`
  (`preg_match`, `preg_match_all`), `C` has **exactly 2** arguments, `body`
  non-empty, and count(`\0`) − count(`\\0`) ≤ 0:
  - **D16a** `body` starts with `.*` → report `L` (info);
  - **D16b** `body` ends with `.*` → report `L` (info).
  Both can fire.
- **D17** Redundant escapes in a class. For every bracket group in `body`
  (leftmost-first, non-overlapping: `[`, one or more characters that are
  neither `[` nor `]`, `]`; call the inside `S`):
  - `S` contains both `\w` and `\d` → report `L` (error: `\d` is inside
    `\w`);
  - else `S` contains both `\W` and `\D` → report `L` (error: `\D` is inside
    `\W`).
  One report per qualifying group.
- **D18** Quantified group whose alternative is itself quantified (ReDoS).
  `body` non-empty.
  1. `N` = `body` with every `(?:` replaced by `(`.
  2. Repeat while found: search `N` for: a character other than `\` (call it
     `p`), then a group `(`…`)` whose inside has no `(`/`)` and whose last
     inside character is not `\`, then a character `n` that is neither `+`
     nor `*`. Replace **every** textual occurrence of that whole match by
     `p` + `n` (i.e. the inner group disappears).
  3. Scan `N` left to right, non-overlapping, for: start-of-string or a
     character other than `>`, then `(`, an inside `G` with no `(`/`)`, `)`,
     a quantifier `Q` ∈ {`+`, `*`}, then end-of-string or a character other
     than `+` (that character is consumed by the match).
  4. For each match, split `G` on `|`; the first non-empty alternative that
     is exactly `\` + one of `dDwWsS` + one of `*+` → one report on `L`
     (error), naming the alternative and `Q`.
  Not reported: `{m,n}` outer quantifiers, possessive `)*+`/`)++`, groups
  directly after `>` (atomic `(?>(…)*)`), alternatives like `\d{1,}`.
- **D19** Probably missing `s` around tag content. `mods` lacks `s`, `body`
  contains `>`, `body` has no line terminator, and `body` contains `>`
  immediately followed by `.`, then `*` or `+`, an optional `?`, then `<`
  → report `L` (info).

### Call checks (only when `C` is not in array mode, once per processed literal)
- **D20** `fn` = `preg_quote` with exactly 1 argument → report `C` (whole
  call; warning: pass the delimiter). Note this is only reached when the
  argument resolved to a literal that parses with D4 (e.g. `'#a.b#'`);
  `preg_quote('a.b')` or `preg_quote($unknown)` are not reported.
- **D21** `fn` = `preg_match_all` with exactly 2 arguments, used as a
  logical operand → report (info: `preg_match` suffices). "Logical operand":
  after skipping enclosing parentheses, the parent is an
  `if`/`elseif`/`while`/`do-while` condition, a `!`, an operand of
  `&&`/`||`/`and`/`or`, or the condition of a full (non-`?:`) ternary.
- **D22** Plain string API. `C` has ≥ 2 arguments and `body` non-empty.
  `B'` as in D14.
  *Text form*: `B'` is, in full: an optional `^`, then either one or more
  characters of `[A-Za-z0-9_-]` or a backslash followed by one of `. * + ?`,
  then an optional `$`. Let `start`/`end` = whether `^`/`$` is present,
  `T` = the middle part with a backslash removed before any of
  `. + * ? -` ("unescape"), `ci` = `mods` contains `i`, `X1`/`X2` = verbatim
  source text of arguments 2/3.
  - **D22a** `fn` = `preg_match`, exactly 2 arguments, text form, `start`
    and `end`, not `ci` → replacement `"T" OP X1` with `OP` = `!==` if
    inverted else `===`.
  - **D22b** `fn` = `preg_match`, exactly 2 arguments, text form, `start`,
    not `end` (any `ci`) → `0 OP F(X1, "T")`, `OP` = `!==` if inverted else
    `===`, `F` = `stripos` if `ci` else `strpos`.
  - **D22c** `fn` = `preg_match`, exactly 2 arguments, text form, neither
    `start` nor `end` → `false OP F(X1, "T")`, `OP` = `===` if inverted else
    `!==`, `F` as in D22b.
  - **D22d** `fn` = `preg_replace`, exactly 3 arguments, text form, neither
    `start` nor `end` → `G("T", X1, X2)`, `G` = `str_ireplace` if `ci`
    else `str_replace`.
  D22a–D22d report the *context* (below) with a warning and the replacement,
  and end D22 for this literal. If none applies (e.g. `preg_match` with
  `start`+`end`+`ci`, or `end` without `start`), continue:
  - **D22e** Trim. `fn` = `preg_replace`, exactly 3 arguments, argument 2 is
    a string literal whose source text is exactly 2 characters (`''` or
    `""`), and `B'` is in full one of:
    (i) `^` c `q`; (ii) c `q` `$`; (iii) `^` c `q` `|` c `q` `$` with the
    same c on both sides — where c is either the two characters `\s` or any
    single character other than `.`, and each `q` is `+` or `*`
    (independently). If `mods` contains `m` or `u` → stop (no report).
    Otherwise: function `H` = `rtrim` if `body` does not start with `^`;
    else `ltrim` if `body` does not end with `$`; else `trim`. Replacement:
    `H(X2)` when c is `\s`, else `H(X2, 'c')` with c unescaped as above.
    Report `C` (whole call; warning); end D22.
  - **D22f** Explode. `fn` = `preg_split`, 2 or 3 arguments, `mods` empty,
    and either
    - `B'` is exactly one character other than `.`, or `[` + one character
      other than `.` + `]`; or
    - `B'` contains **no** "regex syntax", where regex syntax is: a
      character other than `\` immediately followed by one of
      `^ $ . * + ? \ [ ] ( ) { } ! | -`, or a `\` followed by one of
      `d D h H s S v V w W R b`.
    Replacement `explode("U", X1)` (2 args) or `explode("U", X1, X2)` (3
    args), `U` = `B'` unescaped (backslash removed before `. + * ? -`
    everywhere). Report `C` (whole call; warning).
  *Inverted* (only matters for D22a–c): if `C` is a logical operand (D21
  definition), inverted ⇔ `C`'s **direct** parent is `!`. Otherwise, if
  `C`'s direct parent is a binary expression with operator `<`, `==`,
  `!=`/`<>`, `===`, `!==` and the other operand (whichever side) is a
  number literal (optionally with unary minus), inverted ⇔ (op, number
  text) is one of (`<`, `1`), (`==`, `0`), (`===`, `0`), (`!=`, `1`),
  (`!==`, `1`). Otherwise not inverted.
  *Context* (D22a–d): if `C` is a logical operand whose direct parent is
  `!` → that `!` expression; else if `C`'s direct parent is a binary
  expression with operator `==`, `!=`, `===`, `!==`, `<`, `>`, `<=`, `>=`
  and the other operand is a number literal (optionally negated) → that
  binary expression; else `C` itself.
- **D23** Case conversion of the subject. `fn` = `preg_match`, exactly 2
  arguments, argument 2 is a plain function call resolving (as in D1) to the
  global `strtolower`, `strtoupper`, `mb_strtolower` or `mb_strtoupper` → report argument 2 (the whole inner call; warning). Two
  meanings: `mods` contains `i` → the conversion is redundant; otherwise →
  use the `i` modifier instead.

## Exceptions (no report)
- **E1** Method/static calls, other function names, differently-cased
  names; calls without arguments.
- **E2** First argument not resolvable to exactly one string literal;
  literal in another file; interpolated double-quoted strings; empty
  string.
- **E3** `preg_quote`: never D6 or D8 or D13 (but D7, D9–D12, D14–D19 still
  apply to a delimited-looking argument).
- **E4** Array-literal first argument: no D20–D23.
- **E5** D16 with a third argument, with `^` before `.*`, or with a `\0`
  back-reference.
- **E6** D22 with extra arguments (`preg_match` + `$matches`,
  `preg_replace` + limit, `preg_split` + flags), with any regex syntax in
  the body, D22a with `i`, D22e with a non-empty replacement or `m`/`u`,
  `.` as trim character, differing characters in the two trim halves,
  D22f with any modifier.
- **E7** D23 with a third argument or another wrapper function.

## Report
| Finding | Range | Severity |
|---|---|---|
| D6 | literal `L` (with quotes) | warning |
| D7, D8, D12, D13, D17, D18 | `L` | error |
| D9, D10, D11, D14, D15, D16, D19 | `L` | info |
| D20 | whole call `C` | warning |
| D21 | the function **name token** of `C` only (e.g. `preg_match_all`, not the arguments) | info |
| D22a–d | context (`!C`, `C op number`, or `C`) | warning |
| D22e, D22f | whole call `C` | warning |
| D23 | argument 2 (the case-conversion call) | warning |

Several findings on the same literal are independent and all reported
(multiset). Messages (our wording):
- D6 `Pattern has no valid delimiters.`
- D7 `The /e flag was removed from PCRE; use a callback replacement.`
- D8 `'{char}' is not a valid PCRE modifier.`
- D9a `The /D flag has no effect together with /m.`
- D9b `The /D flag is pointless: the pattern has no '$'.`
- D10 `The /s flag is pointless: the pattern has no '.'.`
- D11 `The /i flag is pointless: the pattern has no letters.`
- D12 `The /r flag needs /{i|u} to take effect.`
- D13a `Non-ASCII characters in the pattern need the /u flag.`
- D13b `Unicode escapes (\p, \P, \X) need the /u flag.`
- D14 `Write '{key}' as '{short}' ({same result without /u | matches more under /u}).` (custos: `same result` for the negated shorthand keys `[^\d]`, `[^\w]`, `[^\s]`, which name exactly the complement of their shorthand in every mode)
- D15 `Collapse '{run}' into '{class}' with a counted quantifier.`
- D16 `Drop the leading '.*'; it does not change whether the pattern matches.`
  / `Drop the trailing '.*'; …`
- D17 `Class [{S}] is redundant: {sub} is already covered by {super}.`
- D18 `Nested quantifier ({alt}){Q} risks catastrophic backtracking.`
- D19 `Tag content matched with '.' likely needs the /s flag.`
- D20 `Pass the delimiter to preg_quote() so it is escaped as well.`
- D21 `Use preg_match() when only testing for a match.`
- D22 `Replace with '{replacement}'.`
- D23 `The case conversion is redundant: the pattern is already
  case-insensitive.` / `Drop the case conversion and add the /i flag instead.`

## Fix
Only D22 has a fix. In every case the reported range is replaced by the
replacement text, verbatim, with no added parentheses:
- **F1** (D22a) `"T" === X1` / `"T" !== X1`.
- **F2** (D22b) `0 === strpos(X1, "T")`, `0 !== stripos(X1, "T")`, …
- **F3** (D22c) `false !== strpos(X1, "T")` / `false === stripos(X1, "T")`.
- **F4** (D22d) `str_replace("T", X1, X2)` / `str_ireplace("T", X1, X2)`.
- **F5** (D22e) `ltrim(X2, 'c')`, `rtrim(X2)`, `trim(X2, 'c')`, …
- **F6** (D22f) `explode("U", X1)` / `explode("U", X1, X2)`.
`T`, `c`, `U` are inserted raw between the quotes shown (double quotes for
`T`/`U`, single quotes for `c`); `X1`/`X2` are the original argument texts.
Separators are `, `; single spaces around `===`/`!==`.
The inserted builtin (`strpos`, `stripos`, `str_replace`, `str_ireplace`,
`ltrim`, `rtrim`, `trim`, `explode`) is written with a leading `\` when an
unqualified call at the reported position would not reach the global
function (a `use function` import under that name, or a same-named function declared in the current namespace).

## Options
None.

## PHP versions
- D8 allowed modifiers: `eimsuxADJSUX` below 8.2; `+n` from 8.2; `+n r` from
  8.4. Upstream fixtures without an explicit level run at the IDE test
  default (between 5.6 and 7.0), i.e. the base set; the `n`/`e`/unknown
  modifier fixture runs at 8.2 and the `r` fixture at 8.4.
- Everything else is level-independent (D7 and D12 included).

## Examples

Default level (e.g. 7.0):

```php
<?php
function modifiersDemo($line, $m) {
    preg_match('%k\d+%', $line, $m);
    preg_match('[k\d+]x', $line, $m);
    preg_match('', $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">'k\d+'</warning>, $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">')k\d('</warning>, $line, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">'/k\d/9'</warning>, $line, $m);
    preg_quote('k.d', '/');
    preg_quote('/k/Q', '/');
    preg_replace(<error descr="The /e flag was removed from PCRE; use a callback replacement.">'#k(\d)#e'</error>, '$1', $line);
    preg_replace([<error descr="The /e flag was removed from PCRE; use a callback replacement.">'#k(\d)#e'</error>], '$1', $line);
    preg_match(<error descr="'Q' is not a valid PCRE modifier.">'/k\d/Q'</error>, $line, $m);
    preg_match(<error descr="'n' is not a valid PCRE modifier.">'/(k)\d/n'</error>, $line, $m);

    preg_match('/^k\d$/D', $line, $m);
    preg_match(<weak_warning descr="The /D flag is pointless: the pattern has no '$'.">'/^k\d/D'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /D flag is pointless: the pattern has no '$'.">'/k\$/D'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /D flag has no effect together with /m.">'/k\d$/mD'</weak_warning>, $line, $m);

    preg_match('/k.\d/s', $line, $m);
    preg_match('/k\[.\]/s', $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k\d+/s'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k[.,]\d/s'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /s flag is pointless: the pattern has no '.'.">'/k\.\d/s'</weak_warning>, $line, $m);

    preg_match('/k\d/i', $line, $m);
    preg_match('/\\d/i', $line, $m);
    preg_match('/ид/iu', $line, $m);
    preg_match(<weak_warning descr="The /i flag is pointless: the pattern has no letters.">'/\d{3}-\d{2}/i'</weak_warning>, $line, $m);
    preg_match(<weak_warning descr="The /i flag is pointless: the pattern has no letters.">'/
        \s+ \W
    /ix'</weak_warning>, $line, $m);

    $rx = <warning descr="Pattern has no valid delimiters.">'k\w+'</warning>;
    preg_match($rx, $line, $m);
}
```

At PHP 8.2 `'/(k)\d/n'` is not reported. At PHP 8.4:

```php
<?php
function restrictDemo($w, $m) {
    preg_match('/kelvin/iur', $w, $m);
    preg_match(<error descr="The /r flag needs /u to take effect."><error descr="The /r flag needs /i to take effect.">'/kelvin/r'</error></error>, $w, $m);
    preg_match(<error descr="The /r flag needs /i to take effect.">'/kelvin/ur'</error>, $w, $m);
}
```

(At 8.3 `'/kelvin/iur'` gets one D8 error for `r`.)

Body checks (default level):

```php
<?php
function bodyDemo($s, $m) {
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u).">'/v[0-9]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[:digit:]' as '\d' (same result without /u).">'/[[:digit:]]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\d]' as '\D' (same result).">'/[^\d]+/u'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[A-Za-z0-9_]' as '\w' (same result without /u).">'/[a-zA-Z0-9_]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^A-Za-z0-9_]' as '\W' (same result without /u).">'/[^0-9A-Za-z_]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\s]' as '\S' (same result).">'/[^\s]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\w]' as '\W' (same result).">'/[^\w]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[:word:]' as '\w' (same result without /u).">'/[[:word:]]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^0-9]' as '\D' (same result without /u).">'/[^0-9]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u)."><weak_warning descr="Write '[^\s]' as '\S' (same result).">'/[0-9]-[^\s]/'</weak_warning></weak_warning>, $s, $m);
    preg_match('/[0-9,]/', $s, $m);

    preg_match(<weak_warning descr="Collapse '[ab][ab]' into '[ab]' with a counted quantifier.">'/[ab][ab]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab]{2}[ab]+' into '[ab]' with a counted quantifier.">'/x[ab]{2}[ab]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab][ab]' into '[ab]' with a counted quantifier.">'/[ab][ab][ab]/'</weak_warning>, $s, $m);
    preg_match('/[ab]x[ab]/', $s, $m);
    preg_match('/[ab][ba]/', $s, $m);

    preg_match(<error descr="Class [\w\d.] is redundant: \d is already covered by \w.">'/[\w\d.]/'</error>, $s, $m);
    preg_match(<error descr="Class [-\W\D] is redundant: \D is already covered by \W.">'/[-\W\D]/'</error>, $s, $m);
    preg_match(<error descr="Class [\w\d] is redundant: \d is already covered by \w."><error descr="Class [\W\D] is redundant: \D is already covered by \W.">'/[\w\d][\W\D]/'</error></error>, $s, $m);
    preg_match('/[\d\s]/', $s, $m);
    preg_match('/[\w\D]/', $s, $m);

    preg_match(<error descr="Nested quantifier (\w+)+ risks catastrophic backtracking.">'/(\w+)+$/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\s*)+ risks catastrophic backtracking.">'/(?:\s*)+x/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\d+)* risks catastrophic backtracking.">'/(a|\d+)*z/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\s+)+ risks catastrophic backtracking.">'/k(\s+|-(?=\d))+/'</error>, $s, $m);
    preg_match('/(\d+){2}/', $s, $m);
    preg_match('/(\d+)*+/', $s, $m);
    preg_match('/(?>(\s+|x)+)/', $s, $m);
    preg_match('/(\d{1,})+/', $s, $m);

    preg_match(<weak_warning descr="Tag content matched with '.' likely needs the /s flag.">'#<li>.*</li>#'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Tag content matched with '.' likely needs the /s flag.">'#<li>.+?</li>#'</weak_warning>, $s, $m);
    preg_match('#<li>.*</li>#s', $s, $m);
    preg_match('#<li>[^<]*</li>#', $s, $m);

    preg_match(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/café/'</error>, $s, $m);
    preg_match('/café/u', $s, $m);
    preg_match(<error descr="Unicode escapes (\p, \P, \X) need the /u flag.">'/\pL+/'</error>, $s, $m);
    preg_match(<error descr="Unicode escapes (\p, \P, \X) need the /u flag.">'/k\X/'</error>, $s, $m);
    preg_match('/\\p/', $s, $m);
    preg_quote('/ü/', '/');
}
```

Call-shape checks (default level):

```php
<?php
function callDemo($file, $needle, $name) {
    preg_match(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches.">'/.*\.log/'</weak_warning>, $file);
    preg_match(<weak_warning descr="Drop the trailing '.*'; it does not change whether the pattern matches.">'/tmp-.*/'</weak_warning>, $file);
    preg_match(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches."><weak_warning descr="Drop the trailing '.*'; it does not change whether the pattern matches.">'/.*-.*/'</weak_warning></weak_warning>, $file);
    $n = preg_match_all(<weak_warning descr="Drop the leading '.*'; it does not change whether the pattern matches.">'/.*;/'</weak_warning>, $file);
    preg_match('/.*\.log/', $file, $m);
    preg_match('/^.*\.log/', $file);
    preg_match('/.*=\0/', $file);
    preg_replace('/tmp-.*/', '', $file);

    <warning descr="Pass the delimiter to preg_quote() so it is escaped as well.">preg_quote('#a.b#')</warning>;
    preg_quote('#a.b#', '#');
    preg_quote($needle);
    preg_quote('a.b');

    if (<weak_warning descr="Use preg_match() when only testing for a match.">preg_match_all</weak_warning>('/\d+/', $file)) {}
    $n = preg_match_all('/\d+/', $file);
    if (preg_match_all('/\d+/', $file, $all)) {}

    preg_match('/^[a-z]+$/', <warning descr="Drop the case conversion and add the /i flag instead.">strtolower($name)</warning>);
    preg_match('/^[a-z]+$/i', <warning descr="The case conversion is redundant: the pattern is already case-insensitive.">mb_strtoupper($name)</warning>);
    preg_match('/^[a-z]+$/', strtolower($name), $m);
    preg_match('/^[a-z]+$/', trim($name));
}
```

Plain-API replacements (default level), before:

```php
<?php
function plainDemo($path, $name, $tpl, $raw, $list) {
    $r = [];
    $r[] = <warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path)</warning>;
    $r[] = <warning descr="Replace with 'false !== stripos($path, &quot;cache-dir&quot;)'.">preg_match('#cache-dir#i', $path)</warning>;
    $r[] = <warning descr="Replace with '0 === strpos($path, &quot;.&quot;)'.">preg_match('/^\./', $path)</warning>;
    $r[] = <warning descr="Replace with '0 === stripos($name, &quot;img_&quot;)'.">preg_match('/^img_/i', $name)</warning>;
    $r[] = <warning descr="Replace with '&quot;index&quot; === $name'.">preg_match('/^index$/', $name)</warning>;
    $r[] = <warning descr="Replace with '&quot;index&quot; !== $name'.">!preg_match('/^index$/', $name)</warning>;
    $r[] = <warning descr="Replace with '0 !== strpos($path, &quot;tmp&quot;)'.">preg_match('/^tmp/', $path) == 0</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($path, &quot;tmp&quot;)'.">0 === preg_match('/tmp/', $path)</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path) > 0</warning>;
    if (<warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path)</warning> && $name) {}
    $r[] = preg_match('/^index$/i', $name);
    $r[] = preg_match('/tmp$/', $path);
    $r[] = preg_match('/tmp/', $path, $hit);
    $r[] = preg_match('/tmp\d/', $path);

    $r[] = <warning descr="Replace with 'str_replace(&quot;__NAME__&quot;, $name, $tpl)'.">preg_replace('/__NAME__/', $name, $tpl)</warning>;
    $r[] = <warning descr="Replace with 'str_ireplace(&quot;draft&quot;, '', $tpl)'.">preg_replace('/draft/i', '', $tpl)</warning>;
    $r[] = <warning descr="Replace with 'str_replace(&quot;*&quot;, $name, $tpl)'.">preg_replace('/\*/', $name, $tpl)</warning>;
    $r[] = preg_replace('/draft/', '', $tpl, 2);
    $r[] = preg_replace(['/draft/'], '', $tpl);
    $r[] = preg_replace('/^draft/', 'x', $tpl);

    $r[] = <warning descr="Replace with 'ltrim($raw, '0')'.">preg_replace('/^0+/', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'rtrim($raw, '/')'.">preg_replace('#/+$#', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'trim($raw, '-')'.">preg_replace('/^-*|-+$/', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'trim($raw)'.">preg_replace('/^\s+|\s+$/', "", $raw)</warning>;
    $r[] = <warning descr="Replace with 'rtrim($raw)'.">preg_replace('/\s*$/', '', $raw)</warning>;
    $r[] = preg_replace('/^0+/u', '', $raw);
    $r[] = preg_replace('/^0+/', ' ', $raw);
    $r[] = preg_replace('/^.+/', '', $raw);
    $r[] = preg_replace('/^0+|1+$/', '', $raw);

    $r[] = <warning descr="Replace with 'explode(&quot;;&quot;, $list)'.">preg_split('/;/', $list)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;=>&quot;, $list, 3)'.">preg_split('/=>/', $list, 3)</warning>;
    $r[] = preg_split('/\s*,\s*/', $list);
    $r[] = preg_split('/;/u', $list);
    $r[] = preg_split('/;/', $list, -1, PREG_SPLIT_NO_EMPTY);
    return $r;
}
```

After fix:

```php
<?php
function plainDemo($path, $name, $tpl, $raw, $list) {
    $r = [];
    $r[] = false !== strpos($path, "tmp");
    $r[] = false !== stripos($path, "cache-dir");
    $r[] = 0 === strpos($path, ".");
    $r[] = 0 === stripos($name, "img_");
    $r[] = "index" === $name;
    $r[] = "index" !== $name;
    $r[] = 0 !== strpos($path, "tmp");
    $r[] = false === strpos($path, "tmp");
    $r[] = false !== strpos($path, "tmp");
    if (false !== strpos($path, "tmp") && $name) {}
    $r[] = preg_match('/^index$/i', $name);
    $r[] = preg_match('/tmp$/', $path);
    $r[] = preg_match('/tmp/', $path, $hit);
    $r[] = preg_match('/tmp\d/', $path);

    $r[] = str_replace("__NAME__", $name, $tpl);
    $r[] = str_ireplace("draft", '', $tpl);
    $r[] = str_replace("*", $name, $tpl);
    $r[] = preg_replace('/draft/', '', $tpl, 2);
    $r[] = preg_replace(['/draft/'], '', $tpl);
    $r[] = preg_replace('/^draft/', 'x', $tpl);

    $r[] = ltrim($raw, '0');
    $r[] = rtrim($raw, '/');
    $r[] = trim($raw, '-');
    $r[] = trim($raw);
    $r[] = rtrim($raw);
    $r[] = preg_replace('/^0+/u', '', $raw);
    $r[] = preg_replace('/^0+/', ' ', $raw);
    $r[] = preg_replace('/^.+/', '', $raw);
    $r[] = preg_replace('/^0+|1+$/', '', $raw);

    $r[] = explode(";", $list);
    $r[] = explode("=>", $list, 3);
    $r[] = preg_split('/\s*,\s*/', $list);
    $r[] = preg_split('/;/u', $list);
    $r[] = preg_split('/;/', $list, -1, PREG_SPLIT_NO_EMPTY);
    return $r;
}
```

## Divergences
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no candidate, no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/internals/decisions.md` ("Spec-level false positives").
- **D21 range.** The upstream fixture highlights only the function name of
  the `preg_match_all` call although the finding is attached to the whole
  call (the `preg_quote` finding, attached the same way, covers the whole
  call). Match the fixture: report the name token.
- **Arrays with several patterns (upstream bug).** Upstream processes one
  element (in arbitrary order) and then aborts with a concurrent-modification
  failure when the array has two or more resolvable literals. Recommendation:
  check every element (D2/D5). Only single-element arrays appear in fixtures.
- **Inversion/context gaps (upstream).** `1 > preg_match(…)`,
  `preg_match(…) <= 0` and `0 >= preg_match(…)` are treated as *not*
  inverted while the whole comparison is replaced, flipping the meaning;
  comparisons with other numbers (`== 2`) are replaced too. A parenthesised
  call compared to a number (`(preg_match(…)) === 0`) only has the call
  replaced, producing `(false !== strpos(…)) === 0`. Recommendation (implemented, see "D22 rewrites that keep the meaning" below): treat
  `n > C` like `C < n`, `<= 0`/`0 >=` as inverted, skip numbers other than
  0/1 and skip parenthesised operands. None of this is in the fixtures.
- **Precedence.** Replacements are inserted without parentheses
  (`preg_match('/^a$/', $s) + 1` → `"a" === $s + 1`). Recommendation (implemented, see "D22 rewrites that keep the meaning" below): wrap
  in parentheses when the context's parent binds tighter than `===`.
- **Broken trim / explode output.** A trim character `\` or `'`
  (`/^\++/`-like bodies such as `^\+`) yields invalid PHP (`ltrim($s, '\')`);
  explode copies the whole body raw, so `[;]` gives `explode("[;]", …)`, and
  `"`, `$` or `\` inside produce a broken or different string; a metacharacter
  at the very start of the body (`+x`, `|`) escapes the regex-syntax test.
  `preg_split` limit semantics (`0`/`-1`) differ from `explode`.
  Recommendation (implemented, see "D22 rewrites that keep the meaning" below): skip these cases (use the inner character for `[x]`).
- **Modifiers ignored by D22.** `m` (with `^`/`$`) and `x` change meaning
  but are not checked for D22a–d. Recommendation (implemented, see "D22 rewrites that keep the meaning" below): skip D22a/b when `mods`
  contains `m`. D22 also builds `T` from the normalised body, so a literal
  `a-zA-Z` becomes `A-Za-z` in the suggestion; use the original text.
- **Line terminators.** Upstream's D13 and D19 tests fail on any body
  containing a line break (a multi-line pattern with non-ASCII text is not
  reported), and its D22/D22f end-anchors also accept a single trailing
  newline. Recommendation: for D13 ignore line breaks (report anyway); keep
  D19 as specified. Not covered by fixtures.
- **End anchor and trailing newline (custos diverges from upstream).** In
  PCRE `$` also matches before a final newline unless the `D` modifier is
  set, so `^T$` accepts `"T\n"`: D22a (identity comparison) is offered only
  with `D`, and D22e trims of a character other than `\s` anchored with `$`
  (rtrim/trim) need `D` too (`c+$` stops before the newline; `\s+$`
  consumes it). Two upstream fixtures expect the rewrites; listed in
  `testdata/ea-divergences.json`.
- **`\s` is not trim()'s default set (custos diverges from upstream).**
  PCRE's `\s` is space, `\t`, `\n`, `\v`, `\f`, `\r`; trim()'s default
  list has `\0` instead of `\f`. D22e with `\s` writes the set
  explicitly: `H(X2, " \t\n\r\v\f")`.
- **D22e modifiers (custos diverges from upstream).** Only `D` and `S` are
  neutral; `i` is accepted only when c has no case variant (`/^a+/i` also
  strips `A`); `U` makes `+` lazy (one character removed), `x` ignores
  whitespace, `m`/`u` change anchors/encoding — all skipped.
- **Delimiters as PHP parses them (custos diverges from upstream).**
  Upstream accepts letters, digits and `\` as delimiters, and treats leading
  whitespace as the delimiter itself. So `'abca'` (rejected by PHP) is
  analysed as the body `bc` and can collect misleading pattern/modifier
  reports — or, through `preg_quote`, a "pass the delimiter" report — while
  `' /x/'` (accepted by PHP, which skips the space) is reported as having no
  valid delimiters. custos skips leading whitespace, rejects alphanumeric,
  backslash and NUL delimiters, and reports only "no valid delimiters" for
  them (D4/D6).
- **Per-call duplication (custos diverges from upstream).** Upstream
  repeats every literal finding once per call that reaches the literal
  through value discovery, so one `$re = '/x/e';` used by three calls shows
  the same error three times on the same range. custos reports each literal
  finding once (see the note after D5).
- **`preg_quote` with one argument** is only reported when its literal
  looks delimited. Kept on purpose: a 1-argument call is correct whenever
  the surrounding pattern uses a delimiter preg_quote already escapes
  (`'{' . preg_quote($s) . '}'`, common in Symfony), so reporting every
  1-argument call would add false positives.
- **Heredoc/nowdoc patterns:** upstream handling unverified;
  recommendation: treat a nowdoc / non-interpolated heredoc like a quoted
  literal whose contents are its body.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  the `preg_*` functions (and the case conversions of D23) by the name as
  written (case-sensitive, any namespace qualifier, no resolution), so a
  differently cased call such as `PREG_MATCH(...)` is missed while a
  namespaced or imported user function of the same name is reported (and
  rewritten) as if it were the builtin. custos matches case-insensitively and
  only calls that reach the global function.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `strpos(`, `str_replace(`, `trim(`, `explode(`, … in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F2–F6).
- **Literal non-ASCII text (custos diverges).** Upstream reports any
  non-ASCII character in a pattern without `/u` as an error, but a plain
  literal sequence (`'/\[entité\]/'`) matches the same UTF-8 bytes with or
  without `/u`. custos reports D13a only when the character sits in a
  character class, carries a quantifier, or is a letter under `/i` — the
  cases where byte-wise matching gives a different result.
- **D22 rewrites that keep the meaning (custos diverges).** Upstream's D22
  suggestions can change what the code does; custos only rewrites when the
  result is equivalent:
  - *Comparisons* of a `preg_match()` call with a number literal are
    evaluated for the two possible results 0 and 1 (with the number on
    either side, `n > C` read as `C < n`; `===`/`!==` with a float literal
    never hold): when the comparison holds for 1 only, it is replaced by
    the positive test; for 0 only, by the inverted one; when it holds for
    both or neither (`== 2`, `== -1`, `< 5`) nothing is reported. A call
    in parentheses compared with a number, or used as the operand of any
    other operator (arithmetic, `.`, `@`, casts, `instanceof`, a comparison
    with a non-literal), is not reported. A logical operand (`!C`,
    conditions, `&&`/`||`) keeps the upstream context. A value position
    (assignment, return, argument, `echo`) is not reported (custos
    diverges): `preg_match()` yields 1/0 there and the replacement
    true/false, which changes printed output and fails `: int` returns
    under `strict_types`. The fix wraps the replacement in
    parentheses when its context is the operand of a tighter operator
    (`'n' . !preg_match('/^a$/', $p)` → `'n' . ("a" !== $p)`); the message
    shows it without them.
  - *D22d* (`preg_replace` → `str_replace`) replaces the call only, never
    an enclosing comparison.
  - *Modifiers:* D22a–d are skipped when `mods` contains `A` (anchored at
    the start), and D22a/D22b (anchored forms) when it contains `m`.
  - *Trim (D22e):* not reported when the character is a regex
    metacharacter (`\ ^ $ . [ ] | ( ) ? * + {`) or a quote — `/^\+/` has no
    quantifier and `ltrim($s, '\')` does not parse.
  - *Explode (D22f):* the delimiter is the literal string the pattern
    matches: the character of a one-character class `[c]` (`[.]` included;
    not `\`, `^`, `]`), a single character that is not a metacharacter, or
    a body without regex syntax whose first character is not a
    metacharacter (`/|/` splits between every character), with
    `\. \+ \* \? \-` resolved; a body keeping another escape (`\/`, `\n`)
    is not reported, and `"` is escaped in the replacement. With a third
    argument the call is rewritten only when it is a positive integer
    literal (`preg_split()` treats `0`/`-1` as "no limit", `explode()` does
    not).
- **D9–D11 count the pattern PHP passes to PCRE (custos diverges).**
  Upstream counts `$`, `.` and letters in the literal's source text, so
  `"/^[a-z]*\$/D"` and `"/^[0-9]*\x24/D"` (an anchor `$` written as a PHP
  escape) were told the `D` flag is pointless — it is not (`"ab\n"`
  matches without it). custos runs these three checks on the string value
  (escapes resolved, both quote styles): `"\x2e"` is a `.`, and
  `'/\\d/i'` is the pattern `\d` (no letters, `/i` pointless) while
  `'/\\\\d/i'` is an escaped backslash followed by the letter `d`.
- **preg_quote() text (custos diverges from E3).** `preg_quote()`'s
  argument is literal text, not a pattern: only D20 applies to it (the
  delimiter split still decides whether it looks delimited); D7 and D9–D19
  are not run (`preg_quote('/test/invalidpath/testing', '/')` was told its
  "/e flag" was removed).
- **Mandatory inner groups (custos diverges from D18 step 2).** Step 2
  replaces an inner group by a placeholder atom instead of deleting it,
  unless the group is followed by `?`: `(\s+(?:unsigned|zerofill))*` is not
  `(\s+)*` (each repetition needs the keyword), while `(\s+(?:x)?)` may
  still collapse.
- **Decoded escapes for D13b (custos diverges).** The `\p`/`\P`/`\X`
  search runs on the decoded pattern, not the source text: `'/A\\\P/'`
  (single-quoted) is the pattern `A\\P`, an escaped backslash then `P`
  (phpspreadsheet's autoloader), and `'/\\p/'` is the pattern `\p`.
  Upstream reports the first and misses the second (listed divergence
  `missing-u-modifier.php`).
- **D22d needs a literal replacement (custos diverges).** preg_replace()
  interprets `\0`–`\99`, `$n`, `${n}` and `\\` in its replacement;
  str_replace() inserts it as is. D22d applies only when the replacement
  is a quoted literal without `\` and `$`, a number, or an `(int)`/
  `(float)` cast; other replacements (variables, escaped SQL values,
  translations, `'$0$0'`) are not reported by D22d (Dolibarr's
  `preg_replace('/__HANDLER__/i', "'" . $db->escape($h) . "'", $sql)`
  would have lost the backslash collapsing).
