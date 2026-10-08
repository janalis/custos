<?php
class Node { public $x; }

function label(Node $o, bool $c): string
{
    $f = $o->x;
    if ($c) {
        $f = 'x';
    }
    return (string) $f;          // $f may hold anything
}

function known(bool $c): string
{
    $g = 'y';
    if ($c) {
        $g = 'z';
    }
    return <weak_warning descr="Operand already has the target type; remove the cast.">(string)</weak_warning> $g;
}
