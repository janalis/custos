<?php
function prefixes($url) {
    return [
        strpos($url, '') === 0,
        0 !== stripos($url, ""),
        <warning descr="Use 'strncmp($url, '/', 1)' for a length-independent prefix check.">strpos($url, '/')</warning> === 0,
    ];
}
