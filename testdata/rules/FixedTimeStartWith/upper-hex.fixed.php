<?php
function upperHex($line)
{
    return [
        strncmp($line, "\X09", 1) === 0,
        strncmp($line, "\X41-", 2) === 0,
        0 !== strncasecmp($line, "\X4a\x4B", 2),
        strncmp($line, "\\X41", 4) === 0,
        strncmp($line, "\Xzz", 4) === 0,
        strncmp($line, '\X41', 4) === 0,
    ];
}
