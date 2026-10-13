<?php
$m = new WeakMap(); $o = new stdClass(); <warning descr="Store a value that does not retain the WeakMap key.">$m[$o] = $o</warning>;
