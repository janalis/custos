<?php
// A callee that may write the globals table may replace a file-level variable.
function frfuTable(): array
{
    return ['globals' => $GLOBALS];
}

function frfuUse(string $text): string
{
    return $text;
}

foreach ($argv as $path) {
    $checked = file_get_contents($path);
    if (!preg_match('/--EXPECT--(.*)$/s', $checked, $matches)) {
        continue;
    }
    frfuTable();
    echo frfuUse($checked);
    $raw = file_get_contents($path);
    echo <warning descr="Handle file reading failure before using its result.">frfuUse($raw)</warning>;
}
