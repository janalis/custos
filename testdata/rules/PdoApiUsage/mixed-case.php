<?php
function probe(\PDO $conn) {
    $st = $conn->Prepare('SELECT 1');
    <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$st->EXECUTE()</weak_warning>;
    return $st;
}
