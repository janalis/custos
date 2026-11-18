<?php
$fp = fopen($path, "w+"); if ($fp === false) { return; } if (fseek($fp, 10) !== 0) { return; } if (ftruncate($fp, 0)) { rewind($fp); fwrite($fp, "new"); }

function appendOnly($path) {
    $log = fopen($path, "a+");
    if ($log === false) { return; }
    if (fseek($log, 10) !== 0) { return; }
    if (ftruncate($log, 0)) { fwrite($log, "new"); }
}

function inMemory() {
    $buffer = fopen("php://memory", "w+");
    if ($buffer === false) { return; }
    if (fseek($buffer, 10) !== 0) { return; }
    if (ftruncate($buffer, 0)) { fwrite($buffer, "new"); }
}
