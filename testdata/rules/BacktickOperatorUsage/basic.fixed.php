<?php
$load  = shell_exec("cat /proc/loadavg");
$found = shell_exec("grep -c \"ERROR\" app.log");
function stamp() {
    return shell_exec("date +%s");
}
function glued() {
    return shell_exec("ls $dir");
}
$nested = shell_exec("echo `date`");
