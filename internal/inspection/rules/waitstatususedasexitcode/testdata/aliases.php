<?php
use function pcntl_waitpid as builtinCall11;
$pid=builtinCall11($child,$status); if (<warning descr="Decode the child wait status.">$status===3</warning>) { echo "exit three"; }
