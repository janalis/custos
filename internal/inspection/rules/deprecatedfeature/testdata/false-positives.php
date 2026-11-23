<?php
namespace Shop;

// Replacements and look-alikes.
$id = spl_object_id($order);
is_float($x);
is_double($x); // AliasFunctionsUsage reports it
sort($names, SORT_STRING);
define('LIMIT', 10);
is_a($object, Cart::class);
is_a($class, Cart::class, true);
is_subclass_of($class, Cart::class);
array_walk($items, 'trim');

// A function of the project with the same name, in the namespace.
function metaphone(string $s): string { return $s; }
metaphone('Thompson');

// A method of an unrelated or unknown class, and an override.
function rows($unknown, Basket $basket, SortedIterator $sorted): void
{
    $unknown->asort();
    $basket->asort();
    $sorted->asort();
}
final class Basket { public function asort(): void {} }
final class SortedIterator extends \ArrayIterator { public function asort(int $flags = SORT_REGULAR): true { return true; } }

final class Cart
{
    const LET = 1;
    public function __construct()
    {
        $f = function () { return 1; };
        return;
    }
    public function __debugInfo(): array { return []; }
    public function is(): bool { return true; }
    public function readonly(): void {}
    public function total(): int
    {
        try {
            return 1;
        } finally {
            $cleanup = static fn () => 0;
            $later = function () { return 2; };
        }
    }
}
new \mysqli_stmt($db, 'SELECT 1');
$it = new \ArrayIterator([]);
$it->$method();
define($name, 1);
class_alias(Cart::class, $alias);
final class Ticket { public function __debugInfo(): array|false { return []; } }
final class Bag { public function __debugInfo(): mixed { return []; } }
