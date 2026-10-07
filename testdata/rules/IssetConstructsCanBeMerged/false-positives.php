<?php
$n1 = isset($p) && !isset($q);
$n2 = isset($p) || isset($q);
$n3 = !isset($p) && !isset($q);
$n4 = isset($p) and isset($q);
$n5 = !isset($p) or !isset($q);
$n6 = !(isset($p)) || !isset($q);
$n7 = isset($p) && $q;
$n8 = isset($p) xor isset($q);
$n9 = (bool) isset($p) || !isset($q);
$n10 = isset($p) || ($ok && isset($q));
