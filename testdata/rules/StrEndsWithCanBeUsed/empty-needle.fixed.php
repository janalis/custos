<?php
function tails(string $s, string $t): array {
    return [
        substr($s, -strlen('')) === '',
        "" !== mb_substr($s, -mb_strlen("")),
        substr($s, -strlen('')) === (''),
        str_ends_with($s, '/'),
        !str_ends_with($s, $t),
    ];
}
