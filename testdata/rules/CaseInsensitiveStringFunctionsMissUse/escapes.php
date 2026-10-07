<?php
function split_parts(string $line): array
{
    $a = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($line, "\t")</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">strripos($line, "\x2F")</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">stristr($line, "\u{2014}\n")</weak_warning>;
    $d = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($line, '\\')</weak_warning>;
    $e = stripos($line, "\x41");
    $f = stripos($line, "\u{e9}");
    $g = stripos($line, "\xE9");
    $h = stripos($line, '\t');
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}
