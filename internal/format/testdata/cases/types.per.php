<?php
// presets: per laravel
function f(?Foo $a, Foo|Bar|null $b, Foo&Bar $c, (A&B)|null $d, int|string $e = 1, callable ...$f): static|null {}
class Hooks
{
    public string $name { get => $this->first . ' ' . $this->last; set(string $v) { $this->first = $v; } }
    public function __construct(public int $x { get; set; }) {}
    final public const int X = 1;
}
$fn = fn&(array &$a): ?array => $a;
function &ref()
{
    static $x;
    return $x;
}
$x = $a?->b?->c ?? $d ?: $e;
$y = (A & B)::class;
abstract class X
{
    abstract protected function g(int|float $n): never;
}
