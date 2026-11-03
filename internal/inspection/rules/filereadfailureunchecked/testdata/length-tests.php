<?php

// A length test tells the failure, read as the empty text, apart.
function restore(string $path): ?string
{
    $content = file_get_contents($path);
    if (48 > \strlen($content) || substr($content, -24) !== substr($content, 0, 24)) {
        return null;
    }
    return $content;
}

function variants(string $path): array
{
    $a = file_get_contents($path);
    $b = file_get_contents($path);
    $c = file_get_contents($path);
    $d = file_get_contents($path);
    $e = file_get_contents($path);
    $out = [];
    if ((strlen($a)) > 0x0) {
        $out[] = trim($a);
    }
    if (mb_strlen(($b)) >= 1) {
        $out[] = trim($b);
    }
    if (0 === strlen($c)) {
        return $out;
    }
    if (strlen($d) <= 0) {
        return $out;
    }
    if (!(strlen($e) < 2)) {
        $out[] = trim($e);
    }
    $out[] = trim($c);
    $out[] = trim($d);
    return $out;
}

// Comparisons that do not tell the empty text apart.
function loose(string $path, int $n): array
{
    $f = file_get_contents($path);
    return [
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> >= 0,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> < 0,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> == 3,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> > $n,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> + 1,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> > 1.5,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> <=> 0,
        <warning descr="Handle file reading failure before using its result.">strlen($f)</warning> > 99999999999999999999,
        <warning descr="Handle file reading failure before using its result.">strtoupper($f)</warning> > 0,
    ];
}

function after(string $path): string
{
    $g = file_get_contents($path);
    if (strlen($g) > 0) {
        echo 'ok';
    }
    return <warning descr="Handle file reading failure before using its result.">trim($g)</warning>;
}
