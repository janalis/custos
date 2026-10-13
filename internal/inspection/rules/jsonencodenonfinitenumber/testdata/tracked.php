<?php

function unrelatedLoop(array $rows): string
{
    $stats = ['min' => INF];
    foreach ($rows as &$row) {
        $row = 0;
    }
    $count = function () use (&$rows) {
        return count($rows);
    };

    return <error descr="Encode finite numbers in JSON.">json_encode($stats)</error>;
}

function nestedVariable(): string
{
    $limit = -INF;
    $stats = ['min' => $limit];

    return <error descr="Encode finite numbers in JSON.">json_encode(($stats))</error>;
}

function readOnlyLoop(): string
{
    $stats = ['min' => NAN];
    foreach ($stats as $key => $value) {
        echo $key, $value;
    }

    return <error descr="Encode finite numbers in JSON.">json_encode($stats)</error>;
}
