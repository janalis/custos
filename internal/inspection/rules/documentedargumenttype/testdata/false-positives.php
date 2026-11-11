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
