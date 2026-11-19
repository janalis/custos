<?php

class Box
{
    public $value;

    public function run() {}
}

/**
 * @param Box $box
 * @param Box|int $either
 * @param mixed $anything
 * @param object $thing
 * @param iterable $items
 * @param callable $call
 * @param null $nothing
 */
function documented($untyped, $box, $either, $anything, $thing, $items, $call, $nothing)
{
    $untyped->run();
    $box->run();
    $either->run();
    $anything->run();
    $thing->run();
    $items->run();
    $call->run();
    $nothing->run();
}

function proven(int $id)
{
    $id->run();
}

/** @param int $id */
function quietReads($id)
{
    isset($id->value);
    empty($id->value);
    $a = $id->value ?? 1;
}

/** @param int|Box $value */
function narrowed($value)
{
    if ($value instanceof Box) {
        $value->run();
    }
}

function vivified()
{
    $list = false;
    $list[] = 1;
    $list->run();
}

/** @param bool $flag */
function staleTag(?object $flag)
{
    $flag->run();
}
