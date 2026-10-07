<?php
function tally(int $count, string $label) {
    $count = 5 + $count;          // reversed operands
    $count = $count - $a - $b;    // '-' cannot be regrouped
    $count = $count / $a / $b;
    $count = $count + $k * 2;     // fragment is a binary expression
    $count = $count ** 2;         // '**' not eligible
    $count = $count ?? 2;
    $count = $count && 2;
    $label[0] = $label[0] . 'x';  // string offset (typed string)
    $count += $count + 1;         // compound assignment
    $count = $other + 1;
    $count = $count . 'a' + 1;
    [$a, $b] = $a + 1;
}
