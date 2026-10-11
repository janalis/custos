<?php
function review_conditions($h, $skip) {
    if ((readdir($h)) !== false) { return 1; }
    if ((readdir($h)) === false) { return 2; }
    if ((readdir($h)) !== false) { return 3; }
    if ((readdir($h)) !== false) { return 4; }
    if ((readdir($h)) === false) { return 5; }
    if ((readdir($h)) === false) { return 6; }
    if ($skip) { return 7; } elseif (($e = readdir($h)) !== false) { return $e; }
    if ($skip): elseif ((readdir($h)) === false): return 8; endif;
    for (; ($e = readdir($h)) !== false; ) { echo $e; }
    for ($i = 0; $i++, (readdir($h)) !== false; ) { echo $i; }
    do { echo 1; } while (($e = \readdir($h)) !== false);
    return 0;
}
