<?php
function seekCases($fp, $x, array $a) {
    if (fseek(...)) { echo "closure"; }
    if ($x += fseek($fp, 0)) { echo "sum"; }
    if ($x ??= fseek($fp, 0)) { echo "default"; }
    if ($a[fseek($fp, 0)] = 1) { echo "key"; }
    for (fseek($fp, 0); fseek($fp, 1), $x; fseek($fp, 2)) { break; }
    if (-1 === fseek($fp, 0)) { echo "failed"; }
}
