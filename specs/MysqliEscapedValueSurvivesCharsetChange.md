---
id: MysqliEscapedValueSurvivesCharsetChange
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# MysqliEscapedValueSurvivesCharsetChange

## Summary

Identify query SQL containing unchanged real_escape_string output after set_charset on same mysqli connection to a proven different charset. Require known charset before escaping and successful charset change; unknown changes exclude.

## Detection

- D1. Report query SQL containing unchanged real_escape_string output after set_charset on same mysqli connection to a proven different charset. Require known charset before escaping and successful charset change; unknown changes exclude.
- D1a. Only the last `set_charset` call written on that connection decides each side. Before escaping, it must be the last call ahead of the `real_escape_string` call, execute on every path (or be checked for success) and name a constant charset. Before the query, the last call written between the escape and the query must be checked for success on the path to the query and name a constant charset that differs, ignoring letter case, from the escaping charset. `set_charset` calls on other connections are ignored.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E1a. No report when the charset is switched and then switched back to the escaping charset before the query, nor when the last `set_charset` call before the query (or before the escape) is conditional, unchecked or names an unknown charset: an earlier proven change no longer describes the connection at the query.
- E2. Both `@custos-ignore MysqliEscapedValueSurvivesCharsetChange` and `@noinspection MysqliEscapedValueSurvivesCharsetChange` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Escape values after selecting the connection charset.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { $db->query("SELECT '$v'"); }
```

Valid case:

```php
<?php
$db = new mysqli(); $db->set_charset('gbk'); $v = $db->real_escape_string($input); $db->query("SELECT '$v'");
```

Valid case (charset restored before the query):

```php
<?php
function lookup(mysqli $db, $input) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('latin1')) {
        $db->query("SELECT 1");
        if ($db->set_charset('utf8mb4')) {
            $db->query("SELECT '$v'");
        }
    }
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
