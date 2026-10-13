<?php
// The failure check on the result must survive: these keep realpath().
function viaEmpty() {
    $p = <warning descr="Use 'dirname(__DIR__) . '/a'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../a')</warning>;
    if (empty($p)) { return 'no dir'; }
    return 'dir';
}
function viaTypeTest() {
    $p = <warning descr="Use 'dirname(__DIR__) . '/b'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../b')</warning>;
    if (is_bool($p)) { return 'no dir'; }
    return is_string($q = <warning descr="Use 'dirname(__DIR__) . '/c'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../c')</warning>) ? 'dir' : 'no dir';
}
function viaTry() {
    try {
        $p = <warning descr="Use 'dirname(__DIR__) . '/d'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../d')</warning>;
    } finally {
    }
    if ($p === false) { return 'no dir'; }
    return 'dir';
}
function viaEarlierTest($items) {
    $p = false;
    foreach ($items as $item) {
        if ($p) { return $p; }
        $p = <warning descr="Use 'dirname(__DIR__) . '/e'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../e')</warning>;
    }
    return null;
}
function viaSwitch($d) {
    switch (<warning descr="Use 'dirname(__DIR__) . '/f'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../f')</warning>) {
        case false: return 'no dir';
    }
    switch (true) {
        case <warning descr="Use 'dirname(__DIR__) . '/g'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../g')</warning>: return 'dir';
    }
    $m = match (<warning descr="Use 'dirname(__DIR__) . '/h'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../h')</warning>) { false => 'no dir', default => 'dir' };
    return match (true) { <warning descr="Use 'dirname(__DIR__) . '/i'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../i')</warning> => 'dir', default => $m };
}
function viaLoops($d) {
    do {
        $d--;
    } while (<warning descr="Use 'dirname(__DIR__) . '/j'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../j')</warning>);
    for (; <warning descr="Use 'dirname(__DIR__) . '/k'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../k')</warning>; ) {
        break;
    }
    return <warning descr="Use 'dirname(__DIR__) . '/l'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../l')</warning> xor $d;
}
class Holder
{
    public $root;

    public function boot() {
        $this->root = <warning descr="Use 'dirname(__DIR__) . '/m'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../m')</warning>;
    }

    public function ok() {
        return $this->root !== false;
    }
}
$top = <warning descr="Use 'dirname(__DIR__) . '/n'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../n')</warning>;
function unrelated() {
}
if (!$top) {
    exit(1);
}

// Not a failure test: these are rewritten.
function untested($d, $o) {
    $p = <warning descr="Use 'dirname(__DIR__) . '/o'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../o')</warning>;
    $n = basename(<warning descr="Use 'dirname(__DIR__) . '/p'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../p')</warning>);
    $o->open(<warning descr="Use 'dirname(__DIR__) . '/q'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../q')</warning>);
    $m = match (true) { default => <warning descr="Use 'dirname(__DIR__) . '/r'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../r')</warning> };
    for (<warning descr="Use 'dirname(__DIR__) . '/s'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../s')</warning>; $d; ) {
        break;
    }
    $a = <warning descr="Use 'dirname(__DIR__) . '/t'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../t')</warning>;
    $b = $a;
    $a = $b;
    $top = <warning descr="Use 'dirname(__DIR__) . '/u'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../u')</warning>;
    return [$p, $n, $m, $a, $top];
}
$kind = gettype(<warning descr="Use 'dirname(__DIR__) . '/v'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../v')</warning>);
