<?php
$cache = & <warning descr="Objects are handed over by handle already; assign the new instance without '&'.">new ArrayObject([])</warning>;
$queue =&<warning descr="Objects are handed over by handle already; assign the new instance without '&'.">new SplQueue</warning>;
$plain = new SplStack();

function attach(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">DOMNode &$node</warning>, array &$list, &$any, int &$n) {}
function detach(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">?DOMNode& $node</warning>, ?string &$label, iterable &$it) {}
function rest(<warning descr="Objects are handed over by handle already; drop the '&' before '$nodes'.">DOMNode &...$nodes</warning>) {}

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
