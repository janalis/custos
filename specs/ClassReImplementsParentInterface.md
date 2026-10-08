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
  `implements Sized`). Unresolvable entries are ignored.
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

- **F1** The entry is the **only** entry of the `implements` list: delete the
  entry's text and the `implements` keyword. Whitespace around them is left
  in place, so `extends Base\n    implements\n        Sized {}` becomes
  `extends Base\n    \n         {}` — whitespace-collapsed:
  `extends Base {}`. Implementations may equally delete from the end of the
  `extends` clause to the end of the entry and leave one space; the result
  must keep at least one whitespace character between the parent name and
  `{`.
- **F2** The list has several entries and this is the **first** one: delete
  the entry and the comma that follows it (skipping at most one whitespace
  run between entry and comma); whitespace after the comma stays.
  `implements A, B` → `implements  B`.
- **F3** The list has several entries and this is **not** the first one:
  delete the entry and the comma preceding it (skipping at most one
  whitespace run between comma and entry); the whitespace before the comma
  stays. `implements A, B` → `implements A`.
When every entry of a list is redundant, the successive fixes end with F1 and
the whole `implements` clause disappears.

## Options

None.

## PHP versions

No gating.

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

None known.
