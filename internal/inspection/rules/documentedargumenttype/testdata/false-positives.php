<?php
declare(strict_types=1);

namespace Shop;

class Order {}

/** @param negative-int $offset */
function rewind(int $offset) {}

/** @param positive-int $quantity */
function restock(int $quantity) {}

// Ranges and lengths count only against exact values.
function fromInput(int $n, string $s)
{
    restock($n);
    rewind(strlen($s));
}
restock(3);
rewind(-4);

/** @param int<1, 9>|null $digit */
function dial($digit = null) {}
dial(9);
dial(null);

/** @param non-empty-string|int $code */
function redeem($code) {}
redeem('x');
redeem(0);

// Values the native types already reject are InvalidArgumentType's.
function charge(int $cents) {}
charge('x');

// Templates.
/**
 * @template T
 * @param T $value
 */
function keep($value) {}
keep('x');

/**
 * @template T of Order
 * @param T $value
 */
function hold($value) {}
hold('x');

// A variadic parameter's PHPDoc may describe the collected array.
/** @param string[] $parts */
function glue(string ...$parts) {}
glue('a', 'b');

// A PHPDoc type the native type rejects is stale.
/** @param Order $count */
function count_items(int $count) {}
count_items(3);

// By-reference parameters are written, not only read.
/** @param list<int> $ids */
function collect(&$ids) {}
$ids = 'none';
collect($ids);

// The PHPDoc types of PHP's own functions are not consulted.
str_repeat('-', 3);

// The tag of a variadic parameter types each value it collects.
/** @param string ...$names */
function greetAll(...$names) { listNames($names); }
function listNames(array $names) {}

// An array tag on a variadic parameter may type the collected array or
// each value: neither is assumed.
/** @param string[] ...$names */
function greet(...$names) { foreach ($names as $n) { strlen($n); } }
/** @param array<int, string> ...$rows */
function rows(...$rows) { foreach ($rows as $row) { strlen($row); } }

// $this in a closure may be bound to another object.
class Ticket
{
    public string $label = '';

    public function later()
    {
        return function () { rewind($this->label); };
    }
}

// Writes the reaching definitions do not show leave no exact value.
/** @param non-empty-string $code */
function claim(string $code) {}
function viaReference() { $c = 0; $p = &$c; $p = 5; restock($c); }
function viaArrayReference() { $c = 0; $r = [&$c]; $r[0] = 3; restock($c); }
function viaElementReference() { $c = 0; $m['x'] = &$c; $m['x'] = 4; restock($c); }
function viaPropertyReference(\stdClass $obj) { $c = 0; $obj->r = &$c; $obj->r = 4; restock($c); }
function viaGlobal() { global $stock; $stock = 0; refill(); restock($stock); }
function refill() { global $stock; $stock = 5; }
function viaClosure() { $c = 0; $fn = static function (&$x) { $x = 9; }; $fn($c); restock($c); }
function viaInclude() { $c = 0; include 'stock.php'; restock($c); }
function viaAppend() { $s = ''; $p = &$s; $p .= 'x'; claim($s); }

// Literal string values: a listed value, a string of unknown value, and a
// value that may be listed.
/** @param 'asc'|'desc' $direction */
function sortBy(string $direction) {}
function sorting(string $any, bool $flip)
{
    sortBy('asc');
    sortBy("desc");
    sortBy($any);
    sortBy($flip ? 'asc' : 'up');
}

// The code alone types the argument (strlen() returns an int), but it reads
// $this in a closure, which may be bound to another object: what PHPDoc
// says of it is not taken.
/** @param array $parts */
function joinParts($parts) {}
final class Labeler
{
    public string $name = 'x';

    public function later()
    {
        return function () { joinParts(strlen($this->name)); };
    }
}
