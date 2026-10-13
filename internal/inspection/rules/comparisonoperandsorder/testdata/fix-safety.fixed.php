<?php
// The moved operand must not fuse with a word written against the comparison.
function glued($a, $b)
{
    if ($a[0]==null and$b) {}
    if ($a==1) {}
    return \strtolower($b)=='a';
}
function gluedName($b) { return trim($b)=='a'; }

// An interpolated string reads its variables when evaluated: report only
// when the other operand (or the string itself) has side effects.
function ordered($i, $j, $o)
{
    $a = "v{$i}" == 'v' . $i++;
    $b = "{$o->next()}" == $o->pos;
    $c = $j == "v{$i}";
    return $j === <<<EOT
    v
    EOT;
}
