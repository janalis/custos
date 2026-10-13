<?php
// Before PHP 5.4 `\e` is not an escape either.
$controls = array(
    "\e[0m" => 1,
    "\x1b[0m" => 2,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'\e[0m'</warning> => 3,
    "\u{41}\e" => 1,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'\u{41}\e'</warning> => 2,
);
