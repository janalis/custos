<?php
class Box
{
    public function has(string $k): bool { return true; }
    public function count(int $n): int { return $n; }
}

/** @param string $s */
function doc_param($s): int { return strlen($s); }

/** @return int */
function doc_return($s) { return 1; }

function more($a, $x, ?bool $flag, Box $box, array $arr, int $n) {
    $same = $a instanceof $a;
    $items = [1 >= 2, `ls` >= 1, "k$x" >= 1];
    if ($flag > 1) {}
    if (in_array($x, $arr <> 1)) {}
    $obj = new ArrayObject($x == 1);
    if (unknown_fn($x == 1)) {}
    if (strlen($x, $x == 1)) {}
    if (doc_param($x) == 1) {}
    if (strlen($x == 'a')) {}
    if ($box->count($n) === 2) {}
    if (doc_return($x) == 1) {}
    $neg = -$x ?? 1;
    $c = $a && PHP_EOL;
}
