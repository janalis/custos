<?php
function adjust(array $prices, array $tags) {
    foreach ($prices as $sku => $amount) {
        <weak_warning descr="Iterate '$amount' by reference and assign to it directly instead of writing through the key.">$prices[$sku]</weak_warning> = $amount * 2;
    }
    unset($sku, <weak_warning descr="'$amount' is not a reference here; unsetting it is unnecessary.">$amount</weak_warning>);

    foreach ($tags as &<warning descr="Unset '$tag' right after the loop: it is still a reference to the last element.">$tag</warning>) {
        foreach ($prices as $p) {
            echo $tag, $p;
        }
    }
    unset(<weak_warning descr="'$p' is not a reference here; unsetting it is unnecessary.">$p</weak_warning>);

    foreach ($tags as &$t) { $t = trim($t); }
    /** trimmed */
    unset($t);

    if ($tags) {
        foreach ($tags as & $u) { $u .= '!'; }
    }
    return $tags;
}

function tail(array $list) {
    foreach ($list as &$entry) { $entry++; }
}

function writes(array $rows, $other) {
    foreach ($rows as $i => $row) {
        foreach ($other as $j => $x) {
            <weak_warning descr="Iterate '$row' by reference and assign to it directly instead of writing through the key.">$rows[$i]</weak_warning> .= $x;
        }
        $rows[] = 1;
        $rows['x'] = 1;
        $rows[$i + 1] = 1;
        $other[$i] = 1;
        $fn = function () use ($rows, $i) { $rows[$i] = 2; };
    }
    foreach ($rows as $i => &$row) {
        $rows[$i] = 0;
    }
    unset($row);
    foreach ($rows as $m => $cell) {
        <weak_warning descr="Iterate '$cell' by reference and assign to it directly instead of writing through the key.">$rows[$m]</weak_warning> += 1;
    }
}
