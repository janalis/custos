---
id: ClassConstantUsageCorrectness
group: Probable bugs
kind: semantic
needs: [names, index, stubs]
php: { min: "", max: "" }
---

# ClassConstantUsageCorrectness

## Summary

`Foo::class` is resolved at compile time from the name *as written*, so a
class name typed with the wrong letter case (or imported with the wrong case)
produces a string that differs from the real class name. That string then
breaks case-sensitive comparisons, array keys, container ids and autoloaders
on case-sensitive file systems.

## Detection

Visit every class constant access `X::NAME`.

- **D1** `NAME` is the `class` keyword, compared case-insensitively as PHP
  does (`Foo::CLASS` is the same as `Foo::class`; custos diverges).
- **D2** `X` is a class *name* (not an expression such as `$obj::class`).
  Let `T` be the text of `X` as written (e.g. `Foo`, `Sub\Foo`, `\Foo`).
- **D3** `T` is not `self`, `static` or `parent` (compared
  case-insensitively, as PHP does; custos diverges, see Divergences).
- **D4** `X` resolves (normal PHP name resolution: imports, current
  namespace; class lookup case-insensitive) to a class-like declaration `C`
  (class, interface, trait, enum; project or stubs). Unresolved → stop.
- **D5** Build a list of acceptable spellings `L` (rules below). If `L` is
  **empty** → no report. Otherwise report when `T` equals (case-sensitive)
  **none** of the entries of `L`.

