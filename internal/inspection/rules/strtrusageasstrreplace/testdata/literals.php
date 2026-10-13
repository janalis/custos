<?php
function literals($s) {
    $a = <weak_warning descr="Use 'str_replace(b'-', '_', $s)' instead.">strtr($s, b'-', '_')</weak_warning>;
    $b = <weak_warning descr="Use 'str_replace(&lt;&lt;&lt;EOT
    -
    EOT, '_', $s)' instead.">strtr($s, <<<EOT
    -
    EOT, '_')</weak_warning>;
    $c = strtr($s, <<<'EOT'
    \'
    EOT, '_');
    $f = strtr($s, <<<EOT
    \"
    EOT, '_');
    $g = <weak_warning descr="Use 'str_replace(&lt;&lt;&lt;EOT
    \t
    EOT, '_', $s)' instead.">strtr($s, <<<EOT
    \t
    EOT, '_')</weak_warning>;
    $d = strtr($s, <<<EOT
    EOT, '_');
    $e = strtr($s, <<<EOT
    --
    EOT, '_');
    return [$a, $b, $c, $d, $e, $f, $g];
}
