<?php
function guarded(array $opts, array $map, $key, $flag, $m, $b) {
    $a = isset($opts['on']) && !($key instanceof stdClass) && isset($map[$key]);
    $c = isset($opts['on']) && $flag && isset($map[pick()]);
    $d = isset($opts['on']) && $flag && isset($m->prop);
    $e = !isset($opts['on']) || $flag || !isset($map[pick()]);
    $f = isset($opts['on']) && $flag && isset($b, $map[$key]);
    $g = isset($opts['on']) && $flag && isset($$key);
    $h = isset($opts['on'], $b, $key) && $flag;
    $i = $flag && isset($opts['on'], $map[pick()]);
    return [$a, $c, $d, $e, $f, $g, $h, $i];
}
