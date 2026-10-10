<?php
$m=new WeakMap();$key=new stdClass();$other=new stdClass();<warning descr="Store a value that does not retain the WeakMap key.">$m[$key]=[$key]</warning>; $m[$other]=["text"];
