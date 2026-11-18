<?php
$c=array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c","d"]]); echo $c["hosts"][1];
$d=array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]],["hosts"=>[1=>"z"]]); echo $d["hosts"][1];
$e=array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]],["hosts"=>"gone"]); echo $e["hosts"][1];
