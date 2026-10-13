<?php
$load  = <warning descr="Run the command through shell_exec() instead of backticks.">`cat /proc/loadavg`</warning>;
$found = <warning descr="Run the command through shell_exec() instead of backticks.">`grep -c "ERROR" app.log`</warning>;
function stamp() {
    return <warning descr="Run the command through shell_exec() instead of backticks.">`date +%s`</warning>;
}
function glued() {
    return<warning descr="Run the command through shell_exec() instead of backticks.">`ls $dir`</warning>;
}
$nested = <warning descr="Run the command through shell_exec() instead of backticks.">`echo \`date\``</warning>;
