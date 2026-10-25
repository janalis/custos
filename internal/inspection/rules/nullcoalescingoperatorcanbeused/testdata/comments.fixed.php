<?php
function lookup(array $in, $k)
{
    // The fix would drop the comments: none is offered.
    if (isset($in['a'])) {
        // This is just for performance.
        return $in['a'];
    }
    return 'none';
}

function kept(array $in)
{
    // A comment outside the replaced statement stays.
    return $in['b'] ?? 0;
}

function inline(array $in)
{
    return $in[/* key */ 'c'] ?? 1;
}
