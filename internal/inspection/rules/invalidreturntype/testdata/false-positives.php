<?php
declare(strict_types=1);

namespace Shop;

class Order {}
final class Coupon {}
interface Shippable {}

// Generators: a return gives the completion value, not the call's.
function ids(): \Generator
{
    yield 1;
    return 'done';
}

function delegated(): iterable
{
    yield from [1, 2];
    return 'done';
}

$lazy = fn(): \Generator => yield 'x';

// never and void.
function fail(): never
{
    throw new \RuntimeException();
}

function log_line(string $line): void
{
    if ($line === '') {
        return;
    }
    echo $line;
}

// A closure's return belongs to the closure, not to the function around it.
function handler(): \Closure
{
    return function () {
        return 'text';
    };
}

function mapper(): array
{
    return array_map(fn($n) => (string) $n, [1, 2]);
}

// $this in a closure may be bound to another object.
class Builder
{
    public function binder(): \Closure
    {
        return function (): Builder {
            return $this;
        };
    }
}

// Relative types in a trait name the class using it.
trait Fluent
{
    public function with(): static
    {
        return $this;
    }

    public function same(): self
    {
        return new static();
    }
}

// Values some of which are accepted, or of unknown type.
function pick(bool $first, Order $order): ?Order
{
    return $first ? $order : null;
}

function anything($value): int
{
    return $value;
}

function mixedValue(mixed $value): string
{
    return $value;
}

function downcast(Order $order): Shippable
{
    return $order; // a subclass of Order may implement Shippable
}

/** @return list<int> */
function documented(): array
{
    return [1, 2];
}

// Classes whose hierarchy is not known.
function vendor(): \Vendor\Model
{
    return new \Vendor\Other();
}

// PHPDoc types alone do not prove a mismatch.
/** @param string $id */
function lookup($id): int
{
    return $id;
}

// Locals passed by reference may change unseen.
function fill(array &$into): void {}
function filled(): array
{
    $rows = null;
    fill($rows);
    return $rows;
}

// Covariant declarations: a subtype narrows its parent's return type.
class Repository
{
    public function find(): ?Order
    {
        return null;
    }
}

class OrderRepository extends Repository
{
    public function find(): Order
    {
        return new Order();
    }
}

// Templates.
/**
 * @template T
 * @param T $value
 * @return T
 */
function keep($value)
{
    return 'x';
}

// Top-level returns end an included file.
return ['key' => 'value'];
