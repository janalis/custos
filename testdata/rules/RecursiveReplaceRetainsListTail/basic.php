<?php
$c=<warning descr="Replace the list explicitly to remove its old tail.">array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]])</warning>; echo $c["hosts"][1];
