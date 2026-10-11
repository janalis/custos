---
id: XmlParserNeverReceivesFinalChunk
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XmlParserNeverReceivesFinalChunk

## Summary

Identify xml_parser_free on a locally created parser after at least one xml_parse call, with no xml_parse true final argument and no xml_parse_into_struct completion. Unknown parser aliases or paths exclude.

## Detection

- D1. Report xml_parser_free on a locally created parser after at least one xml_parse call, with no xml_parse true final argument and no xml_parse_into_struct completion. Unknown parser aliases or paths exclude.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.
- D4. Handing the parser to other code before the free excludes the report, because the callee may send the final chunk itself. This covers the parser appearing anywhere in the arguments of a method call (including nullsafe), a static call, a constructor call or a function that is not an available builtin, also when the call sits in an arrow function of the same scope. Only the scope that contains the free is searched; other functions, methods and closures have their own variables. A builtin call nested in such arguments (for example reading the current line number for a log call) only reads the parser and does not exclude.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore XmlParserNeverReceivesFinalChunk` and `@noinspection XmlParserNeverReceivesFinalChunk` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Finish parsing before freeing the XML parser.`

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
$p = xml_parser_create(); xml_parse($p, '<root>'); xml_parser_free($p);
```

Valid case:

```php
<?php
$p = xml_parser_create(); xml_parse($p, '<root/>', true); xml_parser_free($p);
```

Valid case, the parser is handed to a method that may finish the parse:

```php
<?php
class Feed {
    function run() { $p = xml_parser_create(); xml_parse($p, '<root>'); $this->finish($p); xml_parser_free($p); }
    function finish($p) { xml_parse($p, '</root>', true); }
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
