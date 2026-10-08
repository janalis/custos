---
id: UnusedConstructorDependencies
group: Unused
kind: semantic
needs: [names, types, hierarchy, index]
php: { min: "", max: "" }
---

# UnusedConstructorDependencies

## Summary
A private property that is assigned in the constructor but never touched by
any other method of the class (or its traits) is dead state — typically a
leftover injected dependency. Point at the constructor assignments.

## Detection
Visit every class's own constructor.

- **D1** The containing type is a class (not an interface or trait), it
  declares its own constructor (`__construct` in the class body itself),
  and it declares at least one property and at least one method (the
  constructor counts).
- **D2** Candidate properties: the class's **own** declared properties that
  are `private`, not `static`, and not *annotated*. A property is annotated
  when its doc comment contains at least one tag whose name (including the
  `@`) is not entirely lower-case — e.g. `@Id`, `@ORM\Column`,
  `@Inject` — while lower-case tags such as `@var`, `@internal` do not count.
  Stop if there are no candidates.
- **D3** A *relevant reference* to candidate `p` inside some method body is
  any property access named `p` (instance `$x->p` on any object, nullsafe
  `$x?->p`, or static `X::$p`), at any depth including nested closures and
  arrow functions, that either resolves to the candidate property `p` of
  this class or cannot be resolved at all. Accesses resolving to some other
  class's property are ignored. Dynamic names (`$this->{$n}`) are ignored.
- **D4** Collect relevant references sitting directly in the constructor's
  own scope (not inside a closure or arrow function defined there). Stop if
  none.
- **D4a** References inside closures/arrow functions defined in the
  constructor are treated separately, because such a function can run after
  the constructor returns: a reference that is the target of a plain `=`
  assignment is ignored (neither reported nor a use); any other reference
  (a read, a call on it, `++`, a compound assignment…) counts as a use, like
  a D5 reference.
- **D5** Collect relevant references inside every other method declared in
  the class and in every method declared in the traits the class uses
  directly (`use T;`). (Methods of parent classes and of traits used by
  those traits are not scanned.)
- **D6** For each candidate `p` that has D4 references but no use (D4a or
  D5), report every D4 reference to `p` that is the **target** of a plain
  assignment (`$this->p = …` or `$this->p = &…`; compound
    assignments like `.=` and `??=` are not targets). Reads (`$this->p->run()`,
    `++$this->p`, uses on the right-hand side) are not reported.

## Exceptions (no report)
- **E1** Non-private or static properties; properties declared in parents or
  traits; annotated properties.
- **E2** Properties referenced in any other method of the class or of a
  directly used trait (any access — read or write, on `$this` or another
  instance of the class, including inside closures of those methods).
- **E3** Properties referenced in the constructor only from closures; a
  property read by a closure or arrow function defined in the constructor
  (`$this->cb = fn() => $this->logger->info('x');`); plain writes inside such
  closures are never reported.
- **E4** Classes without their own constructor, interfaces, traits.

## Report
- Range: the assigned property access expression (`$this->p`), from `$this`
  through the property name; not the whole assignment.
- Severity: info (fixture markup `weak_warning`).
- Message: `Private property is only used in the constructor; likely dead code.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
trait Describes
{
    public function describe() { return $this->label; }
}

class Mailer
{
    use Describes;

    private $transport;
    private $logger;
    private $label;
    private $retries;
    private $hook;
    private $handler;
    public $onError;
    /** @Inject */
    private $clock;
    /** @var int */
    private $limit;
    protected $debug;

    public function __construct($transport, $logger, $label, $clock, $limit, $handler)
    {
        $this->transport = $transport;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->logger</weak_warning> = $logger;
        $this->logger->info('ready');
        $this->label = $label;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->retries</weak_warning> = 3;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->retries</weak_warning> = $this->retries + 1;
        $register = function () { $this->hook = true; };
        $this->onError = fn($e) => $this->handler->handle($e);
        $this->handler = $handler;
        $this->clock = $clock;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->limit</weak_warning> = $limit;
        $this->debug = false;
    }

    public function send($msg)
    {
        return $this->transport->push($msg);
    }

    public static function inspect(Mailer $other)
    {
        return function () use ($other) { return $other->hook; };
    }
}
```

## Divergences
- **D4/D4a — custos diverges from upstream.** Upstream counts references
  inside closures defined in the constructor as constructor references: a
  property written directly in the constructor and read by a closure or
  arrow function created there (a callback that runs later) is reported as
  dead, and writes inside such closures are reported as well. custos treats
  a read inside those closures as a use, and never reports a write inside
  them. No upstream fixture covers it.
- **Attributes (custos diverges).** A property with a PHP 8 attribute
  (`#[ORM\Column]`, `#[ORM\Id]`) counts as annotated (D2/E1) like one with
  a non-lower-case doc tag: mapped properties are read by the ORM through
  reflection (Doctrine ORM test models: 26 reports).
