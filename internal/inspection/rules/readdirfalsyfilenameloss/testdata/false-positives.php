<?php
function review_false_positives($h, $names, $cached, $other) {
    // Strict comparisons already keep an entry named "0".
    while (!(($e = readdir($h)) === false)) { echo $e; }
    while (!(false === ($e = readdir($h)))) { echo $e; }
    if (!(readdir($h) === false)) { return 1; }
    if ((readdir($h) !== false) == true) { return 2; }
    if (readdir($h) === true) { return 3; }
    // The tested value is not the directory entry.
    if ($names .= readdir($h)) { return 4; }
    if ($cached ??= readdir($h)) { return 5; }
    if ($flag = !readdir($h)) { return 6; }
    if ($list[readdir($h)] = $other) { return 7; }
    if (readdir(...)) { return 8; }
    if (-readdir($h)) { return 9; }
    if (readdir($h) == null) { return 10; }
    if (readdir($h) == $other) { return 11; }
    if (readdir($h) . 'x') { return 12; }
    if (strlen(readdir($h))) { return 13; }
    // Only the last expression of a for condition is tested.
    for (readdir($h); readdir($h), $other; readdir($h)) { echo 1; }
    $entry = readdir($h);
    return $entry;
}
