<?php
// Strings PHP treats as numeric keep loose semantics: no strict fix offered.
function numbers($v) {
    return [
        $v == '1e3',
        $v == ' 1',
        $v == '1.',
        $v != '-2.5E-3',
        $v == "7\n",
        $v == "\x31",
        $v == "{$v}",
        $v === '1e',
        $v === '.',
        $v === '0x1A',
        $v === '1_000',
    ];
}
