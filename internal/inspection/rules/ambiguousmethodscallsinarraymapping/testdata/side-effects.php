<?php
function fill(array $handles, array $cursor, array $seen) {
    foreach ($handles as $h) {
        $lines[fgets($h)] = fgets($h);
        $pairs[next($cursor)] = next($cursor);
        $rand[mt_rand(1, 9)] = MT_RAND(1, 9);
        $ids[uniqid()] = uniqid();
        $sorted[sort($seen)] = sort($seen);
        $wrap[strtoupper(array_shift($cursor))] = strtoupper(array_shift($cursor));
        $inc[md5($i++)] = md5($i++);

        $trimmed[trim($h)] = <warning descr="This call is repeated in the key; store its result in a local variable.">trim($h)</warning>;
        $keys[key($cursor)] = <warning descr="This call is repeated in the key; store its result in a local variable.">key($cursor)</warning>;
        $got[$h->next()] = <warning descr="This call is repeated in the key; store its result in a local variable.">$h->next()</warning>;
    }
}
