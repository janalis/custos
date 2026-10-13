<?php
class Node
{
    public $next;
    public $id;
}

// None of these uses needs the reference.
function patternKey(Node $node, array $row) { [$node->id => $x] = $row; return $x; }
function plainKey(Node $node, array $row) { [$node => $x] = $row; return $x; }
function literal(Node $node) { return [$node, 'k' => $node]; }
function source(Node $node) { foreach ($node as $k => $v) {} }
function copied(Node $node) { $copy = $node; return $copy; }
function imported(Node $node) { return function () use ($node) { return 1; }; }
function returned(Node $node) { return $node; }
function rebound(Node $node) { unset($node); }
function unusedRest(Node ...$rest) { return 1; }
function element(Node $node) { $node['k'] = 1; }
function elementUnion(Node|Other $node) { return $node['k']; }
function boxedElement(?Node $node) { return [$node][0]; }
