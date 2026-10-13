<?php
// preg_quote()'s argument is literal text; a mandatory inner group is not
// an optional one.
function quoted(string $t)
{
    return [
        preg_quote('/test/invalidpath/testing', '/'),
        preg_match('/^[a-z]+(\s+(?:unsigned|zerofill))*\z/i', $t),
        preg_match('/^(\s+(?:x)?)*$/', $t),
        preg_match(<error descr="Nested quantifier (\d+)* risks catastrophic backtracking.">'/^a(\d+)*$/'</error>, $t),
    ];
}
