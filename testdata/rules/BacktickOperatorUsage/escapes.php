<?php
$a = ['k' => 'hello'];
$o = new stdClass; $o->p = ['x' => 'v'];
echo <warning descr="Run the command through shell_exec() instead of backticks.">`echo {$a["k"]}`</warning>;
echo <warning descr="Run the command through shell_exec() instead of backticks.">`echo "{$o->p["x"]}" 'q' ${a["k"]} $a[k]`</warning>;
echo <warning descr="Run the command through shell_exec() instead of backticks.">`echo \"hi\"`</warning>;
echo <warning descr="Run the command through shell_exec() instead of backticks.">`echo \\"hi\\" a\\\`b \x41`</warning>;
