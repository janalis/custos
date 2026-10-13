<?php namespace Independent;
function posix_getpwnam(){}
function posix_getpwuid(){}

$a=posix_getpwnam($name);echo $a["uid"];echo posix_getpwuid(5)["name"];
