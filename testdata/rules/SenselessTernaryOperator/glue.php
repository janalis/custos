<?php
function a($v) { return<warning descr="This ternary always yields 'null'; use it directly.">($v===null)?$v:null</warning>; }
function b($v) { return<warning descr="This ternary always yields '\LIMIT'; use it directly.">($v===\LIMIT)?$v:\LIMIT</warning>; }
function c($v) { return<warning descr="This ternary always yields '$v'; use it directly.">($v!==null)?$v:null</warning>; }
$d = <warning descr="This ternary always yields 'LIMIT'; use it directly.">LIMIT !== 'a' ? LIMIT : 'a'</warning>and $e;
$g = <warning descr="This ternary always yields ''a''; use it directly.">LIMIT === 'a' ? LIMIT : 'a'</warning>and $e;
$h = <warning descr="This ternary always yields 'STRLEN($s)'; use it directly.">$max === strlen($s) ? $max : STRLEN($s)</warning>;
$i = <warning descr="This ternary always yields '$k'; use it directly.">$k === (fn() => next($l)) ? (fn() => next($l)) : $k</warning>;
$j = <warning descr="This ternary always yields '$k'; use it directly.">$k === 1.0 ? 1.0 : $k</warning>;
