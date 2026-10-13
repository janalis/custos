<?php
function loopOnly(SQLite3Result $r) {
    while ($row = <warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray()</warning>) {
        echo json_encode($row);
    }
}
function overwritten(SQLite3Result $r) {
    $row = <warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray(SQLITE3_BOTH)</warning>;
    echo json_encode($row);
    $row = [];
    return $row;
}
function earlierValue(SQLite3Result $r, array $row) {
    echo $row[0];
    $row = <warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray()</warning>;
    return json_encode($row);
}
function unrelatedDynamicScope(SQLite3Result $r) {
    $row = <warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray()</warning>;
    echo json_encode($row);
    return function ($name) { return [$$name, get_defined_vars(), eval('return 1;')]; };
}
