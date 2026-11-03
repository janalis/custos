<?php
// A failed read unserializes to false like corrupt text, which the test rejects.
function merge($h, array $logs): array
{
    $previous = unserialize(stream_get_contents($h), ['allowed_classes' => false]);
    if (is_array($previous)) {
        $logs = array_merge($previous, $logs);
    }
    return $logs;
}

function tests($h, $path)
{
    if (false === ($state = unserialize(file_get_contents($path)))) {
        return null;
    }
    if (unserialize(data: stream_get_contents($h)) instanceof \stdClass) {
        return 1;
    }
    return unserialize(file_get_contents($path)) ?: [];
}

function used($h, $path)
{
    $value = <warning descr="Handle file reading failure before using its result.">unserialize(stream_get_contents($h))</warning>;
    echo count(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
    $kept = <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning> ?? [];
    $loose = <warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning> == 0;
    $other = mine(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>);
    return [$value, $kept, $loose, $other, is_callable(<warning descr="Handle file reading failure before using its result.">unserialize(file_get_contents($path))</warning>)];
}

function mine($v)
{
    return $v;
}
