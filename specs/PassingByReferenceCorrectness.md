---
id: PassingByReferenceCorrectness
group: Probable bugs
kind: semantic
needs: [names, index, hierarchy, types, stubs]
php: { min: "", max: "" }
---

# PassingByReferenceCorrectness

## Summary
A parameter declared by reference (`&$p`) needs a variable-like argument.
Passing the result of a call that does not return by reference, or a `new`
expression, makes PHP emit a notice ("only variables should be passed by
reference") and the callee's writes are lost.

## Detection
- **D1** Candidates: every plain function call `name(...)` and every method
  call (`$o->m(...)`, `A::m(...)`, `parent::m(...)`, …) with a non-empty
  static name.
- **D2** Function calls to `current` or `key` that resolve to the global
  built-ins are skipped.
- **D3** Pre-filter: the call has at least one argument, and not every
  argument is a plain variable. When the configured PHP level is **below
  7.0**, `new …` arguments also count as acceptable for this pre-filter (a
  call whose arguments are all variables or `new` expressions is skipped).
- **D4** Resolve the callee to its declaration (user code or stubs; for
  methods via the object's inferred type/class hierarchy). Unresolved → no
  report.
- **D5** For each position `i` from 0 to `min(#parameters, #arguments) - 1`
  (positional pairing; a variadic parameter only pairs with its own
  position), when parameter `i` is declared by reference:
  - **D5a** argument `i` is a call (function call, method call, static
    method call), not itself written with a leading call-time `&`, whose
    callee resolves to a declaration that is **not** declared to return by
    reference (`function &name(...)`) → report the argument. Unresolvable
    inner callee → no report.
  - **D5b** argument `i` is a `new …` expression → report the argument
    (at every PHP level, provided D3 let the call through).
  Other argument kinds (variables, property fetches, array elements,
  literals, static properties) are never reported by this rule.

## Exceptions (no report)
- **E1** All arguments plain variables (or, below PHP 7.0, variables/`new`).
- **E2** Arguments returned by reference: inner callee declared
  `function &f()` / `public function &m()`.
- **E3** Unresolvable outer or inner callees.
- **E4** Global `current()`/`key()`.
- **E5** Arguments beyond the declared parameter list; by-value parameters.

## Report
- Range: the offending argument expression, whole (e.g. `$repo->load()`,
  `Cfg::get()`, `array_merge(...$parts)`, `new Bag()`), without a trailing
  comma.
- Severity: warning.
- Message: `Pass a variable here: this parameter is taken by reference.`

## Fix
None.

## Options
None.

## PHP versions
- D3's treatment of `new` depends on the configured level (< 7.0 vs ≥ 7.0).
  The EA fixture runs at PHP 7.0, where `new` arguments are reported.

## Examples

```php
<?php
class Store
{
    public $items = [];
    public function &all() { return $this->items; }
    public function fill(&$target, $extra = null) { $target = []; }
    public function plain() { return []; }
    public static function make() { return []; }
}

function push_one(array &$list) { $list[] = 1; }

$s = new Store();
$s->fill($s->items, $s->plain());
$s->fill($buf);
$s->fill($s->all());
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">$s->plain()</warning>);
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">Store::make()</warning>, 1);
$s->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">new ArrayObject()</warning>, 1);
push_one(<warning descr="Pass a variable here: this parameter is taken by reference.">array_values(...[$buf])</warning>);
push_one($s->items['k']);
push_one(unknown_fn());
count($s->plain());
key($s->plain());
```

## Divergences
- Upstream caches, by short name, global functions known to have no
  by-reference parameter and skips later calls with that name, even if a
  namespaced function of the same short name exists. Recommendation: key any
  such cache by resolved FQN (no behavioural difference on fixtures).
- Named arguments are paired positionally upstream (probably wrong for
  `f(b: g())`). Recommendation: pair named arguments with the parameter of
  that name.
- **Prefer-ref parameters (custos diverges).** `array_multisort()` declares
  its parameters by reference in the stubs, but PHP passes them
  "prefer-ref": a temporary such as `array_values($a)` is accepted without
  any notice (its sort order still drives the multisort). Upstream reports
  `array_multisort(array_values($a), SORT_ASC, $a)`; custos skips calls that
  resolve to the global `array_multisort()`. `extract()` is prefer-ref too
  (by reference only for `EXTR_REFS`): `extract(array_merge($defaults,
  $vars))` is accepted silently (view renderers in Joomla, CakePHP, Yii),
  so calls resolving to the global `extract()` are skipped as well.
