<?php
$a = ['k' => 'hello'];
$o = new stdClass; $o->p = ['x' => 'v'];
echo shell_exec("echo {$a["k"]}");
echo shell_exec("echo \"{$o->p["x"]}\" 'q' ${a["k"]} $a[k]");
echo shell_exec("echo \\\"hi\\\"");
echo shell_exec("echo \\\"hi\\\" a\\`b \x41");
