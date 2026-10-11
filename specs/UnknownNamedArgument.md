---
id: UnknownNamedArgument
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "8.0", max: "" }
---

# UnknownNamedArgument

## Summary

Detect a named argument absent from every resolved target parameter list where no target has a variadic parameter.

## Detection

- D1. Report a named argument absent from every resolved target parameter list where no target has a variadic parameter. Parameter names are compared case-sensitively.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.
- D4. The resolved targets of a call are every declaration PHP may bind it to:
  - a function call: the builtin of that name, otherwise every project declaration of the function (a function declared several times and loaded conditionally accepts the names of whichever declaration runs);
  - an instance method call: the method as each class of the receiver's type resolves it, plus the method as every indexed subtype of those classes resolves it (own declaration, trait import or inherited), because the runtime object may be any subtype and PHP binds names against the method that actually runs;
  - a `static::` call: the method of the enclosing class plus those of its indexed subtypes, as for an instance call;
  - a call through a named class, `self` or `parent`: only the method that class resolves.
- D5. Subtypes are not consulted when the nominal method is private or final, or its receiver class is final: no override can replace it.
- D6. A builtin target counts only when its stub parameter names are verified: the declaration belongs to an extension bundled with PHP whose stub names were compared with the engine's reflection, and it is not one of the listed builtins whose stub is known to differ (`session_set_cookie_params`, `dba_fetch`, `ldap_sasl_bind`, `IntlCalendar::set`).

## Exceptions (no report)

- E1. Unresolved targets, variadic collectors, method overload uncertainty and valid inherited parameters are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. No report when any target of D4 is unknown: a receiver class that is not indexed, a class of a union receiver lacking the method, a dynamic method or class name.
- E4. No report when a target is a builtin outside D6 (PECL and other unverified extensions such as Imagick, AMQP or SNMP, and the listed drifted builtins): their stub names document the API but are not the names the extension declares.
- E5. No report when any target, including an override in a subtype, has a variadic parameter, or is declared only by a `@method` tag (served by `__call` / `__callStatic`, which accept any name).
- E6. No report when the receiver class is a trait (`$this`, `self` or `static` inside a trait method): the method runs in the using classes, which may declare the callee with other names.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Use a declared parameter name.`

## Fix

No automatic fix: spelling similarity does not prove the intended parameter.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Available from PHP 8.0. Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
function pause($timeout) {} pause(<error descr="Use a declared parameter name.">timeot:3</error>);
```

A name declared by an override in a subtype is accepted; a name no target declares is reported:

```php
<?php
class Renderer { public function render($template) {} }
class TwigRenderer extends Renderer { public function render($view) {} }
function show(Renderer $r) {
    $r->render(view: 'home');
    $r->render(<error descr="Use a declared parameter name.">tpl: 'home'</error>);
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
