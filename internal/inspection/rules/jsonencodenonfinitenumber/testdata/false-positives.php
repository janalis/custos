<?php

function sanitisedByReference(): string
{
    $stats = ['min' => INF, 'max' => -INF];
    foreach ($stats as $key => &$value) {
        $value = is_finite($value) ? $value : null;
    }
    unset($value);

    return json_encode($stats);
}

function foreachTarget(array $rows): string
{
    $stats = ['min' => INF];
    foreach ($rows as $stats['min']) {
    }

    return json_encode($stats);
}

function capturedByReference(): string
{
    $stats = ['min' => INF];
    $reset = function () use (&$stats) {
        $stats = [];
    };
    $reset();

    return json_encode($stats);
}

function extracted(array $input): string
{
    $stats = ['min' => INF];
    extract($input);

    return json_encode($stats);
}

function nestedSanitised(): string
{
    $limits = [INF, NAN];
    foreach ($limits as &$limit) {
        $limit = 0;
    }
    unset($limit);

    return json_encode(['limits' => $limits]);
}

$top = ['min' => INF];
foreach ($top as &$entry) {
    $entry = 0;
}
unset($entry);
json_encode($top);
