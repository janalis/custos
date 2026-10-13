<?php
function pairs(array $keys, array $values, array $rows)
{
    $n = count($keys);
    foreach ($keys as $iValue) {
        echo $iValue;
    }
    $n = count($values);
    foreach ($values as $iValue) {
        echo $iValue;
    }
    foreach ($keys as $iValue) {
        echo $iValue;
    }
    foreach ($rows as $iValue) {
        echo $iValue;
    }
    for ($i = 0; $i < $n; $i++) {
        echo $keys[$i];
    }
}
