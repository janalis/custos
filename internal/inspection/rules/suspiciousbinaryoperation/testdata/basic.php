<?php
trait Loggable {}
interface Shape {}

function checks($obj, $p, $q, array $tags, ?int $limit) {
    $t1 = <error descr="A trait is never an instanceof target; this is always false.">$obj instanceof Loggable</error>;
    $t2 = $obj instanceof Shape;
    $s1 = <error descr="Both operands are the same.">$p->id === ($p->id)</error>;
    $s2 = <error descr="Both operands are the same.">count($tags) > count($tags)</error>;
    $p <error descr="Comparison result is discarded; did you mean '='?">==</error> $q;
    $map = ['enabled' <error descr="Did you mean '=>' for an array key?">>=</error> true, $p >= 'x'];
    $out = [
        <error descr="'(bool)$p' is never null, so '??' is useless; add parentheses.">(bool)$p</error> ?? 0,
        (<error descr="'!$q' is never null, so '??' is useless; add parentheses.">!$q</error>) ?? 1,
        $p ?? $q,
    ];
    $msg = 'tags: ' <error descr="Concatenating an array literal makes no sense.">.</error> ['a'];
    $ok = [
        $p and <error descr="This constant decides the whole condition.">FALSE</error>,
        $p || <error descr="This constant decides the whole condition.">true</error>,
        $p && <error descr="This constant has no effect in the condition.">\true</error>,
        $p or <error descr="This constant has no effect in the condition.">null</error>,
    ];
    if (<error descr="Null or false operands make this negated comparison misleading; use '$limit >= 10'.">!($limit < 10)</error>) {}
}

class Quota {
    public function fits(string $who): bool {
        if (strlen($who <error descr="This comparison probably belongs outside the call parentheses.">>=</error> 3)) {}
        if ($this->allow($who === 'root')) {}
        return true;
    }
    private function allow(bool $flag) { return $flag; }
}

if (<error descr="Operator precedence is unclear here; add parentheses.">$found = $left != $right</error>) {}
if ($x || <error descr="Operator precedence is unclear here; add parentheses.">$y && $z</error>) {}
if ($row = <error descr="Operator precedence is unclear here; add parentheses.">fetch() || $fallback</error>) {}
if (<error descr="Operator precedence is unclear here; add parentheses.">!  $x >= $limit</error>) {}
if (($x && $y) || $z) {}
$both = $x && $y;
if ((!$x) < $y) {}
