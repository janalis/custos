<?php
function config(string $path): array
{
    $a = json_decode(file_get_contents($path), true, 512, JSON_THROW_ON_ERROR);
    $b = json_decode(file_get_contents($path), flags: \JSON_BIGINT_AS_STRING | \JSON_THROW_ON_ERROR);
    $c = <warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($path), true)</warning>;
    $d = <warning descr="Handle file reading failure before using its result.">json_decode(file_get_contents($path), true, 512, JSON_BIGINT_AS_STRING)</warning>;
    return [$a, $b, $c, $d];
}