Let `NS` be the nearest enclosing `namespace` declaration of the access
(braced or unbraced; files without any namespace statement have none), and
`USES` the `use` import declarations that are *direct* statements of `NS`'s
body — or, when there is no `NS`, the top-level `use` imports of the file.
Each import has a target name (as written, without leading `\`), an FQN and
optionally an explicit alias (`as Alias`).

### L for a fully-qualified `T` (starts with `\`)

- **L1** Look up classes by FQN `T` (case-insensitive) in the index; if one is
  found, add its declared FQN with leading `\` (e.g. `\stdClass`). So `\Foo`
  must match the declared spelling exactly.

### L for a qualified `T` (contains `\`, no leading `\`)

Nothing is added when there is no `NS`.

- **L2** If `NS` is a named namespace (not the global `namespace { }`) and
  either lower(FQN of `C`) starts with lower(FQN of `NS`), or lower(FQN of `C`)
  ends with `\` + lower(`T`): add the last `len(T)` characters of `C`'s
  declared FQN. (E.g. in `namespace Shop`, `Sub\cart` resolving to
  `\Shop\Sub\Cart` yields `Sub\Cart`.)
- **L3** For each import in `USES` with an explicit alias `Al` where lower(`T`)
  starts with lower(`Al`): add `C`'s declared FQN in which every
  occurrence of the import's target text (as written) is replaced by `Al`
  (plain, case-sensitive substring replacement), then with the leading `\`
  removed. (E.g. `use Shop\Deep\Parts as P;` and
  `P\bolt` resolving to `\Shop\Deep\Parts\Bolt` yields `P\Bolt`.)

### L for an unqualified `T` (no `\` at all)

For each import in `USES`:

- **L4** If the import's FQN equals `C`'s FQN case-insensitively:
  - with an explicit alias → add the alias as written;
  - without alias → resolve the import's target to a class `R`; let
    *precise* = `R`'s declared FQN ends with the import target text as
    written (case-sensitive). If not precise, **or** `R`'s declared short name
    differs (case-sensitive) from `T`, add `R`'s declared FQN **with** its
    leading `\` (which can never equal an unqualified `T`, so this forces a
    report unless another entry matches). Otherwise (precise import, short
    name spelled as declared) add `T` itself: the access is correct, even
    when another import of the same class under an alias also contributes
    entries.
- **L5** Otherwise, if the import has an explicit alias equal to `T`
  case-insensitively → add that alias as written.

Consequences worth testing:

- An unqualified name that is not imported (class in the same namespace, or
  global class in a non-namespaced file without imports) gives an empty `L`
  → never reported, whatever its case.
- Correct import + correct case → `L` empty → no report.
- Import with wrong case (`use App\Models\usermodel;`) → every unqualified use
  of that class is reported, even if the access itself is spelled right.
- Correct import used with wrong case (`usermodel::class` after
  `use App\Models\UserModel;`) → reported.
- Alias used with a different case than declared → reported.

## Exceptions (no report)

- **E1** `self::class`, `static::class`, `parent::class`, `$var::class`.
- **E2** Unresolvable class names.
- **E3** Qualified names in files without a namespace statement.
- **E4** Empty acceptable-spelling list (see consequences above).

## Report

- Range: the class name part `X` only (e.g. `Sub\cart` in `Sub\cart::class`,
  `\stdclass` including the leading backslash), not `::class`.
- Severity: error.
- Message: `Letter case of the class name differs from its declaration; ::class will return the wrong string.`

## Fix

None.

## Options

None.

## PHP versions

None (`::class` itself exists since 5.5; no gating).

## Examples

```php
<?php

namespace Catalog {
    class Product {}
    class Variant {}
}

namespace Storefront {
    use Catalog\product;
    use Catalog\Variant as Option;

    return [
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Product</error>::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">OPTION</error>::class,
        Option::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">Widgets\banner</error>::class,
        Widgets\Banner::class,
    ];
}

namespace Storefront\Widgets {
    use Catalog\Product;

    class Banner {}

    return [
        Product::class,
        <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">product</error>::class,
        banner::class,
    ];
}

namespace Checkout {
    use Storefront\Widgets as W;
    use Storefront\Widgets\Banner as Promo;

    class Cart
    {
        public function names()
        {
            return [
                W\Banner::class,
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">W\BANNER</error>::class,
                <error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">\arrayobject</error>::class,
                \ArrayObject::class,
                Promo::class,
                self::class,
                static::class,
            ];
        }
    }
}
```

Non-namespaced file with imports of a global class:

```php
<?php
use ArrayObject;
use ArrayObject as Bag;

Bag::class;
ArrayObject::class;
<error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">arrayObject</error>::class;
```

(Not reported: the plain import is spelled as declared, so `ArrayObject`
is acceptable next to the alias `Bag`.)

## Divergences

- **Keyword case (custos diverges).** `SELF::class` / `Static::class`
  (keywords in another case) are not excluded by D3 upstream and are then
  resolved like class names; custos excludes `self`/`static`/`parent`
  case-insensitively. Upstream also requires the `class` keyword in lower
  case, so `product::CLASS` (which PHP evaluates exactly like
  `product::class`, returning the wrongly cased name) is missed; custos
  matches the keyword case-insensitively (D1).
- L3 performs a plain textual replacement of the import target inside the
  class FQN; if the import is written with a different letter case than the
  declaration, the replacement does nothing and the entry is the full FQN
  without leading `\` (forcing a report). This is correct, not a false
  positive: `::class` builds its string from the import as written, so
  `use Shop\deep\Parts as P; P\Bolt::class` yields `Shop\deep\Parts\Bolt`,
  not the declared `Shop\Deep\Parts\Bolt` — the same outcome as an
  unqualified access through a wrongly cased plain import (L4). Kept.
- L2 adds a suffix of the FQN even when that suffix does not align with a
  name segment boundary; harmless (it simply never matches).
- **Plain and aliased import of one class (custos diverges):** upstream
  adds nothing to `L` for a correct plain import, so when the same class is
  also imported under an alias, only the alias is acceptable and a
  correctly spelled plain access (`use ArrayObject; use ArrayObject as Bag;
  ArrayObject::class`) is reported. That access yields the right string;
  custos adds the correct plain spelling to `L` (L4) and stays silent.
- Group imports (`use A\{B, C as D};`): upstream handling of the target text
  is unverified; recommendation: treat each item as an import whose target is
  the full name (`A\B`).
- **Imports unrelated to the written name (custos diverges).** Upstream's
  L4 adds an import's alias whenever the import targets `C`, even when the
  access does not go through it: in `namespace Shop\Model; use
  Shop\Model\Bag as ModelBag;`, a bare `Bag::class` (resolved through the
  namespace, spelled as declared) is reported because `ModelBag` differs
  from `Bag`. custos considers, for unqualified `T`, only imports whose
  alias (or last segment when unaliased) equals `T` case-insensitively —
  the ones PHP actually uses to resolve `T`.
- **Case of an imported name at the access (custos diverges, L4/L5).**
  PHP resolves an unqualified class name against the imports
  case-insensitively and `::class` then yields the import's target as
  written in the `use` statement: after `use Magento\Framework\Filesystem;`,
  `FileSystem::class` is `'Magento\Framework\Filesystem'`, and an alias
  (`use A\Invoice as Bill;`, `BILL::class`) likewise. Only a wrong-case
  import (`use Catalog\product;` for `Catalog\Product`) gives the wrong
  string, so only that is reported (listed EA divergence;
  ~12 Magento/Yii findings were false).
- **`class_alias()` names (custos).** The index resolves an alias created
  by `class_alias()` to the original class; when the written name and the
  resolved class differ by more than letter case (`\App\Old\Resp::class`
  for an alias of `\App\Http\Resp`), there is no case mismatch to report
  (Grav's `GPM\Response` compatibility alias).
