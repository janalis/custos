<?php
function f($a, $b, DateTime $d, \PDO $pdo) {
    return [
        $a === $b,
        $a !== 'x',
        $a < 'x',
        $d == $a,
        $pdo != $b,
    ];
}
