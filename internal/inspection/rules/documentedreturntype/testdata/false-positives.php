<?php
declare(strict_types=1);

namespace Shop;

class Order {}

// Ranges, lengths and literal values count only against exact values.
/** @return positive-int */
function quantity(int $n)
{
    if ($n < 0) {
        return $n;
    }
    return 1;
}

/** @return non-empty-string */
function code(string $s)
{
    return $s;
}

/** @return negative-int */
function before(bool $far)
{
    return $far ? -10 : -1;
}

/** @return int<0, 9> */
function bounded(int $n)
{
    return $n > 9 ? 9 : 0;
}

/** @return non-empty-string */
function name(bool $long)
{
    return $long ? 'a long name' : "x";
}

/** @return 'asc'|'desc' */
function order(bool $down)
{
    return $down ? 'desc' : 'asc';
}

/** @return 'asc'|'desc' */
function direction(string $input)
{
    return $input;
}

// Values the native type already rejects are InvalidReturnType's.
/** @return int */
function total(): int
{
    return 'x';
}

// Templates of the function or of its class.
/**
 * @template T
 * @param T $value
 * @return T
 */
function keep($value)
{
    return 'x';
}

/** @template TItem */
class Box
{
    /** @return list<TItem> */
    public function items()
    {
        return 'none';
    }
}

// Conditional return types.
/** @return ($asText is true ? string : int) */
function size(bool $asText)
{
    return [];
}

// A @return type the native type rejects is stale.
/** @return Order */
function count_items(): int
{
    return 3;
}

// A tag inherited from a parent declaration is not this method's.
class Repository
{
    /** @return Order */
    public function find()
    {
        return new Order();
    }
}

class LegacyRepository extends Repository
{
    public function find()
    {
        return 'order-1';
    }
}

class CachedRepository extends Repository
{
    /** {@inheritDoc} */
    public function find()
    {
        return 'order-1';
    }
}

// Relative types in a trait name the class using it.
trait Fluent
{
    /** @return $this */
    public function with()
    {
        return new Order();
    }
}

// void and generators.
/** @return void */
function done()
{
    return;
}

/** @return negative-int */
function rows(bool $two)
{
    $rows = $two ? [1, 2] : [1];
    return count($rows);
}

/** @return \Generator<int, string> */
function names()
{
    yield 'a';
    return 3;
}

// $this in a closure may be bound to another object.
class Ticket
{
    public string $label = '';

    public function later()
    {
        return fn (): int => $this->label;
    }
}
