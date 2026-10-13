<?php

function keysOutOfOrder(): string
{
    return json_encode(['a' => [1 => 'x', 0 => 'y']], JSON_FORCE_OBJECT);
}

function gapInKeys(): string
{
    return json_encode(['a' => [0 => 'x', 2 => 'y']], JSON_FORCE_OBJECT);
}

function negativeKey(): string
{
    return json_encode(['a' => [-1 => 'x', 0 => 'y']], JSON_FORCE_OBJECT);
}

function stringKey(): string
{
    return json_encode(['a' => ['x', 'k' => 'y']], JSON_FORCE_OBJECT);
}

function rewrittenByReference(): string
{
    $payload = ['a' => [1, 2]];
    foreach ($payload as &$entry) {
        $entry = 1;
    }
    unset($entry);

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function rewrittenThroughDestructuring(): string
{
    $payload = ['a' => [[1], [2]]];
    foreach ($payload as $name => [&$first]) {
        $first = 1;
    }

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function foreachTarget(array $rows): string
{
    $payload = ['a' => [1, 2]];
    foreach ($rows as $payload['a']) {
    }

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function capturedByReference(): string
{
    $payload = ['a' => [1, 2]];
    $reset = function () use (&$payload) {
        $payload = [];
    };
    $reset();

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function caught(): string
{
    $payload = ['a' => [1, 2]];
    try {
        risky();
    } catch (Exception $payload) {
    }

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function extracted(array $input): string
{
    $payload = ['a' => [1, 2]];
    extract($input);

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function copiedAfterRewrite(): string
{
    $source = ['a' => [1, 2]];
    foreach ($source as &$entry) {
        $entry = 1;
    }
    unset($entry);
    $payload = $source;
    $source = $payload;

    return json_encode($payload, JSON_FORCE_OBJECT);
}

function nestedVariableRewritten(): string
{
    $tags = [1, 2];
    foreach ($tags as $index => &$tag) {
        unset($tags[$index]);
    }

    return json_encode(['a' => $tags], JSON_FORCE_OBJECT);
}
