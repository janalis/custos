<?php
trait Loggable {}
interface Shape {}
class Thing {}

function checks($obj, $p, $q, $untyped, int $n) {
    $a = $obj instanceof self;
    $b = $obj instanceof $q;
    $c = $obj instanceof Unknown;
    $d = $obj instanceof Thing;
    $e = $p == $q && $q;
    $p === $q;
    $map = [$p >= 1, 'k' => 'a' >= 1];
    if (!($untyped < 10)) {}
    if (!($n < 10)) {}
    $s = $p->id === $p->name;
    $x = $p ?? $q;
    $y = ($p + 1) ?? 2;
    $tags = ['a'];
    $msg = 'tags: ' . $tags;
    $z = $p && $q;
    if (($p && $q) || $obj) {}
    if ($p || ($q && $obj)) {}
    if ((!$p) < $q) {}
    if (!$p <=> $q) {}
    while ($r = $p != $q) {}
    if (in_array($p, $tags, $q === 1)) {}
    foo(strlen($p >= 2));
    if (strlen($p >= 2, 1)) {}
}

function untypedRight($p, $limit) {
    if (strlen($p >= $limit)) {}
    if (strlen($p === undefinedCall())) {}
}

function receive() {
    $same = (yield) === (yield);
    $again = (yield from inner()) == (yield from inner());
    return $same && $again;
}

function readTwice($handle, $result) {
    if (readline() === readline()) {}
    if (stream_get_contents($handle) == stream_get_contents($handle)) {}
    if (mysqli_fetch_row($result) != mysqli_fetch_row($result)) {}
    if ((include 'part.php') == (include 'part.php')) {}
    if (`date +%N` == `date +%N`) {}
}
