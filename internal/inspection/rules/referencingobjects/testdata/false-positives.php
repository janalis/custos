<?php
class Node
{
    public $next;
}

// The reference lets the callee replace the caller's variable.
function swapNodes(Node &$a, Node &$b) { list($a, $b) = array($b, $a); }
function advance(Node &$node) { [$node, $step] = [$node->next, 1]; }
function nested(Node &$node, array $rows) { [[$node]] = $rows; }
function keyed(Node &$node, array $row) { ['n' => $node] = $row; }
function walk(Node &$node) { foreach ([$node->next] as $node) {} }
function walkKey(Node &$node, array $map) { foreach ($map as $node => $ignored) {} }
function walkPattern(Node &$node, array $rows) { foreach ($rows as [$node, $x]) {} }
function alias(Node &$a) { $r = &$a; $r = new Node(); }
function boxed(Node &$a) { $box = [&$a]; $box[0] = new Node(); }
function boxedLong(Node &$a) { $box = array(&$a); $box[0] = new Node(); }
function caught(Node &$a) { try { run(); } catch (Exception $a) {} }
function bump(?Node &$a) { $a++; }
function later(Node &$a) { return function () use (&$a) { reset_it(); }; }
function &pick(Node &$a) { return $a; }
function many(Node &...$rest) { $rest[0] = new Node(); }
function manyCopy(Node &...$rest) { return $rest; }
// A null turns into an array in the caller's variable.
function vivify(?Node &$a) { $a[] = 1; }
function vivifyDeep(Node|null &$a) { $a['k']['j'] = 1; }
// Local variables written by name.
function spread(Node &$a, array $vars) { extract($vars); }
function byName(Node &$a, $n) { $$n = new Node(); }
function template(Node &$a, $file) { include $file; }
function script(Node &$a, $code) { eval($code); }

final class Holder
{
    public function __construct(public Node &$node, private readonly Node &$other) {}

    public function &pick(Node &$a) { return $a; }
}
