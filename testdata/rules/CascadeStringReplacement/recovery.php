<?php
// Empty array elements and the octal-looking key 08 do not compile in PHP;
// the parser accepts them. Empty elements are dropped from a merge; the
// unreadable key leaves the merge without a fix.
function broken($s, $t, $u)
{
    $s = str_replace('p', 'q', $s);
    $s = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['a', , 'b'], 'y', $s)</warning>;

    $t = str_replace(['a', , 'a'], 'y', $t);

    $u = str_replace('p', 'q', $u);
    $u = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace([08 => 'a', 'b'], 'y', $u)</warning>;
    return $s . $t . $u;
}

function afterBareReturn($s)
{
    return;
    $s = str_replace('a', 'b', $s);
}
