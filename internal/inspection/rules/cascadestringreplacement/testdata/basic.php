<?php
function slug($title, $raw, $line) {
    $title = str_replace('&', 'and', $title);
    /** spaces */
    $title = <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(' ', '-', $title)</warning>;
    return <warning descr="Fold this str_replace() into the preceding one on the same variable.">str_replace(['.', ','], '', $title)</warning>;
}

function others($raw, $line) {
    $code = str_replace('_', '-', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace(' ', '_', $raw)</warning>);
    $mark = str_replace(['k'], ['!'], <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace('j', '!', $raw)</warning>);
    $clean = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">array("\t", "\t")</weak_warning>, ' ', $line);
    $exp = \str_replace(['p', 'q'], 'z', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace(['r'], 'w', $raw)</warning>);

    $combo = str_replace(<weak_warning descr="All searched items are identical; pass the single string.">['k', 'k']</weak_warning>, '!', <warning descr="Fold this nested str_replace() into the enclosing call.">str_replace('j', '!', $raw)</warning>);

    $x = str_replace('a', 'b', $raw);
    $y = str_replace('c', 'd', $x);
    $keep = str_replace(['a', 'b'], 'c', $line);
    log_it(str_replace('q', 'r', str_replace('s', 't', $line)));
    return $y;
}
