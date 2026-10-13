<?php

// A depth counter that starts at zero and is stepped before the call is no
// longer a known non-positive depth.
function incrementedDepth(string $json)
{
    $depth = 0;
    $depth++;
    return json_decode($json, true, $depth);
}

function prefixIncrementedDepth(array $value)
{
    $depth = 0;
    ++$depth;
    return json_encode($value, 0, $depth);
}

function wideningDepth(string $json): void
{
    $depth = 0;
    while ($depth < 2) {
        $depth++;
        echo json_encode(json_decode($json, true, $depth)), "\n";
    }
}

// The step sits after the call but runs before its next iteration.
function steppedAfterCallInLoop(string $json, array $rows): void
{
    $depth = 0;
    foreach ($rows as $row) {
        if ($depth > 0) {
            json_decode($json, true, $depth);
        }
        $depth++;
    }
}

function decrementedFromLimit(string $json)
{
    $depth = 2147483648;
    $depth--;
    return json_decode($json, true, $depth);
}

function reassignedDepth(string $json)
{
    $depth = 0;
    $depth += 16;
    return json_decode($json, true, $depth);
}
