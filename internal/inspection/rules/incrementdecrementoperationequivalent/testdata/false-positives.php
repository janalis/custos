<?php

/** @var string $word */
$word[2] += 1;
/** @var \ArrayAccess $bag */
$bag['k'] -= 1;
$unknown[0] += 1;
$count = $count - 3;
$count += 2;
$count = $count + $step;
$count = 1 - $count;
$count = 2 + $count;
$a = $b + 1;
$count *= 1;
$count .= 1;
$count += $count + 1;
$count = ($count + 1);
$count += 1.0;
$count += 0x1;
$count += (1);
$count += true;
// A target with side effects is evaluated twice by `T = T + 1`.
function twice(array $a, $r, $i) {
    $a[$i++] = $a[$i++] + 1;
    $r->cur()->n = $r->cur()->n + 1;
    $a[next_key()] = $a[next_key()] - 1;
    $a[$i = 2] = 1 + $a[$i = 2];
}
