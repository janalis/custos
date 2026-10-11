---
id: SignalHandlerHasNoDispatchMechanism
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.1", max: "" }
---

# SignalHandlerHasNoDispatchMechanism

## Summary

Dispatch signals when asynchronous handling is disabled.

## Detection

- D1. Require preceding explicit pcntl_async_signals(false), then pcntl_signal with a literal closure handler, before a while loop in the same lexical scope. Report the loop when neither its condition nor its body can dispatch: the only calls found there are the builtin sleep and usleep.
- D1a. Anything else that runs code in the condition or body removes the proof, not only pcntl_signal_dispatch: any other function call (builtin, application or dynamic), a method or nullsafe method call, a static call, an object construction, an include/require, eval, and yield or yield from (the generator's consumer may dispatch).
- D1b. An active ticks directive removes the proof, because pending signals are then delivered between statements. It is active when the loop sits inside a block-form declare naming ticks, or when a statement-form declare naming ticks appears earlier in the same file, at any nesting depth. The directive name is matched case-insensitively; the latest applicable directive wins; a literal zero switches ticks off; a block-form declare that ends before the loop has no effect on it. Other directives such as strict_types never count.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. Loops governed by ticks (D1b) and loops whose condition or body runs unknown code (D1a) are not reported.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Dispatch signals when asynchronous handling is disabled.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 7.1 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
pcntl_async_signals(false); pcntl_signal(SIGTERM,function(){echo "stop";}); <warning descr="Dispatch signals when asynchronous handling is disabled.">while (true) { usleep(10000); }</warning>
```

Valid case:

```php
<?php
pcntl_async_signals(false); pcntl_signal(SIGTERM,function(){echo "stop";}); while (true) { pcntl_signal_dispatch(); usleep(10000); }
```

A loop that delegates to other code is left alone:

```php
<?php
function run($worker) {
    pcntl_async_signals(false);
    pcntl_signal(SIGTERM, function () { echo "stop"; });
    while ($worker->hasWork()) { usleep(10000); }
}
```

So is a file that enables ticks:

```php
<?php
declare(ticks=1);
pcntl_async_signals(false); pcntl_signal(SIGTERM, function () { echo "stop"; }); while (true) { usleep(10000); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
