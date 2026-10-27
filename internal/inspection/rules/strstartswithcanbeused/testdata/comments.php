<?php
function rank(string $a): bool
{
    return <weak_warning descr="Replace with 'str_starts_with($a, 'x')'.">strpos($a, // haystack
        'x' /* needle */) === 0</weak_warning>;
}
