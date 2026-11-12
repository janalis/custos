<?php
namespace Shop;

// Magic handlers anywhere in the hierarchy.
class Getter
{
    public function __get($name) {}
}
class Lazy extends Getter {}
class Probe
{
    public function __isset($name) {}
}

/** @property int $total */
class Invoice {}

/** @property-read string $email */
interface Contact {}
class Customer implements Contact {}

/** @mixin Invoice */
class Facade {}
class Report extends Facade {}

// Objects that accept any property.
class Row extends \stdClass {}
#[\AllowDynamicProperties]
class Bag {}
class Sack extends Bag {}

trait Counts
{
    public function count(): int
    {
        // $this and self:: in a trait are the using class's.
        return $this->size + self::$base;
    }
}

class Basket
{
    use Counts, Keeps;

    public function run(): void
    {
        // A property a trait brings.
        echo $this->kept;
        // $this in a closure or an arrow function may be another object.
        $fn = function () {
            return $this->size + self::$base;
        };
        $arrow = fn () => $this->size;
    }
}

trait Keeps
{
    public int $kept = 0;
}

class Shelf
{
    use Keeps;
}

// Properties only a subclass or an implementation declares.
interface Shape {}
class Circle implements Shape
{
    public float $radius = 1.0;
}
class Animal {}
class Dog extends Animal
{
    public string $breed = '';
}

final class Note {}
final class Memo
{
    public string $text = '';
}

// Declared twice, or only when a condition holds.
class Twice {}
class Twice {}
if (!class_exists(Polyfill::class)) {
    class Polyfill {}
}

// A hierarchy the index does not fully know.
class Child extends \Vendor\Missing\Base {}

function reads(
    Lazy $lazy, Probe $probe, Invoice $invoice, Customer $customer, Report $report,
    Row $row, \stdClass $plain, Bag $bag, Sack $sack, Shelf $shelf,
    Shape $shape, Animal $animal, Note|Memo $either, Twice $twice,
    Polyfill $polyfill, Child $child, $untyped, object $object, mixed $mixed,
    \SimpleXMLElement $xml, \ArrayObject $entries, Shape&Circle $both,
): array {
    return [
        $lazy->anything, $probe->anything, $invoice->total, $customer->email,
        $report->total, $row->id, $plain->id, $bag->id, $sack->id, $shelf->kept,
        $shape->radius, $animal->breed, $either->text, $twice->missing,
        $polyfill->missing, $child->missing, $untyped->missing,
        $untyped?->missing, $object->missing, $mixed->missing,
        $xml->child, $entries->key, $both->radius,
    ];
}

function dynamic(Note $note): void
{
    $name = 'missing';
    echo $note->$name, $note->{'missing'}, Note::${'missing'};
    $class = Note::class;
    echo $class::$missing;
}

/** @param Note $note */
function documented($note): string
{
    // Only PHPDoc types the receiver.
    return $note->text;
}

function contexts(Note $a, Note $b, array $list): void
{
    // Writes, which may create the property, and checks, which never warn.
    $a->created = 1;
    $a->appended[] = 1;
    $a->counter++;
    $a->joined .= 'x';
    $a->defaulted ??= 'x';
    $alias = &$a->aliased;
    unset($b->removed);
    foreach ($list as $a->looped) {}
    foreach ($list as $a->key => [$a->unpacked]) {}
    [$a->first, 'k' => [$a->second]] = $list;
    list($a->third) = $list;
    $refs = [&$a->refd];
    sort($a->sorted);
    unknown_function($a->passed);
    echo isset($b->asked), empty($b->asked), $b->asked ?? 'none';
    echo isset($b->nested->deep), $b->nested['k'] ?? null;
    echo isset(Note::$asked), Note::$asked ?? null;
    // A property this file creates may be read anywhere.
    echo $b->created, $b->appended[0];
}

function guarded(Note $note): mixed
{
    if (property_exists($note, 'shown')) {
        echo $note->shown;
    }
    if (isset($note->checked)) {
        echo $note->checked;
    } elseif (property_exists($note, 'other')) {
        echo $note->other;
    }
    property_exists($note, 'tested') && print($note->tested);
    if (!property_exists($note, 'early')) {
        return null;
    }
    return $note->early;
}

function outside(): void
{
    echo self::$base, parent::$base;
}
