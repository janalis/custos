<?php
function plainLocal() {
    $rows = [1, 0, 2];
    $copy = function () use ($rows) { return $rows; };
    return json_encode(\array_values(array_filter($rows)));
}
