<?php
// Reading the watch array, also through builtins, restores nothing.
function reads($a, $b, $running, $other) {
    $read = [$a, $b]; $write = null; $except = null;
    while ($running) {
        [$other, $write] = [1, null];
        foreach ($read as $k => $s) { fread($s, 10); }
        <warning descr="Restore stream watch arrays before each selection.">stream_select($read, $write, $except, 1)</warning>;
        echo count($read);
    }
}
