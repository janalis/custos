<?php
function probe(\PDO $conn) {
    $st = $conn->query('SELECT 1');
    return $st;
}
