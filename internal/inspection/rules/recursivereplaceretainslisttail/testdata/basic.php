<?php
$c=<warning descr="Replace the list explicitly to remove its old tail.">array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]])</warning>; echo $c["hosts"][1];
$d=<warning descr="Replace the list explicitly to remove its old tail.">array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]],["port"=>80])</warning>; echo $d["hosts"][1];
