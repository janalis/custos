<?php
// '+ 1' on a string or a bool is not what '++' does.
function counters(string $s, bool $flag, int $n, ?int $m, string $t, string $u) {
    $s = $s + 1;
    $t += 1;
    $u = 1 + $u;
    $flag = $flag + 1;
    ++$n;
    ++$n;
    --$m;
    return [$s, $t, $u, $flag, $n, $m];
}
