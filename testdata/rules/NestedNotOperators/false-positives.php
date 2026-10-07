<?php
$one   = !$token;
$paren = !($token);
$cast  = !(bool)$token;
$minus = !-$token;
$bits  = !~$token;
$mixed = !$a && !$b;
$args  = f(!$a, !$b);
