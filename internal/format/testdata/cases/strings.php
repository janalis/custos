<?php // presets: per symfony; risky
$a = "x {$b['k']} ${c} $d->e {$f->g()} \$h";
$n = <<<EOT
  {$a}
  EOT . "x" . <<<'NOW'
    raw
    NOW;
$m = `cmd {$x}`;
$o = "nested {$a("y")}";
$p = "{$a->{'b'}}";
$q = "{$f(new Foo, array(1), "dq", $a == 1)} {$g(function () { return 1; })}";
echo <<<EOT
A
EOT;
__halt_compiler(); raw   data
