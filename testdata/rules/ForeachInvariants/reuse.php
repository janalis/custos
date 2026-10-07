<?php
function pairs(array $keys, array $values, array $rows)
{
    $n = count($keys);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) {
        echo $keys[$i];
    }
    $n = count($values);
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0; $i < $n; $i++) {
        echo $values[$i];
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0, $max = count($keys); $i < $max; $i++) {
        echo $keys[$i];
    }
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($i = 0, $max = count($rows); $i < $max; $i++) {
        echo $rows[$i];
    }
    for ($i = 0; $i < $n; $i++) {
        echo $keys[$i];
    }
}
