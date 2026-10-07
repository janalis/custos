<?php
function scan(string $line, string $tag) {
    $hit  = <weak_warning descr="Replace with 'str_contains($line, $tag)'.">strpos($line, $tag) !== false</weak_warning>;
    $hit2 = <weak_warning descr="Replace with 'str_contains($line, '#')'.">FALSE !== mb_strpos($line, '#')</weak_warning>;
    $miss = <weak_warning descr="Replace with '!\str_contains(trim($line), $tag)'.">\strpos(trim($line), $tag) === false</weak_warning>;
    return [$hit, $hit2, $miss];
}
