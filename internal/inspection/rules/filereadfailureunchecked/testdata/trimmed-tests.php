<?php

// The failure value trims to the empty text, which an emptiness test
// sends where an empty file goes.
function inline(string $file): array
{
    if ('' === $content = trim(file_get_contents($file))) {
        return [];
    }
    return explode("\n", $content);
}

function stored(string $file): array
{
    $content = rtrim(file_get_contents($file));
    if ($content === '') {
        return [];
    }
    return explode("\n", $content);
}

function forms(string $file): array
{
    $out = [];
    if (empty(ltrim(file_get_contents($file)))) {
        $out[] = 'empty';
    }
    if (!(trim(file_get_contents($file)))) {
        $out[] = 'blank';
    }
    $out[] = trim(file_get_contents($file)) ?: 'none';
    if ($line = trim(file_get_contents($file))) {
        $out[] = $line;
    } elseif (trim(file_get_contents($file)) != "") {
        $out[] = 'never';
    } elseif (trim(file_get_contents($file))) {
        $out[] = 'still never';
    }
    return $out;
}

// Other uses of the trimmed text.
function untested(string $file, string $other): array
{
    $a = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning>;
    echo $a;
    $b = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning>;
    $b = 'x';
    $c = [];
    $c[] = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning>;
    $d = <warning descr="Handle file reading failure before using its result.">strtoupper(file_get_contents($file))</warning> === '';
    $e = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning> === 'x';
    $f = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning> === $other;
    $g = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning> > '';
    $h .= <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning>;
    $i = <warning descr="Handle file reading failure before using its result.">trim(file_get_contents($file))</warning>;
    return [$b, $c, $d, $e, $f, $g, $h];
}
