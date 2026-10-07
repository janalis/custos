<?php
function literals($s) {
    $a = str_replace(b'-', '_', $s);
    $b = str_replace(<<<EOT
    -
    EOT, '_', $s);
    $c = strtr($s, <<<'EOT'
    \'
    EOT, '_');
    $f = strtr($s, <<<EOT
    \"
    EOT, '_');
    $g = str_replace(<<<EOT
    \t
    EOT, '_', $s);
    $d = strtr($s, <<<EOT
    EOT, '_');
    $e = strtr($s, <<<EOT
    --
    EOT, '_');
    return [$a, $b, $c, $d, $e, $f, $g];
}
