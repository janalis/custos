<?php
// The guard tests another capture, has the wrong polarity, or runs too late.
preg_match("/^(a)?b$/", "b", $m);
echo isset($m[0]) ? <warning descr="Guard the unmatched optional capture or request null placeholders.">$m[1]</warning> : 'none';
preg_match("/^(a)?b$/", "b", $n);
echo isset($n[1]) ? 'set' : <warning descr="Guard the unmatched optional capture or request null placeholders.">$n[1]</warning>;
preg_match("/^(a)?b$/", "b", $o);
echo isset($o[1]) || <warning descr="Guard the unmatched optional capture or request null placeholders.">$o[1]</warning> === 'a' ? 'ok' : 'ko';
preg_match("/^(a)?b$/", "b", $p);
echo <warning descr="Guard the unmatched optional capture or request null placeholders.">$p[1]</warning> === 'a' && isset($p[1]) ? 'ok' : 'ko';
preg_match("/^(a)?b$/", "b", $q);
echo $fallback ?? <warning descr="Guard the unmatched optional capture or request null placeholders.">$q[1]</warning>;
preg_match("/^(a)?b$/", "b", $r);
echo (isset($r[1]) || $loose) ? <warning descr="Guard the unmatched optional capture or request null placeholders.">$r[1]</warning> : 'none';
preg_match("/^(a)?b$/", "b", $s);
echo -$count && <warning descr="Guard the unmatched optional capture or request null placeholders.">$s[1]</warning> ? 'ok' : 'ko';
preg_match("/^(a)?b$/", "b", $t);
echo isset($other, $other[1], $t[$i], $t->p[1]) && <warning descr="Guard the unmatched optional capture or request null placeholders.">$t[1]</warning> ? 'ok' : 'ko';
preg_match("/^(a)?b$/", "b", $u);
$copy = [<warning descr="Guard the unmatched optional capture or request null placeholders.">$u[1]</warning>];
preg_match("/^(a)?b$/", "b", $v);
$plain = <warning descr="Guard the unmatched optional capture or request null placeholders.">$v[1]</warning>;
preg_match("/^(a)?b$/", "b", $w);
echo <warning descr="Guard the unmatched optional capture or request null placeholders.">$w[1]</warning> ? 'ok' : 'ko';
