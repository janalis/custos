<?php
function prefixes($url, $head, $part)
{
    return [
        <warning descr="Use 'strncmp($url, 'https', 5)' for a length-independent prefix check.">strpos($url, 'https')</warning> === 0,
        0 !== <warning descr="Use 'strncasecmp($head, 'Accept:', 7)' for a length-independent prefix check.">stripos($head, 'Accept:')</warning>,
        <warning descr="Use '\strncmp($url, 'C:\\', 3)' for a length-independent prefix check.">\strpos($url, 'C:\\')</warning> !== 0,
        0 === <warning descr="Use 'strncmp($url, 'it\'s', 4)' for a length-independent prefix check.">strpos($url, 'it\'s')</warning>,
        0 === <warning descr="Use 'strncasecmp($url, &quot;x\ty&quot;, 3)' for a length-independent prefix check.">stripos($url, "x\ty")</warning>,
        <warning descr="Use 'strncmp($url, 'é/', 3)' for a length-independent prefix check.">strpos($url, 'é/')</warning> === 0,
    ];
}
