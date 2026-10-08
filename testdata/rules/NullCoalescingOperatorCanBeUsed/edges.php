<?php
class Node {
    /** @var int */
    public $id;
    /** @var string|null */
    public $label;
    /** @var int */
    public static $count;

    public function self(): static { return $this; }

    public function size() {
        $s = $this->self();
        return <weak_warning descr="Simplify to '$s->id ?? 0' using the null coalescing operator.">$s ? $s->id : 0</weak_warning>;
    }
}
/** @var Node|null $maybe */
/** @var Node $node */
/** @var object $obj */
/** @var int|Node $mixed */
/** @var string $cls */

// Form A: conditions and candidates that do not match
$a1 = $x == null ? $x : 1;                       // loose comparison
$a2 = probe($x) ? $x : 1;                        // other call
$a3 = isset($x) ? $y : 1;                        // isset candidate mismatch
$a4 = !empty($x) ? $x : null;                    // empty: candidate not a property
$a5 = !empty($maybe) ? $other->id : null;        // empty: base mismatch
$a6 = $node ? $node : 1;                         // truthy: candidate not a property
$a7 = array_key_exists($k) ? $map[$k] : null;    // argument count
$a8 = $x !== null ? $y : 1;                      // null comparison: mismatch
$a9 = !empty($maybe) ? Node::$count : 0;         // static candidate, nullable probe
$a10 = $node ? $node->label : 'none';            // nullable candidate, real fallback
$a11 = $mixed ? $mixed->id : 0;                  // probe may be a scalar
$a12 = $x === $y ? $x : 1;                       // no null operand
$a13 = true ? $x : 1;                            // unsupported condition
$a14 = !empty($unknown) ? $unknown::$count : null; // static candidate, unknown probe

// Form A: reported
$b1 = <weak_warning descr="Simplify to '$x ?? ($y instanceof Node)' using the null coalescing operator.">isset($x) ? $x : $y instanceof Node</weak_warning>;
$b3 = <weak_warning descr="Simplify to '$node::$count ?? null' using the null coalescing operator.">!empty($node) ? $node::$count : null</weak_warning>;
// '' or null as class throws inside ??, while !empty() skipped the lookup
$b4 = !empty($cls) ? $cls::$count : null;
$b5 = !empty($maybe) ? $maybe::$count : null;

// Form B: shapes that do not match
function noIfs($in, $node) {
    if (isset($in)) { $v = $in; } elseif ($node) { $v = 1; }
    if (isset($in)): return $in; endif;
    if (probe($in)) { return $in; }
    if (isset($in)) return $in;
    if (isset($in)) { echo $in; }
    if (isset($in)) { return $in; } else { echo 1; return 2; }
    if (isset($in)) { return $in; } else { $v = 2; }
    if (isset($in)) { $v .= $in; } else { $v = 2; }
    if (isset($in)) { $in['k'] = $in; } else { $v = 2; }
    $w = &$node;
    if (isset($in)) { $w = $in; }
    $w = $v = 1;
    if (isset($in)) { $w = $in; }
    if (isset($in)) { return; }
    if (isset($in)) { return $node; }
    if (isset($in)) { return $in; }
    echo 'next';
    while (true) {
        if (isset($in)) { return $in; }
    }
}
function mismatch($in, $node) {
    if (isset($in)) { return $node; }
    return 1;
}
function bareFirst($in) {
    if (isset($in)) { return; }
    return 1;
}
function firstAssign($in) {
    if (isset($in)) { $w = $in; }
}
function emptyAtEnd($o) {
    if (empty($o)) { return $o->id; }
}
function falsyAtEnd($o) {
    if (!$o) { return $o->id; }
}
function akeAtEnd($m, $k) {
    if (array_key_exists($k, $m)) { return $m[$k]; }
}
function nullAtEnd($x) {
    if ($x !== null) { return $x; }
}
// The probe reads the target the previous statement assigned.
function mimeOf(array $parts, array $mimeMap) {
    $type = $parts[1];
    if (isset($mimeMap[$type])) {
        $type = $mimeMap[$type];
    }
    $key = $parts[0];
    if (isset($mimeMap['x'])) {
        $key = $mimeMap['x'] . $key;
    }
    return [$type, $key];
}

if (isset($top)) { return $top; }
