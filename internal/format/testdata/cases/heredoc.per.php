<?php
// presets: per laravel; risky
$a = <<<EOT
    hello
    EOT;
$b = <<<"EOT"
  x {$y}
  EOT;
$c = sprintf(<<<'SQL'
    SELECT 1
    SQL, $x);
$d = `ls $dir`;
$e = "a\n";
$f = "it's";
