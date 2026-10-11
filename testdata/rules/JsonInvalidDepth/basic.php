<?php
<error descr="Supply a supported JSON depth.">json_decode('{}', true, 0)</error>;

$topLevelDepth = 0;
<error descr="Supply a supported JSON depth.">json_encode([], 0, $topLevelDepth)</error>;

function unchangedDepth(string $json, array $rows)
{
    $depth = 0;
    $other = 0;
    $other++;
    foreach ($rows as $key => $row) {
        $log = function () use (&$other, $key) {
            return $other;
        };
    }
    try {
        $log();
    } catch (Exception $failure) {
    }
    return <error descr="Supply a supported JSON depth.">json_decode($json, true, $depth)</error>;
}

function incrementedAfterwards(string $json)
{
    $depth = 0;
    $decoded = <error descr="Supply a supported JSON depth.">json_decode($json, true, $depth)</error>;
    $depth++;
    return $decoded;
}
