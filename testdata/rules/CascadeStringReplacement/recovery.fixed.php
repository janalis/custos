<?php
// Empty array elements and the octal-looking key 08 do not compile in PHP;
// the parser accepts them. Empty elements are dropped from a merge; the
// unreadable key leaves the merge without a fix.
function broken($s, $t, $u)
{
    $s = str_replace(array('p', 'a', 'b'), array('q', 'y', 'y'), $s);

    $t = str_replace(['a', , 'a'], 'y', $t);

    $u = str_replace('p', 'q', $u);
    $u = str_replace([08 => 'a', 'b'], 'y', $u);
    return $s . $t . $u;
}

function afterBareReturn($s)
{
    return;
    $s = str_replace('a', 'b', $s);
}
