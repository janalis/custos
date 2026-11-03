<?php
function wrapped(string $s): bool
{
    if (str_contains($s, "\r") || !str_contains($s, "\n")) {
        return true;
    }
    return (/* kept */ str_contains($s, "x"));
}
