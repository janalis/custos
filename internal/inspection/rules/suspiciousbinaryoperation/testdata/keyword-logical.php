<?php
function gate($ready, $cached, $forced, $dry) {
    if (<error descr="Operator precedence is unclear here; add parentheses.">$ready and $cached</error> or $forced) {}
    if ($ready or <error descr="Operator precedence is unclear here; add parentheses.">$cached and $forced</error>) {}
    if (<error descr="Operator precedence is unclear here; add parentheses.">$ready xor $dry</error> or $forced) {}

    if (($ready and $cached) or $forced) {}
    if ($ready and $cached and $forced) {}
    if ($ready && $cached or $forced) {}
    if ($ready or $cached || $forced) {}
    $handle = fopen('php://memory', 'r') or die('no');
    return $handle;
}
