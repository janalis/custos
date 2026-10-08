<?php
// '+ 1' on a string or a bool is not what '++' does.
function counters(string $s, bool $flag, int $n, ?int $m, string $t, string $u) {
    $s = $s + 1;
    $t += 1;
    $u = 1 + $u;
    $flag = $flag + 1;
    <weak_warning descr="Use '++$n' instead.">$n = $n + 1</weak_warning>;
    <weak_warning descr="Use '++$n' instead.">$n = 1 + $n</weak_warning>;
    <weak_warning descr="Use '--$m' instead.">$m -= 1</weak_warning>;
    return [$s, $t, $u, $flag, $n, $m];
}
