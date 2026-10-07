<?php
function pickRate($rate, $level) {
    $a = in_array($rate, ['-1', '1.5'], true);
    $b = array_search($rate, ['.5', '1e3'], true);
    $c = in_array($rate, ['low', ' +2 '], true);
    $d = in_array($level, ['1.0.0', 'e5', '-']);
    return [$a, $b, $c, $d];
}
