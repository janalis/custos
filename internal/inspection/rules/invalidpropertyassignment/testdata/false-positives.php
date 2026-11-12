<?php
declare(strict_types=1);

namespace Shop;

class Order {}
class RushOrder extends Order {}
interface Shippable {}

class Cart
{
    public int $count = 0;
    public float $ratio = 0.0;
    public int|string $code = 0;
    public Shippable $parcel;
    public RushOrder $rush;
    public $untyped;
    /** @var int */
    public $documented;
    public array $items = [];

    public function fill(Order $order, $unknown, mixed $anything, int|string $either)
    {
        // Accepted, or possibly accepted, values.
        $this->ratio = 1;
        $this->code = 'x';
        $this->count = $unknown;
        $this->count = $anything;
        $this->count = $either;
        // A non-final class that a subclass may make Shippable; a value
        // declared as a parent class.
        $this->parcel = $order;
        $this->rush = $order;
        // Untyped properties: only PHPDoc speaks, which is another rule's.
        $this->untyped = 'x';
        $this->documented = 'x';
        // Dynamic property names.
        $name = 'count';
        $this->$name = 'x';
        $this->{'count'} = 'x';
        // By-reference writes and aliases.
        $value = 'x';
        $this->count = &$value;
        $alias = &$this->count;
        $alias = 'x';
        // Not a property write.
        $this->items[] = 'x';
        $this->items['k'] = 'x';
        // `??=` stores its right-hand side only when the property is null.
        $this->count ??= 1;
    }
}

// Hooked and virtual properties run code instead of storing the value.
class Temperature
{
    public float $celsius = 0.0 {
        set(float|string $value) {
            $this->celsius = (float) $value;
        }
    }

    public string $label {
        get => 'x';
    }

    public function warm()
    {
        $this->celsius = 'hot';
        $this->label = 1;
    }
}

// `__set`, in the class or a subclass, may intercept the write.
class Magic
{
    public int $count = 0;

    public function __set($name, $value) {}

    public function fill()
    {
        $this->count = 'x';
    }
}

class Plain
{
    public int $count = 0;
}

class Lazy extends Plain
{
    public function __construct() { unset($this->count); }

    public function __set($name, $value) {}
}

function viaParent(Plain $p) { $p->count = 'x'; }

// Readonly properties: a second write fails as a re-initialization.
final class Invoice
{
    public readonly int $number;
    public readonly int $year;

    public function __construct(public readonly int $code)
    {
        $this->number = 1;
        $this->number = 'A-1';
        $this->code = 'x';
    }

    public function renumber()
    {
        $this->year = 'x';
    }
}

function renumber(Invoice $i) { $i->number = 'x'; }

// Properties the writing scope cannot see: PHP fails before the type check.
class Vault
{
    private int $secret = 0;
    protected int $shared = 0;
}

function breakIn(Vault $v) { $v->secret = 'x'; $v->shared = 'x'; }

class Sibling
{
    public function poke(Vault $v) { $v->shared = 'x'; }
}

// Code in a closure may be bound to another object or scope.
class Binder
{
    protected int $count = 0;

    public function later(Binder $other)
    {
        $f = function () { $this->count = 'x'; };
        $g = function () use ($other) { $other->count = 'x'; };
        $h = fn () => $other->count = 'x';
    }
}

// Receivers the code does not type, or types only by PHPDoc.
function untypedReceiver($cart) { $cart->count = 'x'; }

/** @param Cart $cart */
function documentedReceiver($cart) { $cart->count = 'x'; }

// Trait methods run in the classes using the trait.
trait Counts
{
    public int $total = 0;

    public function reset() { $this->total = 'x'; self::$cache = 'x'; }
}

// A class whose ancestors are not all known.
class Imported extends \Vendor\Missing
{
    public int $count = 0;

    public function fill() { $this->count = 'x'; }
}

// Undeclared properties, `@property` tags, a static write to an instance
// property.
/** @property int $magic */
class Loose
{
    public int $count = 0;

    public function fill()
    {
        $this->missing = 'x';
        $this->magic = 'x';
        Loose::$count = 'x';
    }
}

// PHP's own classes.
class Failure extends \Exception
{
    public function move() { $this->line = 'x'; }
}

// Static properties through a dynamic class.
function dynamicClass(string $cls) { $cls::$lines = 'x'; }

// Template-typed properties.
/** @template T */
final class Box
{
    /** @var T */
    public $value;

    public function put() { $this->value = 'x'; }
}

// No class scope for `parent`.
final class Orphan
{
    public static int $size = 0;

    public function grow() { parent::$size = 'x'; }
}

// A cyclic hierarchy, which PHP refuses to load.
class Loop extends Knot
{
    public int $count = 0;
}

class Knot extends Loop {}

function tie(Loop $l) { $l->count = 1; }

// A subtype whose hierarchy the index does not hold.
class Sized
{
    public int $size = 0;
}

class Odd extends Sized implements \Vendor\Missing\Contract {}

function resize(Sized $s) { $s->size = 'x'; }
