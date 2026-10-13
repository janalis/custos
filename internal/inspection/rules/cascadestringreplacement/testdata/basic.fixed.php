<?php
function slug($title, $raw, $line) {
    /** spaces */
    return str_replace(['&', ' ', '.', ','], ['and', '-', '', ''], $title);
}

function others($raw, $line) {
    $code = str_replace([' ', '_'], ['_', '-'], $raw);
    $mark = str_replace(['j', 'k'], '!', $raw);
    $clean = str_replace("\t", ' ', $line);
    $exp = \str_replace(['r', 'p', 'q'], ['w', 'z', 'z'], $raw);

    $combo = str_replace(['j', 'k'], '!', $raw);

    $x = str_replace('a', 'b', $raw);
    $y = str_replace('c', 'd', $x);
    $keep = str_replace(['a', 'b'], 'c', $line);
    log_it(str_replace('q', 'r', str_replace('s', 't', $line)));
    return $y;
}
