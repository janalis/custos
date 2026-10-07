<?php
function gate($ready, $cached, $forced, $dry) {
    if (($ready and $cached) or $forced) {}
    if ($ready or ($cached and $forced)) {}
    if (($ready xor $dry) or $forced) {}

    if (($ready and $cached) or $forced) {}
    if ($ready and $cached and $forced) {}
    if ($ready && $cached or $forced) {}
    if ($ready or $cached || $forced) {}
    $handle = fopen('php://memory', 'r') or die('no');
    return $handle;
}
