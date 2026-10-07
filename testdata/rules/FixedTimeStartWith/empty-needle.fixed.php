<?php
function prefixes($url) {
    return [
        strpos($url, '') === 0,
        0 !== stripos($url, ""),
        strncmp($url, '/', 1) === 0,
    ];
}
