<?php
// The loop starts from the value the caller gave.
function line_breaks(array $tokens, int $index, int $direction)
{
    for ($index += $direction; isset($tokens[$index]); $index += $direction) {
    }
    $next = function ($i) {
        for ($i = $i + 1; $i < 10; ++$i) {
        }
        return $i;
    };
    for ($pos = 0; $pos < $index; ++$pos) {
        for ($pos = $pos * 2; $pos < $index; ++$pos) {
        }
    }
    return [$index, $next];
}

function restart(int $index, int $from)
{
    <error descr="Loop variable '$index' overwrites a function parameter.">for</error> ($index = $from; $index < 10; ++$index) {
    }
    <error descr="Loop variable '$from' overwrites a function parameter.">for</error> ($from = 0, $$index = 1; $from < 10; ++$from) {
    }
}
