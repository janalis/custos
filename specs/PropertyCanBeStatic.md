---
id: PropertyCanBeStatic
group: Architecture
kind: semantic
needs: [names, hierarchy, index]
php: { min: "", max: "" }
---

# PropertyCanBeStatic

## Summary
A non-public instance property initialised with a sizeable array literal
(several nested arrays or strings) is copied into every object. When such data
is really per-class configuration it is cheaper as a static property — or, on
PHP 5.6+, as a class constant.

## Detection
- **D1** A property declaration in a class body (each property of a
  multi-property declaration `private $a = …, $b = …;` is checked on its own).
  Promoted constructor parameters are not property declarations here.
- **D2** The property is not static and not public: it is `private` or
  `protected` (a `var` property is public and skipped). Class constants are
  never considered.
- **D3** Its default value is an array literal (`[...]` or `array(...)`).
- **D4** Parent check: if the class has an `extends` clause whose parent
  resolves, look up a property with the same name in the parent class
  *including everything the parent inherits* (its ancestors' properties of any
  visibility, private included, and properties of traits used anywhere in that
  chain). If found → no report. An unresolvable parent, or no parent, passes.
- **D5** Walk the array literal's top-level elements in order; for a
  `key => value` element take the value, for a plain element take the
  element's expression. Count the elements whose (unparenthesised) value is
  an **array literal** or a **string literal** (single/double quoted,
  interpolated, heredoc, nowdoc). Spread elements and all other values (numbers,
  constants, concatenations, `null`, …) are not counted but do not reset the
  count. Report once the count reaches **3**.

## Exceptions (no report)
- **E1** Public, `var` and static properties; constants.
- **E2** Defaults that are not array literals.
- **E3** Fewer than 3 array/string values among the top-level elements
  (nested levels are not counted).
- **E4** The property re-declares a property already present in the parent
  chain (D4).
- **E5** Suppression comments (a doc comment carrying `@noinspection` with the legacy ID
  on the declaration) are handled by the generic suppression mechanism.

## Report
- Range: the property name including `$` (e.g. `$labels`), not the default
  value.
- Severity: info.
- Message: PHP < 5.6: `Large array default on an instance property; consider
  a static property.` PHP ≥ 5.6: `Large array default on an instance property;
  consider a static property or a class constant.`

## Fix
None.

## Options
None.

## PHP versions
- No gating of the detection itself.
- Only the message depends on the configured level: the "class constant"
  suggestion appears from PHP 5.6 (array constants). The upstream positive
  fixture runs at PHP 5.5; the false-positive fixture runs at the test default
  (between 5.6 and 7.0) — messages are not compared.

## Examples
At PHP 8.x:

```php
<?php
class Base {
    protected $routes = [];
}

class Router extends Base {
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$verbs</weak_warning> = ['GET', 'POST', 'PUT', 'DELETE'];
    protected <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$matrix</weak_warning> = array('a' => [1], 'b' => [2], 7, 'c' => "x");
    private array <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$mixed</weak_warning> = [1, 'one', 2, <<<TXT
two
TXT, 3, ['three']];

    protected $routes = ['a', 'b', 'c'];
    private $pair = ['x', 'y'];
    private $nested = [['a', 'b', 'c', 'd']];
    public $open = ['p', 'q', 'r'];
    var $legacy = ['p', 'q', 'r'];
    private static $shared = ['p', 'q', 'r'];
    private $numbers = [1, 2, 3, 4];
    const NAMES = ['p', 'q', 'r'];
}
```

## Divergences
None known.
- **Per-instance state (custos diverges).** A property the class writes
  through `$this` (`$this->p = …`, `$this->p['k'] = …`, `$this->p[] = …`,
  compound assignments, `++`/`--`, `unset($this->p['k'])`) holds state
  of each instance; a static property would share it between instances
  (Roundcube's `rcube_db::$options`, set per connection). Such properties
  are not reported.
