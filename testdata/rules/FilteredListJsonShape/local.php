<?php
function plainLocal() {
    $rows = [1, 0, 2];
    $copy = function () use ($rows) { return $rows; };
    return <warning descr="Reindex the filtered array before encoding a JSON list.">json_encode(array_filter($rows))</warning>;
}
