<?php
// The moved operand must not fuse with a word written against the comparison.
function glued($a, $b)
{
    if (<weak_warning descr="Put the constant operand on the right side of the comparison.">null==$a[0]</weak_warning>and$b) {}
    if (<weak_warning descr="Put the constant operand on the right side of the comparison.">1==$a</weak_warning>) {}
    return<weak_warning descr="Put the constant operand on the right side of the comparison.">'a'==\strtolower($b)</weak_warning>;
}
function gluedName($b) { return<weak_warning descr="Put the constant operand on the right side of the comparison.">'a'==trim($b)</weak_warning>; }

// An interpolated string reads its variables when evaluated: report only
// when the other operand (or the string itself) has side effects.
function ordered($i, $j, $o)
{
    $a = <weak_warning descr="Put the constant operand on the right side of the comparison.">"v{$i}" == 'v' . $i++</weak_warning>;
    $b = <weak_warning descr="Put the constant operand on the right side of the comparison.">"{$o->next()}" == $o->pos</weak_warning>;
    $c = <weak_warning descr="Put the constant operand on the right side of the comparison.">"v{$i}" == $j</weak_warning>;
    return <weak_warning descr="Put the constant operand on the right side of the comparison."><<<EOT
    v
    EOT === $j</weak_warning>;
}
