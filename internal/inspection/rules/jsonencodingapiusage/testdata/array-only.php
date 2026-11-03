<?php

// A failed decode gives null, which a parameter of PHP's own functions that
// only accepts an array rejects with a TypeError.
function keys(string $raw, array $base): array
{
    $a = array_keys(json_decode($raw, true));
    $b = array_merge($base, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
    $c = count(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
    $d = array_search('x', json_decode($raw, true));
    $e = strtoupper(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw)</weak_warning>);
    $f = mine(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
    try {
        $g = array_keys(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>);
    } catch (\TypeError $e) {
        $g = [];
    }
    $item = json_decode($raw, true);
    $h = array_column($item['rows'], 'id');
    $i = array_keys(json_decode($raw, true)['rows']);
    $other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>;
    $j = count($other['rows']);
    // A failed encode gives the empty text, which fails the decode it feeds.
    $k = array_keys(json_decode(json_encode($base), true));
    $l = json_decode((@json_encode($base)), true, 512, JSON_THROW_ON_ERROR);
    $m = json_decode(flags: JSON_THROW_ON_ERROR, json: json_encode($base));
    $n = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($base)</weak_warning>, true)</weak_warning>;
    $o = array_keys(json_decode($raw, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($base)</weak_warning>));
    $p = strtoupper(<weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($base)</weak_warning>);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l, $m, $n, $o, $p];
}

function mine(array $a): array
{
    return $a;
}
