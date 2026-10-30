<?php
function ranked(string $name, string $dir): int
{
    return str_starts_with($name,
            $dir . '/tests') ? 1 : 0;
}
