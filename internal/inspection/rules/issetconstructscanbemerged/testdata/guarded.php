<?php
function guarded(array $opts, array $map, $key, $flag, $m, $b) {
    $a = isset($opts['on']) && !($key instanceof stdClass) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($map[$key])</weak_warning>;
    $c = isset($opts['on']) && $flag && <weak_warning descr="Merge this check into the preceding isset() call.">isset($map[pick()])</weak_warning>;
    $d = isset($opts['on']) && $flag && <weak_warning descr="Merge this check into the preceding isset() call.">isset($m->prop)</weak_warning>;
    $e = !isset($opts['on']) || $flag || !<weak_warning descr="Merge this check into the preceding !isset() call.">isset($map[pick()])</weak_warning>;
    $f = isset($opts['on']) && $flag && <weak_warning descr="Merge this check into the preceding isset() call.">isset($b, $map[$key])</weak_warning>;
    $g = isset($opts['on']) && $flag && <weak_warning descr="Merge this check into the preceding isset() call.">isset($$key)</weak_warning>;
    $h = isset($opts['on']) && $flag && <weak_warning descr="Merge this check into the preceding isset() call.">isset($b, $key)</weak_warning>;
    $i = $flag && isset($opts['on']) && <weak_warning descr="Merge this check into the preceding isset() call.">isset($map[pick()])</weak_warning>;
    return [$a, $c, $d, $e, $f, $g, $h, $i];
}
