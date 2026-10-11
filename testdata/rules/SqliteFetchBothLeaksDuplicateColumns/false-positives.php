<?php
function writeThenRead(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row), "\n";
    $row['seen'] = true;
    return $row[0];
}
function unsetThenRead(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row), "\n";
    unset($row['secret']);
    return $row[0];
}
function appendThenDestructure(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row);
    $row[] = 1;
    list($id) = $row;
    return $id;
}
function loopCarried(array $results) {
    $row = null;
    foreach ($results as $r) {
        if ($row !== null) { echo "previous ", $row[0], "\n"; }
        $row = $r->fetchArray();
        echo json_encode($row), "\n";
    }
}
function loopOverwrite(SQLite3Result $r, array $xs) {
    foreach ($xs as $x) {
        $row = $r->fetchArray();
        echo json_encode($row);
        $row = [$x];
        echo $row[0];
    }
}
function compacted(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row);
    return compact('row');
}
function definedVars(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row);
    return \get_defined_vars();
}
function extracted(SQLite3Result $r, array $data) {
    $row = $r->fetchArray();
    EXTRACT($data, EXTR_REFS);
    echo json_encode($row);
}
function variableVariable(SQLite3Result $r) {
    $row = $r->fetchArray();
    $name = 'row';
    echo json_encode($row), ${$name}[0];
}
function included(SQLite3Result $r) {
    $row = $r->fetchArray();
    echo json_encode($row);
    include __DIR__ . '/template.php';
}
function evaluated(SQLite3Result $r, $code) {
    $row = $r->fetchArray();
    echo json_encode($row);
    eval($code);
}
function jumped(SQLite3Result $r) {
    $row = null;
    again:
    if ($row) { return $row[0]; }
    $row = $r->fetchArray();
    echo json_encode($row);
    goto again;
}
function byReferenceParameter(SQLite3Result $r, &$row) {
    $row = $r->fetchArray();
    echo json_encode($row);
}
function persisted(SQLite3Result $r) {
    static $row;
    if ($row) { return $row[0]; }
    $row = $r->fetchArray();
    echo json_encode($row);
}
function imported(SQLite3Result $r) {
    global $row;
    $row = $r->fetchArray();
    echo json_encode($row);
}
function capturedByReference(SQLite3Result $r) {
    $row = null;
    return function () use ($r, &$row) {
        $row = $r->fetchArray();
        echo json_encode($row);
    };
}
function compound(SQLite3Result $r) {
    $row = [];
    $row += $r->fetchArray();
    echo json_encode($row);
}
function callableAndSpread(SQLite3Result $r) {
    echo json_encode($r->fetchArray(...));
    echo json_encode($r->fetchArray(...[SQLITE3_NUM]));
}
