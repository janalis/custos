<?php

function destructuredRoot(): string
{
    $a = [];
    $a['self'] = &$a;
    [$a] = [[1]];
    return json_encode($a);
}

function foreachRoot(): string
{
    $a = [];
    $a['self'] = &$a;
    foreach ([[1]] as $a) {
    }
    return json_encode($a);
}

function foreachSlot(): string
{
    $a = [];
    $b = [];
    $a['b'] = &$b;
    $b['a'] = &$a;
    foreach ([1] as $b['a']) {
    }
    return json_encode($a);
}

function destructuredSlot(): string
{
    $a = [];
    $b = [];
    $a['b'] = &$b;
    $b['a'] = &$a;
    [$b['a']] = [1];
    return json_encode($a);
}

function writeThroughAlias(): string
{
    $a = [];
    $d = [];
    $a['self'] = &$a;
    $d['x'] = &$a;
    $d['x'] = 5;
    return json_encode($a);
}

function dynamicVariable(string $name): string
{
    $a = [];
    $a['self'] = &$a;
    $$name = [1];
    return json_encode($a);
}
