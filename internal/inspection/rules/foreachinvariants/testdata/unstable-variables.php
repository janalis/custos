<?php
function trimmed(array $rows)
{
    $limit = count($rows);
    if ($limit > 10) {
        $limit -= 1;
    }
    for ($i = 0; $i < $limit; $i++) {
        echo $rows[$i];
    }

    $max = count($rows);
    --$max;
    for ($j = 0; $j < $max; $j++) {
        echo $rows[$j];
    }
}
