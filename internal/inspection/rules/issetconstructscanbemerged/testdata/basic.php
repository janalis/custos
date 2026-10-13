<?php
function probe(array $cfg, $row, $extra) {
    $r1 = isset($cfg['host']) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['port'])</weak_warning>;
    $r2 = isset($row->id, $row->name) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($extra)</weak_warning>;
    $r3 = $extra > 2 && isset($cfg['a']) && (<weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['b'])</weak_warning> && isset($cfg['c']));
    $r4 = isset($cfg['x']) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($cfg['y'])</weak_warning> && ($extra || $row);
    $r5 = !isset($row->id) || !<weak_warning descr="Merge this check into the preceding !isset() call.">isset($row->name)</weak_warning> || $extra;

    $n1 = isset($cfg['host']) && !isset($cfg['port']);
    $n2 = isset($cfg['host']) || isset($cfg['port']);
    $n3 = !isset($cfg['host']) && !isset($cfg['port']);
    $n4 = isset($cfg['host']) and isset($cfg['port']);
    return [$r1, $r2, $r3, $r4, $r5, $n1, $n2, $n3, $n4];
}
