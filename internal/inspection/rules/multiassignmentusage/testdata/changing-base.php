<?php
function unpack_rows(array $row, $repo, $o, array $cfg) {
    $row = $row[0];
    $second = $row[1];

    $head = $repo->load()[0];
    $tail = $repo->load()[1];

    $a = fetch()[0];
    $b = fetch()[1];

    $o = $o->list[0];
    $p = $o->list[1];

    $cfg[1] = $cfg[0];
    $x = $cfg[2];

    $i = $list[$n++][0];
    $j = $list[$n++][1];

    $this->first = $this->items[0];
    <weak_warning descr="Use one destructuring assignment from '$this->items' instead.">$next = $this->items[1]</weak_warning>;

    $out = $cfg[0];
    <weak_warning descr="Use one destructuring assignment from '$cfg' instead.">$out2 = $cfg[1]</weak_warning>;
    return [$second, $head, $tail, $a, $b, $p, $x, $i, $j, $next, $out2];
}
