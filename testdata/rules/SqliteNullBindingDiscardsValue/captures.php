<?php
function capturedByValue(SQLite3Stmt $s) {
    $value = 'draft';
    $copy = function () use ($value) { $value = null; };
    $copy();
    <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, $value, SQLITE3_NULL)</warning>;
}

function readByArrowFunction(SQLite3Stmt $s) {
    $value = 12;
    $read = fn() => [$value, 'k' => $value];
    <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, $value, SQLITE3_NULL)</warning>;
}

function otherVariableCaptured(SQLite3Stmt $s) {
    $value = 1.5;
    $other = 'x';
    $touch = function () use (&$other) { $inner = function () use (&$value) {}; };
    $refs = [&$other, 'k' => 1];
    <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, $value, SQLITE3_NULL)</warning>;
}

function nestedScopesAreSeparate(SQLite3Stmt $s) {
    $value = 'draft';
    function helper() { $value = 'x'; $g = function () use (&$value) {}; }
    $o = new class { function m() { $value = 1; return [&$value]; } };
    <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, $value, SQLITE3_NULL)</warning>;
}

$outer = function (SQLite3Stmt $s) {
    $value = 'draft';
    <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, $value, SQLITE3_NULL)</warning>;
};

$top = 'draft';
$stmt = (new SQLite3(':memory:'))->prepare('SELECT ?');
<warning descr="Pass null when using the SQLite null binding type.">$stmt->bindValue(1, $top, SQLITE3_NULL)</warning>;
