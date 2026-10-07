---
id: NonSecureParseStrUsage
group: Security
kind: syntax
needs: []
php: { min: "", max: "" }
---

# NonSecureParseStrUsage

## Summary
Called with a single argument, `parse_str()` / `mb_parse_str()` create
variables in the current scope from user-supplied query strings, which can
overwrite existing variables. Always pass the result array as second
argument.

## Detection
- **D1** A plain function call (not a method/static call) that resolves to
  the global function `parse_str` or `mb_parse_str`. Names compare
  case-insensitively (`Parse_Str`); a qualified non-global name
  (`Ns\parse_str`), a non-global `use function` import or a same-named
  function declared in the current namespace does not count.
- **D2** The call has exactly **one** argument.

## Exceptions (no report)
- **E1** Two (or more) arguments, or zero arguments.
- **E2** Calls resolving to a non-global function; method calls
  (`$q->parse_str($s)`).

## Report
- Range: the function **name identifier only** (`parse_str` /
  `mb_parse_str`), excluding any leading `\`/namespace qualifier, the
  whitespace before `(` and the argument list. (Upstream fixtures write a
  space between the name and `(`; the highlight stops at the name.)
- Severity: error.
- Message: `Pass a result array as second argument instead of creating variables.`

## Fix
None.

## PHP versions
No gating (the one-argument form is removed in PHP 8.0 but still reported
at any level).

## Options
None.

## Examples

```php
<?php
function readQuery($raw)
{
    <error descr="Pass a result array as second argument instead of creating variables.">parse_str</error>($raw);
    \<error descr="Pass a result array as second argument instead of creating variables.">mb_parse_str</error> ($raw);

    parse_str($raw, $fields);
    mb_parse_str($raw, $more);
    return [$fields, $more];
}
```

## Divergences
- **Function-name matching (custos diverges from upstream).** Upstream
  matches the written last segment case-sensitively without resolution, so
  `Parse_Str($s)` is missed while `Ns\parse_str($s)` or a namespace's own
  `parse_str()` is reported. custos resolves the call to the global function,
  in any case (D1).
