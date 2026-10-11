<?php

function explicitOrderedKeys(): string
{
    return json_encode(['a' => ['0' => 'x', 1 => 'y']], <warning descr="Preserve nested list representation during JSON encoding.">JSON_FORCE_OBJECT</warning>);
}

function overwrittenInPlace(): string
{
    return json_encode(['a' => [0 => 'x', 1 => 'y', 0 => 'z', 'w']], <warning descr="Preserve nested list representation during JSON encoding.">JSON_FORCE_OBJECT</warning>);
}

function untouchedVariable(array $rows): string
{
    $other = [1, 2];
    $payload = ['a' => [1, 2]];
    foreach ($other as &$entry) {
        $entry = 1;
    }
    foreach ($rows as [$left, $right]) {
        echo $left, $right;
    }
    $log = function () use ($payload) {
        return count($payload);
    };
    try {
        $log();
    } catch (Exception) {
    } catch (Error $error) {
    }
    $copy = $payload;

    return json_encode($copy, <warning descr="Preserve nested list representation during JSON encoding.">JSON_FORCE_OBJECT</warning>);
}
