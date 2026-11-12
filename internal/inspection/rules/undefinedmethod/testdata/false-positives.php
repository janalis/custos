<?php
namespace Shop;

// Magic handlers anywhere in the hierarchy.
class Proxy
{
    public function __call($name, $args) {}
    public static function __callStatic($name, $args) {}
}
class Gateway extends Proxy {}

/** @method int total() */
class Invoice {}

/** @method void notify() */
interface Notifiable {}
class Customer implements Notifiable {}

/** @mixin Invoice */
class Facade {}
class Report extends Facade {}

trait Counts
{
    public function count(): int
    {
        // $this, self:: and static:: in a trait are the using class's.
        $this->reset();
        self::boot();
        static::boot();
        return 0;
    }
    public function tally(): int { return 0; }
}

class Basket
{
    use Counts { tally as total; }

    public function run(): void
    {
        // Methods a trait brings, under their own name or an alias.
        $this->count();
        $this->total();
        // $this in a closure or an arrow function may be another object.
        $fn = function () {
            $this->bind();
            self::boot();
        };
        $arrow = fn () => $this->bind() && static::boot();
    }
}

// Methods only a subclass or an implementation declares.
interface Shape {}
class Circle implements Shape
{
    public function radius(): float { return 1.0; }
}
class Animal {}
class Dog extends Animal
{
    public function bark(): void {}
}

final class Note {}
final class Memo
{
    public function send(): void {}
}

// Declared twice, or only when a condition holds.
class Twice {}
class Twice {}
if (!class_exists(Polyfill::class)) {
    class Polyfill {}
}

// A hierarchy the index does not fully know.
class Child extends \Vendor\Missing\Base {}

enum Suit: string
{
    case Hearts = 'h';
}

function calls(
    Gateway $gateway, Invoice $invoice, Customer $customer, Report $report,
    Shape $shape, Animal $animal, Note|Memo $either, Twice $twice,
    Polyfill $polyfill, Child $child, $untyped, object $object, mixed $mixed,
    \RecursiveIteratorIterator $files, Shape&Circle $both,
): void {
    $gateway->anything();
    Gateway::anything();
    $invoice->total();
    $customer->notify();
    $report->total();
    $shape->radius();
    $animal->bark();
    $either->send();
    $twice->shown();
    $polyfill->load();
    $child->load();
    $untyped->load();
    $untyped?->load();
    $object->load();
    $mixed->load();
    $files->getSubPathname();
    $both->radius();
    Suit::cases();
    Suit::from('h');
    // Dynamic method names.
    $name = 'load';
    $invoice->$name();
    Invoice::$name();
}

/** @param Note $note */
function documented($note): void
{
    // Only PHPDoc types the receiver.
    $note->send();
}

function guarded(Note $note, Invoice $invoice): void
{
    if (method_exists($note, 'send')) {
        $note->send();
    }
    method_exists($note, 'play') && $note->play();
    $result = is_callable([$note, 'stop']) ? $note->stop() : null;
    if (!method_exists($note, 'pause')) {
        return;
    }
    $note->pause();
    while (method_exists($note, 'skip')) {
        $note->skip();
    }
}

class Plain
{
    public function run(): void
    {
        // parent:: in a class without a parent resolves to nothing.
        parent::run();
    }
}

function outside(): void
{
    self::boot();
    $note = new Note();
    $note->{'dynamic'}();
}
