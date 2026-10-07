<?php
function prefixes($url, $head, $part)
{
    return [
        strncmp($url, 'https', 5) === 0,
        0 !== strncasecmp($head, 'Accept:', 7),
        \strncmp($url, 'C:\\', 3) !== 0,
        0 === strncmp($url, 'it\'s', 4),
        0 === strncasecmp($url, "x\ty", 3),
        strncmp($url, 'é/', 3) === 0,
    ];
}
