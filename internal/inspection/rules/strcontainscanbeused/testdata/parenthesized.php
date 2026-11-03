<?php
function wrapped(string $s): bool
{
    if ((<weak_warning descr="Replace with 'str_contains($s, &quot;\r&quot;)'.">strpos($s, "\r") !== false</weak_warning>) || (<weak_warning descr="Replace with '!str_contains($s, &quot;\n&quot;)'.">strpos($s, "\n") === false</weak_warning>)) {
        return true;
    }
    return (/* kept */ <weak_warning descr="Replace with 'str_contains($s, &quot;x&quot;)'.">strpos($s, "x") !== false</weak_warning>);
}
