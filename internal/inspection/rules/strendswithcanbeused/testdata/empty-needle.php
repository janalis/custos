<?php
function tails(string $s, string $t): array {
    return [
        substr($s, -strlen('')) === '',
        "" !== mb_substr($s, -mb_strlen("")),
        substr($s, -strlen('')) === (''),
        <weak_warning descr="Replace with 'str_ends_with($s, '/')'.">substr($s, -strlen('/')) === '/'</weak_warning>,
        <weak_warning descr="Replace with '!str_ends_with($s, $t)'.">substr($s, -strlen($t)) !== $t</weak_warning>,
    ];
}
