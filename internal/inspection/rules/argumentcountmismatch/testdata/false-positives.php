<?php
namespace Shop;

function total($items, $tax) {}
function collect(...$items) {}

// Unpacked arguments, first-class callables and unknown named arguments.
$args = [[], 0.2];
total(...$args);
$fn = total(...);
total(items: [], rate: 0.2);
collect();

// Extra arguments to the project's own functions are legal.
function sum() { return array_sum(func_get_args()); }
sum(1, 2, 3);
total([], 0.2, 'extra');

// Unresolved callees.
missing_function();
$name = 'total';
$name();
function viaMixed($thing) { $thing->add(); }

// Declared twice: which declaration runs depends on the loading code.
if (PHP_VERSION_ID >= 80000) {
    function pick($a, $b) {}
} else {
    function pick($a) {}
}
pick(1);

// Overloaded builtins whose stub declares one form.
array_keys(['a' => 1]);
strtr('abc', ['a' => 'b']);
implode([]);
mt_rand();
\stream_filter_append(STDIN, 'dechunk', STREAM_FILTER_READ);

// An override accepting the call.
class Renderer { public function render($template, $data) {} }
class Plain extends Renderer { public function render($template, $data = []) {} }
class Loose extends Renderer { public function render(...$parts) {} }
function show(Renderer $r) { $r->render('home'); }

// Magic methods and `@method` tags.
/** @method void ship(string $address) */
class Parcel
{
    public function __call($name, $args) {}
    private function seal($tape) {}
}
function send(Parcel $p) { $p->ship(); $p->seal(); $p->unknown(); }

class Registry { public static function __callStatic($name, $args) {} protected static function lookup($key) {} }
Registry::lookup();

// Trait receivers run in the classes using the trait.
trait Loads
{
    abstract public function load($path);
    public function boot() { $this->load(); }
}

// Unknown ancestors may declare the method.
class Legacy extends \Vendor\Missing\Base {}
function old(Legacy $l) { $l->run(); }

// Redeclared classes.
if (PHP_VERSION_ID >= 80000) {
    class Twice { public function go($a) {} }
} else {
    class Twice { public function go() {} }
}
function twice(Twice $t) { $t->go(); }

// Abstract classes, interfaces, classes without a constructor.
abstract class Shape { public function __construct($sides) {} }
class Point {}
$p = new Point(1, 2);
$p = new Point;

// Union receivers whose classes disagree.
class A { public function run($x) {} }
class B { public function run() {} }
function union(A|B $o) { $o->run(); }

// Extra arguments to a builtin method overridden by the project.
class Clock extends \DateTimeImmutable { public function format($format, $zone = null): string { return ''; } }
function clock(\DateTimeImmutable $d) { $d->format('c', 'UTC'); }

// Identical declarations loaded conditionally.
if (PHP_VERSION_ID >= 80000) {
    function same($a) {}
} else {
    function same($a) {}
}
same();

// Abstract classes, interfaces and enums are not constructed.
$shape = new Shape();
$count = new \Countable();
enum Suit { case Hearts; }
$suit = new Suit();

// An interface that is not indexed may declare anything.
class Wired implements \Vendor\Missing\Contract { public function __construct($a) {} }
$wired = new Wired();

// A builtin constructor whose stub signature differs from PHP's.
$formatter = new \IntlDateFormatter('en');

// Static calls inside a trait.
trait Makes
{
    public static function make($size) {}
    public function boot() { self::make(); static::make(); }
}
/** @param Makes $maker */
function viaTrait($maker) { $maker->make(); Makes::make(); }

// A subtype declared twice may override the method either way.
class Sender { public function send($to) {} }
if (PHP_VERSION_ID >= 80000) {
    class SmsSender extends Sender {}
} else {
    class SmsSender extends Sender { public function send($to = null) {} }
}
function notify(Sender $s) { $s->send(); }
