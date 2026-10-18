<?php
// Code the scope evaluates leaves only the guards written around the read.
function thenGuard($name, $code) { $a = posix_getpwnam($name); echo $a ? $a['dir'] : ''; eval($code); }
function elseGuard($name, $code) { $a = posix_getpwnam($name); echo !$a ? '' : $a['dir']; eval($code); }
function andGuard($name, $code) { $a = posix_getpwnam($name); $a && print($a['dir']); eval($code); }
