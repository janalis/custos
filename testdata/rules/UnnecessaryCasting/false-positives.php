<?php
/** @param int $n */
function docOnly($n, ?int $m) {
    $x = $undefined ?? 0;
    return [(int) $n, (int) $m, (int) ($undefined ?? 0), (unset) $n, (string) strstr('a', 'b')];
}

class Box {
    /** @var int */
    protected $size;
    public function sizes() {
        return [(int) $this->size, (int) $this->legacy()];
    }
    /** @return int */
    public function legacy() { return 1; }
}

function rebound(array $lines, array $refs) {
    foreach ($lines as $v) {
        $v = trim($v);
    }
    foreach ($refs as $v) {
        $id = (string) $v;
    }
}

class Integer {}
class Boolean {}

function aliasClasses(Integer $i, Boolean $b) {
    return [(int) $i, (bool) $b];
}

function coalesced(?int $y) {
    $x = $y ?? 5;
    return (int) $x;
}

function reassignedParam($v) {
    $v = trim($v);
    return (string) $v;
}
