<?php
namespace App {
    function strpos($h, $n) { return 0; }

    $a = strpos($url, 'https') === 0;
}

namespace {
    $b = strncmp($url, 'https', 5) === 0;
    $c = \strncasecmp($url, 'ftp', 3) !== 0;
}
