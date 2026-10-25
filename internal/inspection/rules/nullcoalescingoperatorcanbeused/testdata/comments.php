<?php
function lookup(array $in, $k)
{
    // The fix would drop the comments: none is offered.
    <weak_warning descr="Simplify to 'return $in['a'] ?? 'none'' using the null coalescing operator.">if</weak_warning> (isset($in['a'])) {
        // This is just for performance.
        return $in['a'];
    }
    return 'none';
}

function kept(array $in)
{
    // A comment outside the replaced statement stays.
    <weak_warning descr="Simplify to 'return $in['b'] ?? 0' using the null coalescing operator.">if</weak_warning> (isset($in['b'])) {
        return $in['b'];
    }
    return 0;
}

function inline(array $in)
{
    return <weak_warning descr="Simplify to '$in[/* key */ 'c'] ?? 1' using the null coalescing operator.">isset($in[/* key */ 'c']) ? $in[/* key */ 'c'] : 1</weak_warning>;
}
