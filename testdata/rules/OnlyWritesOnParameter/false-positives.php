<?php
function reads($a, $b, $c, $d, $e, $f, $g, Foo $obj, ?Bar $nullable, int|Baz $union, object $o, &$ref)
{
    $a[] = 1;
    return $a;
}

function more($b, $c, $d, $e, $f, $g, $h, $i)
{
    $b['k'] .= 'x';
    $c['k'] ??= 0;
    $d[] = 1;
    foo($d);
    $e .= 'x';
    echo "$e";
    $f++;
    if (isset($f)) {}
    $g[] = 1;
    $fn = fn() => $g;
    $h[] = 1;
    $cb = function () use ($h) { return $h; };
    $i = $i + 1;
    foo($fn, $cb);
}

function locals()
{
    $tmp = 0;
    foo($x = 1);
    if ($y = bar()) {}
    while ($row = next($rows)) {}
}

abstract class Base
{
    abstract public function run($x);
}

interface Runner
{
    public function go($y);
}

final class Registry
{
    private array $entries = [];

    public function register(string $name, \Closure $run, ?\Closure $guard = null, array $tags = [], $extra = null): void
    {
        $entry = [
            'name' => $name,
            'run' => $run,
            'guard' => $guard,
            'meta' => ['tags' => $tags, 'extra' => [$extra]],
        ];
        $this->entries[] = $entry;
    }
}

function anonymous_class_args($container)
{
    $org = 'acme';
    return function ($c) use ($org) {
        // Constructor arguments of an anonymous class are read in this scope.
        return new class ($c, $org) {
            public function __construct($a, $b) {}
        };
    };
}
