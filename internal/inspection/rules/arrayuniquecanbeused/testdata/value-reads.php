<?php
function read() {
    $ids = [1, 2, 1];
    foreach ($ids as $id) {
        echo $id, strlen(...)('x');
    }
    echo count($ids), new ArrayObject;
    return <weak_warning descr="Use 'array_values(array_unique($ids))' instead (array_unique() is fast since PHP 7.2).">array_keys(array_count_values($ids))</weak_warning>;
}
function spread(int ...$n) {
    return <weak_warning descr="Use 'count(array_unique($n))' instead (array_unique() is fast since PHP 7.2).">count(array_count_values($n))</weak_warning>;
}
$top = ['a', 'b'];
echo <weak_warning descr="Use 'count(array_unique($top))' instead (array_unique() is fast since PHP 7.2).">count(array_count_values($top))</weak_warning>;
