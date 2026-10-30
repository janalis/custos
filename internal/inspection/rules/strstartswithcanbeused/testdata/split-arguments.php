<?php
function ranked(string $name, string $dir): int
{
    return <weak_warning descr="Replace with 'str_starts_with($name, $dir . '/tests')'.">strpos($name,
            $dir . '/tests') === 0</weak_warning> ? 1 : 0;
}
