---
id: HostnameSubstitution
group: Security
kind: syntax
needs: []
php: { min: "", max: "" }
---

# HostnameSubstitution

## Summary

`$_SERVER['HTTP_HOST']` and `$_SERVER['SERVER_NAME']` can be controlled by the
client (Host header). Using them to build e-mail addresses or storing them in
domain/host/email-like variables or properties without a whitelist check lets
an attacker inject their own domain.

## Detection

### Source

- **D1** An array access `$_SERVER[KEY]` where the base is exactly the
  variable `$_SERVER` and `KEY` is a string literal (either quote style)
  whose content is exactly `SERVER_NAME` or `HTTP_HOST` (case-sensitive).
  Call this node `S` and the key content `ATTR`.

### Context search

- **D2** Walk up the ancestors of `S`, one at a time, stopping at the file
  root or at the first function / method / closure / arrow function / class
  boundary (nothing found → no report). The **first** ancestor that is
  either a concatenation (`.`) or a plain assignment (`=` only; compound
  assignments such as `.=` are passed through like any other node) decides:
  - concatenation → **concatenation check** on that node (D3);
  - plain assignment → **assignment check** (D4–D6).
  Either way the walk ends there.

### Concatenation check on a concatenation node `C`

- **D3** `X` is the node being checked (the source `S`, or a variable
  occurrence in D5) and `C` the concatenation that holds it. Take the whole
  concatenation chain around `C`: climb from `C` through enclosing
  concatenations and parentheses, then flatten that chain into its operands
  in source order, looking through parentheses (`'a' . ('b' . $x)` has the
  operands `'a'`, `'b'`, `$x`). Let `O` be the operand that contains `X`.
  Report `O` (parentheses stripped, kind **E**) when the operand right
  before `O` is a string literal whose raw content ends with `@`, and `S`
  is not whitelisted (D7). So `'a' . 'user@' . X`, `'user@' . (X . '.com')`
  and `'user@' . trim(X)` are all reported, at `X`, `X` and `trim(X)`.

### Assignment check on an assignment `T = …`

- **D4 (property target)** `T` is a property fetch (`$o->name`,
  `Cls::$name`) with a non-empty static name that contains `domain`,
  `email` or `host` (case-insensitive substring) and `S` is not whitelisted
  → report `S` (kind **N**).
- **D5 (variable target inside a function)** `T` is a variable and `S` lies
  inside a function/method/closure `F`: walk all variable occurrences in
  `F` in source order (nested closures included). After passing the target
  occurrence `T` itself, every later occurrence with the same name whose
  parent — looking through parentheses — is a concatenation `C` gets the
  concatenation check (D3) on `C` with `X` = that occurrence (still using
  `S`/`ATTR` for the whitelist test and message).
  One report per such occurrence, per assignment.
- **D5a (overwrite stops the flow)** The walk of D5 stops at the first later
  occurrence that is the target of a plain assignment `$v = expr;` written
  as an expression statement, when
  - the statement list containing that statement encloses the source
    assignment (same block, or an outer block of the same function — so the
    overwrite runs on every path after the source), and
  - `expr` does not mention the variable `$v` (nor any variable-variable).
  After such an overwrite the variable no longer holds the host name, so
  later uses are not checked. A re-assignment *derived* from the variable
  (`$v = str_replace('www.', '', $v);`, `$v = strtolower($v);`) keeps the
  host-controlled value and does not stop the walk; neither does an
  overwrite inside a branch, loop or nested closure.
- **D6 (variable target at top level)** `T` is a variable and there is no
  enclosing function: report `S` (kind **N**) when the variable name
  contains `domain`, `email` or `host` (case-insensitive) and `S` is not
  whitelisted (never whitelisted at top level, see D7).

### Whitelist

- **D7** `S` is whitelisted when its nearest enclosing function/method/
  closure body contains (anywhere) a plain function call resolving to the
  global `in_array` (any case; a same-named namespaced function does not
  whitelist) whose first argument is an array access equivalent to `S` (same text,
  e.g. `in_array($_SERVER['HTTP_HOST'], $allowed)`). Top-level code is never
  whitelisted.

