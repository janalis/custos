<?php
function same(callable $mask, string $s): bool
{
    $source = "author: 用户";
    $a = strlen($source) === strlen($mask($source));
    $b = strpos($s, ':') < strlen('é:');
    $c = (strlen('é')) <=> \strlen($s);
    $d = strlen('é') + 1 >= strrpos($s, 'x');
    $e = <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning> === mb_strlen($s);
    $f = <warning descr="Count Unicode characters with the intended counting model.">strlen('é')</warning> > 1;
    return $a && $b && $c && $d && $e && $f;
}
