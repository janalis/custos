<?php
// Before PHP 7.0 `\u{...}` is not an escape: the key keeps its backslash.
$letters = array(
    "\u{41}" => 1,
    'A' => 2,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'\u{41}'</warning> => 3,
    "\\u{42}" => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'\u{42}'</warning> => 2,
    "\x41\u{43}\t" => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">"A\\u{43}\x09"</warning> => 2,
    b"\u{44}" => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'\u{44}'</warning> => 2,
    "\e" => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">"\x1b"</warning> => 2,
    '\e' => 3,
    "\x45" => 1,
    <warning descr="Same key and value already present; remove this entry.">'E' => 1</warning>,
);
