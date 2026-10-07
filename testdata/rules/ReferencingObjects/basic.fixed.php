<?php
$cache = new ArrayObject([]);
$queue = new SplQueue;
$plain = new SplStack();

function attach(DOMNode $node, array &$list, &$any, int &$n) {}
function detach(?DOMNode $node, ?string &$label, iterable &$it) {}
function rest(DOMNode ...$nodes) {}

interface Visitor {
    public function visit(
        Tree & $tree,
        mixed &$extra
    );
}

function replace(DOMNode &$node) {
    $node = new DOMText('x');
}
function check(DOMNode &$a, DOMNode &$b, DOMNode &$c, DOMNode &$d = null) {
    if (($a)) {}
    return !$b || ($c ? 1 : 0);
}
function scalars(\string &$s, int|null &$i, ?\string &$t, Foo|string &$u) {}
$fn = function (DOMNode &$node) {};
