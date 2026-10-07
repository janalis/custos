<?php
function pickRate($rate, $level) {
    $a = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($rate, ['-1', '1.5'])</weak_warning>;
    $b = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">array_search($rate, ['.5', '1e3'])</weak_warning>;
    $c = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($rate, ['low', ' +2 '])</weak_warning>;
    $d = in_array($level, ['1.0.0', 'e5', '-']);
    return [$a, $b, $c, $d];
}
