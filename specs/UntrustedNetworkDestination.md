---
id: UntrustedNetworkDestination
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedNetworkDestination

## Summary

Detect request-derived destination reaching curl, URL stream reads or equivalent network sinks without a proven outbound destination policy.

## Detection

- D1. Report request-derived destination reaching curl, URL stream reads or equivalent network sinks without a proven outbound destination policy. The sinks are:
  - `curl_exec` on a handle whose effective URL is request-derived. The URL is the `curl_init` argument, replaced in execution order by every dominating `curl_setopt($h, CURLOPT_URL, …)` and by the `CURLOPT_URL` entry of every dominating `curl_setopt_array` call whose option array is a literal. The last such assignment decides, so a later fixed URL clears an earlier request-derived one. An option array that cannot be read as a literal changes nothing.
  - `file_get_contents`, `fopen`, `readfile`, `file` (the filename) and `copy` (the source), when the request can choose a remote stream: either the request-derived value supplies the first bytes of the name, and with them the scheme, or the name starts with a literal `http://`, `https://`, `ftp://` or `ftps://` and the origin is not fixed (D4).
  - a call to a project function whose summary shows the argument reaching `curl_init` or a `CURLOPT_URL` assignment.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required. The option constant is resolved too: `\CURLOPT_URL` and an unqualified name falling back to the global constant both count; a namespaced constant with the same short name does not.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. The leading operand of a destination is found by looking through parentheses, the left side of concatenations, the first part of an interpolated string and variables with a single reaching value. The origin is fixed when that operand is literal text spelling a scheme, `://`, a non-empty authority and the `/`, `?` or `#` that ends it. The closing delimiter is required: without it a later operand could still extend or replace the host. A variable is followed to its reaching assignment only when its scope cannot rebind it behind that assignment; a reference alias, a by-reference closure capture, `foreach` or argument, a `catch` binding or `static` declaration of the same name, `extract()`, an assignment through a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry keeps the report.

Project method wrappers are followed only when a single visible declaration is fixed by a final receiver class, a final method, or a private method accessible at the call. Open overridable methods, ambiguous declarations, conditional class bodies and trait bodies do not establish a wrapper sink. Static calls must name a static method. These contracts compose through nested project wrappers and refresh when class or method dispatch metadata changes.

Numeric array values retain their network-destination provenance through wrappers: numerical content alone does not validate an outbound destination.

## Exceptions (no report)

- E1. Fixed destinations, policy validating schemes/resolved addresses/redirects, non-network local streams and unknown flows are excluded. In particular:
  - a destination with a fixed origin (D4), whatever request data follows in the path, query or fragment;
  - a destination that is a `match` expression (directly or through a variable with a single reaching value) whose every arm is a string literal, a concatenation starting with a fixed origin, a `throw`, or a nested `match` of the same shape. Arms are judged syntactically; an arm that is a variable is not proven. A `match` followed by a concatenated suffix is not covered;
  - a stream name whose leading operand is other literal text (a local path, a non-network wrapper), is not request-derived, or cannot contain `://`: an integer, float or boolean cast, or a call to `intval`, `floatval`, `basename`, `urlencode` or `rawurlencode`.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and any other subject prove nothing. The subject is a plain variable or an element of a request superglobal read under one constant key (`switch ($_GET['action'])`, not `$_FILES`); the element is the matched label only while nothing that may run code is evaluated: a call, an object creation, an assignment, an `unset`, or the use of a local variable, property, offset of a local or class constant ends it, as does a reference to the superglobal. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Validate the outbound network destination and redirects.`

## Fix

No automatic fix: outbound allowlists, DNS and redirect handling are deployment-specific.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$ch = curl_init($_GET['url']); <error descr="Validate the outbound network destination and redirects.">curl_exec($ch)</error>;

$h = curl_init();
curl_setopt_array($h, [\CURLOPT_URL => $_GET['feed']]);
<error descr="Validate the outbound network destination and redirects.">curl_exec($h)</error>;

$page = <error descr="Validate the outbound network destination and redirects.">file_get_contents('http://' . $_GET['mirror'] . '/index.json')</error>;

// Not reported: the origin is written in the source.
$fixed = curl_init('https://api.example.test/search?q=' . urlencode($_GET['q']));
curl_exec($fixed);
readfile('/srv/exports/' . $_GET['name']);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
