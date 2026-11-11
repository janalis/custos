<?php
declare(strict_types=1);

namespace Shop;

class Order {}
class RushOrder extends Order {}
interface Shippable {}

function charge(int $cents) {}
function rush(RushOrder $order) {}
function ship(Shippable $item) {}
function run(callable $task) {}
function tally(int ...$amounts) {}
function grow(int &$count) {}

// Some values may be accepted: a union, a non-final class that a subclass
// may make Shippable, a value declared as a parent class.
function mixedValues(Order $order, int|string $amount, $unknown)
{
    charge($amount);
    charge($unknown);
    ship($order);
    rush($order);
}

// int widens to float; a function name may be callable.
function ratio(float $r) {}
ratio(1);
run('strlen');

// By-reference parameters are written, not only read.
$label = 'many';
grow($label);
$list = 'b,a';
sort($list);

// Unpacked, named and first-class callable arguments that bind to no
// known parameter.
tally(...['1', '2']);
charge(...['x']);
charge(price: 'x');
strlen(string: 42);
$fn = charge(...);

// The local may be rewritten unseen: by reference, or by a closure.
function aliased()
{
    $amount = 'ten';
    $alias = &$amount;
    $alias = 10;
    charge($amount);
}

// Unresolved callees.
$name = 'Shop\charge';
$name('x');
function viaUnknown($thing) { $thing->charge('x'); }
missing_function('x');

// Declared twice: which declaration runs depends on the loading code.
if (PHP_VERSION_ID >= 80000) {
    function pick(int $a) {}
} else {
    function pick(string $a) {}
}
pick('x');

// An override accepting the value.
class Register { public function put(int $amount) {} }
class Lenient extends Register { public function put(int|string $amount) {} }
function record(Register $r) { $r->put('x'); }

// `@method` tags, __call and __callStatic.
/** @method void stamp(int $n) */
class Parcel
{
    public function __call($name, $args) {}
    public static function __callStatic($name, $args) {}
    private function seal(int $tape) {}
}
function send(Parcel $p) { $p->stamp('x'); $p->unknown('x'); $p->seal('x'); }
Parcel::lookup('x');

// A class whose ancestors are not all known.
class Imported extends \Vendor\Missing { public function put(int $n) {} }
(new Imported())->put('x');

// Trait methods run in the classes using the trait.
trait Counts
{
    public function add(int $n) {}
    public function again() { $this->add('x'); }
}

// Builtins whose stub declares other parameter types than PHP: an
// overloaded form, an array of options.
strtr('abc', ['a' => 'b']);
setcookie('name', 'value', ['expires' => 0]);

// A [object, method] pair may be callable; $this in a closure may be bound
// to another object.
class Hooks
{
    public function boot()
    {
        run([$this, 'boot']);
        $later = function () { rush($this); };
        $soon = fn () => rush($this);
    }
}

// An override that renames the parameter: the named argument binds to none.
class Ledger { public function post(int $amount) {} }
class Journal extends Ledger { public function post(int $value) {} }
function book(Ledger $l) { $l->post(amount: 'x'); }
