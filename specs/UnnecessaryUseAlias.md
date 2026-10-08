---
id: UnnecessaryUseAlias
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnnecessaryUseAlias

## Summary

An import alias identical to the last segment of the imported name
(`use App\Mail\Sender as Sender;`) changes nothing. Drop the `as` clause.

## Detection

Visit every namespace import clause (`use ...;` at file/namespace level),
including `use function`, `use const` and each item of a group import
`use Prefix\{A as A, B}`. Trait imports inside class bodies are ignored.

- **D1** The clause has an explicit, non-empty alias (`as Name`).
- **D2** The fully qualified imported name (with a leading `\`, after joining
  a group prefix with the item) ends with `\` + alias, compared
  **case-sensitively** for class and constant imports and
  **case-insensitively** for `use function` imports (PHP function names
  are case-insensitive, so `use function Util\Slugify as slugify;` changes
  nothing). A single-segment import such as `use \Redis as Redis;`
  or `use Redis as Redis;` qualifies (FQN `\Redis`).

## Exceptions (no report)

- **E1** Alias differs from the last segment, including differences in case
  only for class and constant imports (`use App\Sender as sender;`,
  `use const App\LIMIT as Limit;`).
- **E2** No alias.
- **E3** Trait `use` statements (and their `insteadof`/`as` adaptations).

## Report

- Range: the alias identifier only (the name after `as`).
- Severity: info.
- Message: `Alias {alias} repeats the imported name; remove it.`

## Fix

- **F1** Delete the ` as Alias` part: from the end of the imported name to the
  end of the alias (the whitespace before `as`, the `as` keyword, and the
  alias). The `;`, `,` or `}` that follows is kept.
  `use \Redis as Redis;` → `use \Redis;`;
  `use App\{Sender as Sender, Queue};` → `use App\{Sender, Queue};`.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
namespace Shop;

use Psr\Log\LoggerInterface as <weak_warning descr="Alias LoggerInterface repeats the imported name; remove it.">LoggerInterface</weak_warning>;
use \Redis as <weak_warning descr="Alias Redis repeats the imported name; remove it.">Redis</weak_warning>;
use function Util\slugify as <weak_warning descr="Alias slugify repeats the imported name; remove it.">slugify</weak_warning>;
use Psr\Cache\{CacheItemInterface as <weak_warning descr="Alias CacheItemInterface repeats the imported name; remove it.">CacheItemInterface</weak_warning>, InvalidArgumentException};
use Psr\Clock\ClockInterface as Clock;
use Psr\Link\LinkInterface as linkinterface;

class Cart {
    use \Shop\Traits\Totals { total as total; }
}
```

```php
<?php
namespace Shop;

use Psr\Log\LoggerInterface;
use \Redis;
use function Util\slugify;
use Psr\Cache\{CacheItemInterface, InvalidArgumentException};
use Psr\Clock\ClockInterface as Clock;
use Psr\Link\LinkInterface as linkinterface;

class Cart {
    use \Shop\Traits\Totals { total as total; }
}
```

## Divergences

- Group imports and `use function`/`use const` are not covered by EA fixtures;
  they are included because upstream treats every import clause the same way.
- **Function imports (custos diverges).** Upstream compares the alias with
  the last segment case-sensitively for every import kind. Function names
  are case-insensitive in PHP, so `use function Util\Slugify as slugify;`
  is just as redundant as an exact repeat; custos compares function imports
  case-insensitively. Class imports keep the case-sensitive comparison (a
  case-only alias can be a deliberate spelling choice), and constant names
  are case-sensitive, so they keep it too.
