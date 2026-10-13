<?php
// D1 ordering
if (strlen($name) > 3 && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$enabled</weak_warning>) {}
if ($repo->find($id) || empty($cache)) {}      // method call: may have side effects
if (\json_encode($id) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">empty($cache)</weak_warning>) {}
if (isset($rows[md5($k)]) && $rows) {}        // S3: isset also guards $rows itself
if (isset($rows[md5($k)]) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$limit</weak_warning>) {}
if (!(trim($v) && (<weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$v</weak_warning>))) {}
if ($ok) {} elseif (json_decode($raw) || <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">is_string($w)</weak_warning>) {}
if (preg_match('/^v\d/', $tag) && <weak_warning descr="Cheaper check placed after a costlier one; evaluate it first.">$strict</weak_warning>) {}
if ($enabled && strlen($name) > 3) {}
if (is_int($n) && $n) {}
if (($total = $cart->sum()) && $total > 10) {}
if (count($items) > 0 && $items[0]) {}
if (($head = array_pop($stack)) && $stack) {}
if (!isset($map[lower($key)]) && !array_key_exists($key, $map)) {}
if (fetch($a) <weak_warning descr="Use '&&' instead of 'and'.">and</weak_warning> $b) {} // one operand for D1

// D2 keyword operators
if ($p <weak_warning descr="Use '&&' instead of 'and'.">AND</weak_warning> $q) {}
if ($p || ($q <weak_warning descr="Use '||' instead of 'or'.">or</weak_warning> $r)) {}
if ($p xor $q) {}

// D3 equality next to instanceof
if (<weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$node != null</weak_warning> && $node instanceof Leaf && $node->ok) {}
if ($node instanceof Leaf || $node === null) {}

// D4 redundant instanceof
interface Shape {}
class Circle implements Shape {}
class Ring extends Circle {}
if ($s instanceof Shape || <warning descr="Redundant instanceof: another check on the same value already covers this type.">$s instanceof Ring</warning>) {}
if ($s instanceof Circle && <warning descr="Redundant instanceof: another check on the same value already covers this type.">$s instanceof Shape</warning>) {}
if (<warning descr="Redundant instanceof: another check on the same value already covers this type.">$s instanceof Shape</warning> && $s instanceof Ring) {}
if ($s instanceof Circle || $t instanceof Shape) {}
