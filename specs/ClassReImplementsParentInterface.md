---
id: ClassReImplementsParentInterface
group: Architecture
kind: semantic
needs: [names, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# ClassReImplementsParentInterface

## Summary

Listing an interface in `implements` when the parent class already implements
it (directly or through its own ancestors) is redundant noise. Remove the
entry.

## Detection

- **D1** A class declaration (named or anonymous) with a non-empty
  `implements` list. Interfaces (`interface X extends …`) are not concerned.
- **D2** Resolve each `implements` entry to an interface declaration (through
  namespace and `use` aliases, e.g. `use \Countable as Sized;` then
  `implements Sized`). Unresolvable entries and declarations unavailable at
  the configured PHP version are ignored. A project polyfill remains a
  valid interface at that version.
- **D3** Resolve the direct parent class `P` (`extends`). No parent, or an
  unresolvable one → nothing.
- **D4** Build the inherited set of `P` (P itself excluded): every interface
  implemented by `P` or by any of its ancestor classes, closed transitively
  over interface inheritance (`interface B extends A` → `A` too), plus traits
  used along the chain (traits never match an `implements` entry, so they
  can be ignored).
- **D5** Report every entry whose resolved interface is in the inherited set
  (each entry node at most once).

## Exceptions (no report)

- **E1** Classes without a parent.
- **E2** An entry the parent does not cover: e.g. parent implements `A`, the
  class implements `B extends A` → `B` is not reported.
- **E3** Redundancy *within* the class's own list (`implements A, B` where
  `B extends A`, no parent involved) is not reported.
- **E4** The parent only covers the entry through a built-in class whose
  link to that interface does not exist at the target PHP version. The inherited set uses
  the index's version-specific interface lists, so a built-in interface
  link is followed only when PHP provides it at the target version.
  `class E extends \Exception implements \Stringable` is therefore silent
  below 8.0 (exceptions became `Stringable` in 8.0; with a polyfilled
  interface the entry is what makes the class `Stringable`) and reported
  from 8.0 on. Links between project classes are never gated.

## Report

- Range: the entry's name reference exactly as written (e.g. `Sized`,
  `\Countable`, `Contracts\Cache`).
- Severity: warning.
- Message: `'{interface FQN}' is already implemented by '{parent FQN}';
  remove it here.` (FQNs with leading `\`.)

## Fix

Each report carries a fix acting on its own entry; fixes are applied one at a
time (the conformance runner re-analyses until fixpoint), so each fix sees the
list as left by the previous ones.

- **F1** The entry is the **only** entry of the `implements` list: delete
  everything from the end of the `extends` clause to the end of the entry,
  `implements` keyword included. One space is left in its place unless
  whitespace already follows the entry, so at least one whitespace character
  remains between the parent name and `{`:
  `extends Base implements Sized {}` becomes `extends Base {}`.
- **F2** The list has several entries and this is the **first** one: delete
  the entry and the comma that follows it (skipping at most one whitespace
  run between entry and comma); whitespace after the comma stays.
  `implements A, B` → `implements  B`.
- **F3** The list has several entries and this is **not** the first one:
  delete the entry and the comma preceding it (skipping at most one
  whitespace run between comma and entry); the whitespace before the comma
  stays. `implements A, B` → `implements A`.
- **F4** No fix is offered when the text the fix would delete contains a
  comment (`extends Base /* note */ implements A`,
  `implements A /* why */, B`, `implements B, /* doc */ A`): the entry is
  still reported and the code is left as written. Comments outside the
  deleted text (after the entry in F1, after the comma in F2) do not
  withhold the fix and are preserved.

When every entry of a list is redundant, the successive fixes end with F1 and
the whole `implements` clause disappears.

- **F5** When the entry is an unqualified name brought in by a `use`
  import that names only it (`use Lib\Contract;` or `use Lib\Contract as
  C;`, not a group or multi-name import), in the same namespace, and the
  imported name appears nowhere else in the file as a whole word (case
  folded; code, comments and strings all count), the fix also deletes that
  import: the whole line when the import fills it (and one of the blank
  lines around it when it sat between two), otherwise just the statement.

## Options

None.

## PHP versions

The rule runs at every version. The target version only decides which
built-in interface links count (E4).

## Examples

```php
<?php
namespace Store {
    use Countable as Sized;

    interface Readable {}
    interface Seekable extends Readable {}
    abstract class Stream implements Readable {}
    abstract class FileStream implements Seekable, \Countable {}

    class Socket extends Stream
        implements
            <warning descr="'\Store\Readable' is already implemented by '\Store\Stream'; remove it here.">Readable</warning> {}

    class Disk extends FileStream implements <warning descr="'\Store\Readable' is already implemented by '\Store\FileStream'; remove it here.">Readable</warning>, <warning descr="'\Countable' is already implemented by '\Store\FileStream'; remove it here.">Sized</warning> {}

    class Tape extends FileStream implements \JsonSerializable, <warning descr="'\Store\Seekable' is already implemented by '\Store\FileStream'; remove it here.">Seekable</warning> {
        public function jsonSerialize(): mixed { return null; }
        public function count(): int { return 0; }
    }

    class Pipe extends Stream implements Seekable {}
    class Loose implements Readable, Seekable {}
}
```

```php
<?php
namespace Store {
    use Countable as Sized;

    interface Readable {}
    interface Seekable extends Readable {}
    abstract class Stream implements Readable {}
    abstract class FileStream implements Seekable, \Countable {}

    class Socket extends Stream
        {}

    class Disk extends FileStream {}

    class Tape extends FileStream implements \JsonSerializable {
        public function jsonSerialize(): mixed { return null; }
        public function count(): int { return 0; }
    }

    class Pipe extends Stream implements Seekable {}
    class Loose implements Readable, Seekable {}
}
```

## Divergences

- **F5 — custos diverges from upstream.** Upstream leaves the import of
  the removed interface in place, unused.
- **Version-aware built-in hierarchy (E4).** A deliberate custos guard, not
  taken from upstream: a built-in link to an interface PHP does not declare
  at the configured target version is ignored, because removing the entry
  would change `instanceof` results on that runtime.
- **Comments withhold the fix (F4).** A deliberate custos guard, not taken
  from upstream: the report is kept and no fix is offered when the deleted
  text would take a comment with it.
