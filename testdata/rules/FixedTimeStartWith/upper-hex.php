<?php
function upperHex($line)
{
    return [
        <warning descr="Use 'strncmp($line, &quot;\X09&quot;, 1)' for a length-independent prefix check.">strpos($line, "\X09")</warning> === 0,
        <warning descr="Use 'strncmp($line, &quot;\X41-&quot;, 2)' for a length-independent prefix check.">strpos($line, "\X41-")</warning> === 0,
        0 !== <warning descr="Use 'strncasecmp($line, &quot;\X4a\x4B&quot;, 2)' for a length-independent prefix check.">stripos($line, "\X4a\x4B")</warning>,
        <warning descr="Use 'strncmp($line, &quot;\\X41&quot;, 4)' for a length-independent prefix check.">strpos($line, "\\X41")</warning> === 0,
        <warning descr="Use 'strncmp($line, &quot;\Xzz&quot;, 4)' for a length-independent prefix check.">strpos($line, "\Xzz")</warning> === 0,
        <warning descr="Use 'strncmp($line, '\X41', 4)' for a length-independent prefix check.">strpos($line, '\X41')</warning> === 0,
    ];
}
