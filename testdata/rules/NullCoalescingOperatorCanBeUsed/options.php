<?php
$a = isset($o['k']) ? $o['k'] : 5;
function f($in) {
    <weak_warning descr="Simplify to 'return $in['w'] ?? 'none'' using the null coalescing operator.">if</weak_warning> (isset($in['w'])) {
        return $in['w'];
    }
    return 'none';
}
