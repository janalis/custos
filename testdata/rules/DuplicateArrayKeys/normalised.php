<?php
$codes = [
    'A' => 'letter',
    <warning descr="Same key and value already present; remove this entry.">"\x41" => 'letter'</warning>,
    'it\'s' => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">"it's"</warning> => 2,
    "tab\t" => 1,
    <warning descr="Same key and value already present; remove this entry.">"tab\x09" => 1</warning>,
    '1' => 'one',
    <warning descr="Key already used earlier; the earlier entry is overwritten.">1</warning> => 'uno',
    16 => 'x',
    <warning descr="Same key and value already present; remove this entry.">0x10 => 'x'</warning>,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'16'</warning> => 'y',
    -3 => 'neg',
    <warning descr="Same key and value already present; remove this entry.">'-3' => 'neg'</warning>,
    0b11 => 'three',
    <warning descr="Key already used earlier; the earlier entry is overwritten.">3</warning> => 'drei',
];
