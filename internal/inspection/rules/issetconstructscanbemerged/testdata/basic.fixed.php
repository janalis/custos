<?php
function probe(array $cfg, $row, $extra) {
    $r1 = isset($cfg['host'], $cfg['port']);
    $r2 = isset($row->id, $row->name, $extra);
    $r3 = $extra > 2 && isset($cfg['a'], $cfg['b']) && isset($cfg['c']);
    $r4 = isset($cfg['x'], $cfg['y']) && ($extra || $row);
    $r5 = !isset($row->id, $row->name) || $extra;

    $n1 = isset($cfg['host']) && !isset($cfg['port']);
    $n2 = isset($cfg['host']) || isset($cfg['port']);
    $n3 = !isset($cfg['host']) && !isset($cfg['port']);
    $n4 = isset($cfg['host']) and isset($cfg['port']);
    return [$r1, $r2, $r3, $r4, $r5, $n1, $n2, $n3, $n4];
}
