<?php
function clearedThroughCapture(SQLite3Stmt $s) {
    $value = 'draft';
    $clear = function () use (&$value) { $value = null; };
    $clear();
    $s->bindValue(1, $value, SQLITE3_NULL);
}

function clearedThroughStaticCapture(SQLite3Stmt $s) {
    $value = 'draft';
    $clear = static function () use ($s, &$value) { $value = null; };
    $clear();
    $s->bindValue(value: ($value), param: 1, type: SQLITE3_NULL);
}

function clearedThroughArrayReference(SQLite3Stmt $s) {
    $value = 'draft';
    $holder = ['slot' => &$value];
    reset_slots($holder);
    $s->bindValue(1, $value, SQLITE3_NULL);
}

function capturedAfterLoopBody(SQLite3Stmt $s, array $rows) {
    $value = 'draft';
    foreach ($rows as $row) {
        $s->bindValue(1, $value, SQLITE3_NULL);
        $row(function () use (&$value) { $value = null; });
    }
}

$top = 'draft';
$wipe = function () use (&$top) { $top = null; };
$wipe();
$stmt = (new SQLite3(':memory:'))->prepare('SELECT ?');
$stmt->bindValue(1, $top, SQLITE3_NULL);
