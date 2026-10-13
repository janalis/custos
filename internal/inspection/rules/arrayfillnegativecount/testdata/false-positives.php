<?php

function countedRows(array $rows)
{
    $last = -1;
    foreach ($rows as $row) {
        $last++;
    }
    return array_fill(0, $last, null);
}

function steppedTwice()
{
    $width = -1;
    ++$width;
    $width++;
    return array_fill(0, $width, 'c');
}

function copiedAfterStep()
{
    $seed = -1;
    $seed++;
    $width = $seed;
    return array_fill(0, ($width), 'c');
}

function copiedThroughTemporary()
{
    $width = -1;
    $next = $width;
    $next++;
    $width = $next;
    return array_fill(0, $width, 'c');
}

function closureWritesCount()
{
    $width = -1;
    $grow = function () use (&$width) {
        $width = 6;
    };
    $grow();
    return array_fill(0, $width, 'c');
}

function referencedArraySlot()
{
    $width = -1;
    $slots = [&$width];
    $slots[0]++;
    return array_fill(0, $width, 'c');
}

function extractedCount(array $settings)
{
    $width = -1;
    extract($settings);
    return array_fill(0, $width, 'c');
}

function reassigned()
{
    $width = -1;
    $width = 2;
    return array_fill(0, $width, 'c');
}

function unknownCount($width)
{
    return array_fill(0, $width, 'c');
}

array_fill(0, 0, 'c');
array_fill(-3, 2, 'c');
