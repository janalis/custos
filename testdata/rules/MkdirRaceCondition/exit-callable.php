<?php
function prepareAfterCallable($root) {
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($root) || is_dir(...)'.">mkdir($root)</error> or exit(...);
    <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($root) || is_dir(...)'.">mkdir($root)</error> or die(...);
    mkdir($root) or exit(1);
    mkdir($root) or ((exit(...))(1));
}
