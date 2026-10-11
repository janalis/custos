<?php
class Node
{
    public $next;
    public $id;
}

// None of these uses needs the reference.
function patternKey(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>, array $row) { [$node->id => $x] = $row; return $x; }
function plainKey(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>, array $row) { [$node => $x] = $row; return $x; }
function literal(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { return [$node, 'k' => $node]; }
function source(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { foreach ($node as $k => $v) {} }
function copied(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { $copy = $node; return $copy; }
function imported(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { return function () use ($node) { return 1; }; }
function returned(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { return $node; }
function rebound(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { unset($node); }
function unusedRest(<warning descr="Objects are handed over by handle already; drop the '&' before '$rest'.">Node &...$rest</warning>) { return 1; }
function element(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node &$node</warning>) { $node['k'] = 1; }
function elementUnion(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">Node|Other &$node</warning>) { return $node['k']; }
function boxedElement(<warning descr="Objects are handed over by handle already; drop the '&' before '$node'.">?Node &$node</warning>) { return [$node][0]; }