## Exceptions (no report)

- **E1** Other `$_SERVER` keys, non-literal keys, other superglobals.
- **E2** No concatenation/assignment ancestor before the function/class
  boundary (`echo $_SERVER['HTTP_HOST'];`, `return $_SERVER['SERVER_NAME'];`).
- **E3** Concatenations where the operand right before the host operand is
  not a literal ending with `@` (`'Host: ' . $_SERVER['HTTP_HOST']`,
  `$_SERVER['HTTP_HOST'] . '@relay'`, `$user . $host`).
- **E4** Assignment targets that are array elements (`$a['k'] = …`) or
  whose name does not match (D4/D6); inside functions, variables never used
  later directly inside a concatenation, or used only after being
  overwritten with an unrelated value (D5a).
- **E5** Whitelisted via `in_array` (D7).

## Report

- Range: kind E — the chain operand `O` that holds the host, parentheses
  stripped (for `'@' . strtolower($_SERVER['HTTP_HOST'])` the whole
  `strtolower(...)` call; for the variable flow D5, the later variable
  occurrence); kind N — the `$_SERVER[...]` access `S`.
- Severity: error.
- Messages:
  - E: `E-mail address built from client-controlled $_SERVER['{ATTR}']; validate it against a whitelist.`
  - N: `Client-controlled host name stored here; validate it against a whitelist.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
$sender  = 'noreply@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
$bounce  = 'x' . 'bounce@' . <error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">trim($_SERVER["SERVER_NAME"])</error> . '.local';
$siteHost = <error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['SERVER_NAME']</error>;
$label   = 'Served by ' . $_SERVER['SERVER_NAME'];
$current = $_SERVER['HTTP_HOST'];

function contact()
{
    $base = strtolower($_SERVER['HTTP_HOST']);
    $base = str_replace('www.', '', $base);
    return 'help@' . <error descr="E-mail address built from client-controlled $_SERVER['HTTP_HOST']; validate it against a whitelist.">$base</error>;
}

function fallback()
{
    $base = $_SERVER['HTTP_HOST'];
    $base = 'example.org';
    return 'help@' . $base;
}

class Mailer
{
    public function configure()
    {
        $this->mailDomain = <error descr="Client-controlled host name stored here; validate it against a whitelist.">$_SERVER['HTTP_HOST']</error>;
    }

    public function safe(array $known)
    {
        if (in_array($_SERVER['HTTP_HOST'], $known, true)) {
            $this->replyHost = $_SERVER['HTTP_HOST'];
        }
    }
}
```

## Divergences

- **Whitelist call matching (custos diverges from upstream).** Upstream
  accepts any call whose written last segment is exactly `in_array` (D7),
  so `In_Array(...)` does not whitelist while a namespace's own `in_array()`
  does. custos requires the call to resolve to the global `in_array`, in any
  case.
- Overwritten variables (custos diverges from upstream). Upstream follows
  every later occurrence of the variable regardless of re-assignments, so
  `$base = $_SERVER['HTTP_HOST']; $base = 'example.org'; 'help@' . $base`
  is reported although the e-mail address no longer contains the host
  name. custos stops at an unconditional overwrite with a value that does
  not depend on the variable (D5a). String transformations of the variable
  itself (`str_replace('www.', '', $base)`) are not sanitising — the client
  still controls the result — so they keep being followed, as upstream does.
- **Operand-accurate concatenation check (custos diverges):** upstream only
  inspects the right operand of the first concatenation found and the
  operand directly to its left, so a host that is the left or a middle
  operand of that node — typically after parentheses regroup the chain, as
  in `'admin@' . ($host . '.example')` — yields a report on another node or
  none at all. custos flattens the whole chain and checks the operand that
  actually holds the host against its predecessor (D3).
- **Fetches inside `isset()` / `empty()` (custos).** A `$_SERVER['SERVER_NAME']`
  that is only tested (`!empty($_SERVER['SERVER_NAME']) ? $_SERVER['SERVER_NAME'] : …`)
  is not a source: only the read flows into the address, so the finding is
  reported once instead of twice (Swiftmailer's message ids).
