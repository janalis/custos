---
id: DisallowWritingIntoStaticProperties
group: Code style
kind: semantic
needs: [names, index, hierarchy]
php: { min: "", max: "" }
---

# DisallowWritingIntoStaticProperties

## Summary

Static properties are global mutable state. This opt-in rule flags assignments
to static properties: by default only writes performed outside the class that
declares the property; optionally every write.

## Detection

D1. Node: an assignment expression whose target (left-hand side) is directly a
    static property access `X::$name` (the `::` form), with a statically known
    property name. Plain `=`, by-reference `= &`, and compound assignments
    (`+=`, `.=`, `??=`, …) all count.

D2. Option `ALLOW_WRITE_FROM_SOURCE_CLASS = false`: every assignment matching D1
    is reported, whatever the class part is (`self`, `static`, `parent`, a class
    name, an unknown class, an expression such as `$obj::$p`) and wherever it
    occurs.

D3. Option `ALLOW_WRITE_FROM_SOURCE_CLASS = true` (default): the class part
    must be a class-name reference (a name, including the keywords `static`
    and `parent`), and its written spelling must not be `self` in any letter
    case (`self`, `SELF`, `Self` — PHP keywords are case-insensitive). Then:
    - D3a. If the nearest enclosing function-like scope is a **method**: resolve
      the property (resolve the class name through `use` imports/aliases and the
      current namespace, `static`/`parent` relative to the method's class, then
      look the property up in that class and its ancestors). Report when the
      property resolves and the class that **declares** the resolved property
      is not the class/trait containing the method. A child class overriding
      (re-declaring) the property counts as declaring it.
    - D3b. If the nearest enclosing scope is **not** a method — top-level code,
      a plain function, a closure or an arrow function (upstream: even one
      nested inside a method; custos: see D3c) — report unconditionally,
      whether or not the class or property exists.
    - D3c. (custos, see Divergences) A closure or arrow function nested in a
      method is checked as that method (D3a): it has the method's class
      scope (`self`/`static` resolve to it).

## Exceptions (no report)

E1. Reads of static properties (`echo X::$p;`).
E2. Writes that are not assignments: `X::$p++`, `--X::$p`, `unset(X::$p)`,
    writes into an element `X::$p[] = 1` / `X::$p['k'] = 1` (target is an
    array access, not the property itself), passing by reference.
E3. Instance property writes `$this->p = 1`.
E4. Dynamic property name `X::${$n} = 1` / `X::$$n = 1`.
E5. With the default option:
    - target class spelled `self`, in any letter case (anywhere);
    - class part that is not a name (`$obj::$p = 1`);
    - in a method, the property cannot be resolved (unknown class or property);
    - in a method, the property is declared by the method's own class, reached
      through any spelling (`static::`, the class's own name, an alias of it).

## Report

- Range: the whole assignment expression, from the start of the target to the
  end of the assigned value (the trailing `;` excluded).
- Severity: info (fixtures tag it `weak_warning`).
- Message (default mode): `Modify this static property only from the class that declares it.`
- Message (all-writes mode): `Avoid modifying static properties.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| ALLOW_WRITE_FROM_SOURCE_CLASS | bool | true | true: only writes from outside the declaring class are reported (D3). false: every static property assignment is reported (D2). |

## PHP versions

None.

## Examples

Default option (`ALLOW_WRITE_FROM_SOURCE_CLASS = true`):

```php
<?php

namespace Shop;

use Shop\Cache as Store;

class Cache
{
    public static $hits = 0;
    public static $ttl;

    public function touch()
    {
        self::$hits += 1;
        static::$ttl = 30;
        Store::$ttl = 60;

        array_map(function ($v) {
            <weak_warning descr="Modify this static property only from the class that declares it.">Cache::$hits = $v</weak_warning>;
        }, []);
    }
}

class DiskCache extends Cache
{
    public static $ttl = 5;

    public function reset()
    {
        DiskCache::$ttl = 0;
        <weak_warning descr="Modify this static property only from the class that declares it.">parent::$hits = 0</weak_warning>;
        <weak_warning descr="Modify this static property only from the class that declares it.">DiskCache::$hits = 0</weak_warning>;
        Unknown::$whatever = 1;
        Cache::$hits++;
    }
}

function warmUp()
{
    <weak_warning descr="Modify this static property only from the class that declares it.">Cache::$ttl = 10</weak_warning>;
}

<weak_warning descr="Modify this static property only from the class that declares it.">Nowhere::$thing .= 'x'</weak_warning>;
echo Cache::$hits;
```

All-writes mode (`ALLOW_WRITE_FROM_SOURCE_CLASS = false`):

```php
<?php

class Counter
{
    public static $value = 0;

    public function bump()
    {
        <weak_warning descr="Avoid modifying static properties.">self::$value = 1</weak_warning>;
        <weak_warning descr="Avoid modifying static properties.">static::$value *= 2</weak_warning>;
        return self::$value;
    }
}

<weak_warning descr="Avoid modifying static properties.">Counter::$value = 7</weak_warning>;
```

## Divergences

- Properties brought in by a trait used by the method's class are resolved
  upstream to the trait, so a write like `static::$p = 1` from the using class
  would be reported. Recommendation: treat properties imported from traits the
  caller class uses as declared by the caller (no report). No fixture covers it.
- **`self` in another letter case (custos diverges):** upstream compares
  the class part with `self` case-sensitively, so `SELF::$p = 1` falls
  through to resolution and is reported inside closures and functions,
  while `self::$p = 1` is not. Both mean the same class; custos excludes
  `self` case-insensitively (D3).
- **Methods of anonymous classes (custos diverges from its earlier
  behaviour):** an anonymous class has no name, so D3a could not compare it
  with the declaring class and such writes were never reported. custos now
  applies D3a to them as to any other class: `Base::$p = 1` or
  `parent::$p = 1` inside `new class extends Base { … }` is reported, a
  property of a trait the anonymous class uses counts as its own, and
  `static::$p` (unresolvable without a name) is not reported.
- **custos diverges — closures inside methods (D3c).** Upstream reports
  every write from a closure or arrow function, even one defined in a
  method of the declaring class (`static::$factory = null` in a callback of
  Laravel's `Str`), while it accepts `self::$p = …` in the same closure.
  Closures have the class scope of the method that defines them, so custos
  checks them like the method. One upstream case expects the old reports
  (listed in `testdata/ea-divergences.json`).
