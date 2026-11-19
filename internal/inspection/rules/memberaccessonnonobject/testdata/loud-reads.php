<?php

function coalesceRight(int $id, $fallback)
{
    return $fallback ?? <error descr="Access this member on an object, not int.">$id</error>->value;
}

function plainBinary(int $id)
{
    return <error descr="Access this member on an object, not int.">$id</error>->value + 1;
}
