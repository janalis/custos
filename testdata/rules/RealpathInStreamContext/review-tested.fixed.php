<?php
// The failure check on the result must survive: these keep realpath().
function viaEmpty() {
    $p = realpath(__DIR__ . '/../a');
    if (empty($p)) { return 'no dir'; }
    return 'dir';
}
function viaTypeTest() {
    $p = realpath(__DIR__ . '/../b');
    if (is_bool($p)) { return 'no dir'; }
    return is_string($q = realpath(__DIR__ . '/../c')) ? 'dir' : 'no dir';
}
function viaTry() {
    try {
        $p = realpath(__DIR__ . '/../d');
    } finally {
    }
    if ($p === false) { return 'no dir'; }
    return 'dir';
}
function viaEarlierTest($items) {
    $p = false;
    foreach ($items as $item) {
        if ($p) { return $p; }
        $p = realpath(__DIR__ . '/../e');
    }
    return null;
}
function viaSwitch($d) {
    switch (realpath(__DIR__ . '/../f')) {
        case false: return 'no dir';
    }
    switch (true) {
        case realpath(__DIR__ . '/../g'): return 'dir';
    }
    $m = match (realpath(__DIR__ . '/../h')) { false => 'no dir', default => 'dir' };
    return match (true) { realpath(__DIR__ . '/../i') => 'dir', default => $m };
}
function viaLoops($d) {
    do {
        $d--;
    } while (realpath(__DIR__ . '/../j'));
    for (; realpath(__DIR__ . '/../k'); ) {
        break;
    }
    return realpath(__DIR__ . '/../l') xor $d;
}
class Holder
{
    public $root;

    public function boot() {
        $this->root = realpath(__DIR__ . '/../m');
    }

    public function ok() {
        return $this->root !== false;
    }
}
$top = realpath(__DIR__ . '/../n');
function unrelated() {
}
if (!$top) {
    exit(1);
}

// Not a failure test: these are rewritten.
function untested($d, $o) {
    $p = dirname(__DIR__) . '/o';
    $n = basename(dirname(__DIR__) . '/p');
    $o->open(dirname(__DIR__) . '/q');
    $m = match (true) { default => dirname(__DIR__) . '/r' };
    for (dirname(__DIR__) . '/s'; $d; ) {
        break;
    }
    $a = dirname(__DIR__) . '/t';
    $b = $a;
    $a = $b;
    $top = dirname(__DIR__) . '/u';
    return [$p, $n, $m, $a, $top];
}
$kind = gettype(realpath(__DIR__ . '/../v'));
