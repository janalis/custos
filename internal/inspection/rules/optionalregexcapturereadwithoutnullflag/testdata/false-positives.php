<?php
// Guards evaluated before the read in the same expression.
preg_match("/^(a)?b$/", "b", $m);
echo isset($m[1]) ? $m[1] : 'none', "\n";
preg_match("/^(a)?b$/", "b", $n);
if (!empty($n[1]) && $n[1] === 'a') { echo "a\n"; }
preg_match("/^(a)?b$/", "b", $o);
echo !isset($o[1]) || $o[1] === 'a' ? 'ok' : 'ko';
preg_match("/^(a)?b$/", "b", $p);
echo empty($p[1]) ? 'none' : $p[1];
preg_match("/^(a)?b$/", "b", $q);
echo (isset($q[0]) && isset($q[1])) ? $q[1] : 'none';
// By-reference binds create the element.
preg_match("/^(a)?b$/", "b", $r);
$ref = &$r[1];
preg_match("/^(a)?b$/", "b", $s);
$refs = [&$s[1]];
// A whole-array existence call already makes the statement unsupported.
preg_match("/^(a)?b$/", "b", $t);
echo array_key_exists(1, $t) ? $t[1] : 'none';
