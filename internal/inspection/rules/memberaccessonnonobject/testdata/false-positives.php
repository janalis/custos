<?php

class Box
{
    public $value;

    public function run() {}
}

/** @param int $documented */
function declaredElsewhere($untyped, $documented, mixed $anything, object $thing, iterable $items, callable $call, Box $box, ?Box $maybe, Box|int $either)
{
    $untyped->run();
    $documented->run();
    $anything->run();
    $thing->run();
    $items->run();
    $call->run();
    $box->run();
    $maybe->run();
    $either->run();
}

function onlyNull(null $nothing)
{
    $nothing->run();
}

function quietReads(int $id, array $row)
{
    isset($id->value);
    isset($id->value->deeper);
    isset($row->list['key']);
    empty($id->value);
    $a = $id->value ?? 1;
    $b = ($id->value) ?? 2;
}

function narrowed(int|Box $value)
{
    if ($value instanceof Box) {
        $value->run();
    }
}

final class Transports
{
    private ?\WeakMap $map = null;

    public function each(): void
    {
        // A WeakMap's keys are objects; the empty array beside it adds none.
        foreach ($this->map ?? [] as $transport => $kept) {
            $transport->run();
        }
    }
}

function iteratorKeys($untyped, \Traversable $items)
{
    // An iterator may yield keys of any type, objects included.
    foreach ($untyped as $key => $value) {
        $key->run();
    }
    foreach ($items as $key => $value) {
        $key->run();
    }
}

function reflected(array $classes)
{
    // A template bound to objects never takes the array key it was given.
    foreach (\array_keys($classes) as $class) {
        $instance = new \ReflectionClass($class)->newInstanceWithoutConstructor();
        $instance->run();
    }
}
