<?php
function slug($title)
{
    $title = str_replace('&', /* why */ 'and', $title);
    $title = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(' ', '-', $title)</warning>;
    $nested = str_replace('a', 'b', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace('c', /* keep */ 'd', $title)</warning>);
    $same = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">['x', /* one */ 'x']</weak_warning>, 'y', $title);
    return [$title, $nested, $same];
}
