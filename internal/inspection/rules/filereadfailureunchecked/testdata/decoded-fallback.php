<?php

function runtimeConfig(string $file): array
{
    return (is_file($file) ? json_decode(file_get_contents($file), true) : [])['extra']['runtime'] ?? [];
}

function decodedChecks(string $file): array
{
    $a = json_decode(file_get_contents($file)) ?? new stdClass();
    $b = isset(json_decode(file_get_contents($file))->name);
    $c = empty(json_decode(file_get_contents($file), true)['items']);
    $d = <warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file), true)</warning>['extra'];
    $e = json_decode(file_get_contents($file), true) ? 1 : 0;
    $f = [<warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file), true)</warning>] ?? [];
    $g = <warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file), true)</warning> + 1 ?? 0;
    $h = $a->{<warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file))</warning>} ?? null;
    $i = $c[<warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file))</warning>] ?? null;
    $j = $file ? <warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($file))</warning> : null;

    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j];
}
