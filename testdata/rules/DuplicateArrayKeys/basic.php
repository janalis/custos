<?php
$settings = [
    7 => 'a',
    <warning descr="Same key and value already present; remove this entry.">7 => 'a'</warning>,
    'retries' => 3,
    <warning descr="Same key and value already present; remove this entry.">"retries" => 3</warning>,
    'hosts' => ['db1'],
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'hosts'</warning> => ['db1'],
    'timeout' => 10,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'timeout'</warning> => 20,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'timeout'</warning> => 10,
    'label' => $prefix . 'x',
    <warning descr="Same key and value already present; remove this entry.">'label' => $prefix./* c */'x'</warning>,
    "slot-$n" => 1,
    "slot-$n" => 1,
    'mode' => 'fast',
    'nested' => ['mode' => 'slow'],
];

$legacy = array(
    'a' => $x,
    <warning descr="Same key and value already present; remove this entry.">'a' => $x</warning>,
    'b' => [],
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'b'</warning> => [],
);

['k' => $first, <warning descr="Same key and value already present; remove this entry.">'k' => $first</warning>] = $row;
