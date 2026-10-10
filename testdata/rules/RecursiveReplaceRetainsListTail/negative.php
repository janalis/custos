<?php
$c=array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c","d"]]); echo $c["hosts"][1];
