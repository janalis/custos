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
    $same = <error descr="Both operands are the same.">$a instanceof $a</error>;
    $items = [1 >= 2, `ls` >= 1, "k$x" <error descr="Did you mean '=>' for an array key?">>=</error> 1];
    if (<error descr="Null or false operands make this negated comparison misleading; use '$flag > 1'.">!($flag <= 1)</error>) {}
    if (in_array($x, $arr <> 1)) {}
    $obj = new ArrayObject($x == 1);
    if (unknown_fn($x == 1)) {}
    if (strlen($x, $x == 1)) {}
    if (doc_param($x <error descr="This comparison probably belongs outside the call parentheses.">==</error> 1)) {}
    if (strlen($x == 'a')) {}
    if ($box->count($n <error descr="This comparison probably belongs outside the call parentheses.">===</error> 2)) {}
    if (doc_return($x <error descr="This comparison probably belongs outside the call parentheses.">==</error> 1)) {}
    $neg = -$x ?? 1;
    $c = $a && PHP_EOL;
}
