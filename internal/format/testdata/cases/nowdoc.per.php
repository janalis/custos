<?php
// presets: per; set: heredoc_to_nowdoc=true, single_quote=true, yoda=yoda, increment_style=pre
$a = <<<'EOT'
    plain text
    EOT;
$b = <<<'EOT'
    quoted label
    EOT;
$c = <<<EOT
    has \t escape
    EOT;
if (Foo::BAR == $x || -1 === $y || [] != $z || __LINE__ !== $w) {}
if ($a = 1 == $b) {}
if ($p == $q) {}
if ($m ? 1 : 2 == 3) {}
if (foo(
    $x,
) == 1) {}
$i /* c */ ++;
$s = "\\";
