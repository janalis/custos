<?php

function coalesceRight(int $id, $fallback)
{
    return $fallback ?? <error descr="Access this member on an object, not int.">$id</error>->value;
}

function usedAsIndex(int $id, array $map, object $o)
{
    isset($map[<error descr="Access this member on an object, not int.">$id</error>->name]);
    empty($map[<error descr="Access this member on an object, not int.">$id</error>->name]);
    $a = $map[<error descr="Access this member on an object, not int.">$id</error>->name] ?? 0;
    isset($o->{<error descr="Access this member on an object, not int.">$id</error>->name});
}

function plainBinary(int $id)
{
    return <error descr="Access this member on an object, not int.">$id</error>->value + 1;
}
