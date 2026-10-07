<?php
trait Loggable {}
interface Shape {}

function checks($obj, $p, $q, array $tags, ?int $limit) {
    $t1 = $obj instanceof Loggable;
    $t2 = $obj instanceof Shape;
    $s1 = $p->id === ($p->id);
    $s2 = count($tags) > count($tags);
    $p == $q;
    $map = ['enabled' >= true, $p >= 'x'];
    $out = [
        (bool)$p ?? 0,
        (!$q) ?? 1,
        $p ?? $q,
    ];
    $msg = 'tags: ' . ['a'];
    $ok = [
        $p and FALSE,
        $p || true,
        $p && \true,
        $p or null,
    ];
    if ($limit >= 10) {}
}

class Quota {
    public function fits(string $who): bool {
        if (strlen($who) >= 3) {}
        if ($this->allow($who === 'root')) {}
        return true;
    }
    private function allow(bool $flag) { return $flag; }
}

if ($found = ($left != $right)) {}
if ($x || ($y && $z)) {}
if ($row = (fetch() || $fallback)) {}
if ((!$x) >= $limit) {}
if (($x && $y) || $z) {}
$both = $x && $y;
if ((!$x) < $y) {}
