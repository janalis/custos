<?php
function trimTail(string $line, int $from) {
    $broken = substr($line, $from, strlen($line) - $from, 'UTF-8');
    $alsoBroken = substr($line, 0, strlen($line) - 3, null);
    $fine = mb_substr($line, $from, null, 'UTF-8');
    return [$broken, $alsoBroken, $fine];
}
