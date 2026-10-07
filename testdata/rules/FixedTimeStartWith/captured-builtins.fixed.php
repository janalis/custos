<?php
namespace Urls {
    function strncmp($a, $b, $n) { return 0; }
    use function Other\strncasecmp;

    $a = \strncmp($url, 'www.', 4) === 0;
    $b = \strncasecmp($url, 'http', 4) !== 0;
}
