<?php
function loopOnly(SQLite3Result $r) {
    while ($row = $r->fetchArray(SQLITE3_ASSOC)) {
        echo json_encode($row);
    }
}
function overwritten(SQLite3Result $r) {
    $row = $r->fetchArray(SQLITE3_ASSOC);
    echo json_encode($row);
    $row = [];
    return $row;
}
function earlierValue(SQLite3Result $r, array $row) {
    echo $row[0];
    $row = $r->fetchArray(SQLITE3_ASSOC);
    return json_encode($row);
}
function unrelatedDynamicScope(SQLite3Result $r) {
    $row = $r->fetchArray(SQLITE3_ASSOC);
    echo json_encode($row);
    return function ($name) { return [$$name, get_defined_vars(), eval('return 1;')]; };
}
