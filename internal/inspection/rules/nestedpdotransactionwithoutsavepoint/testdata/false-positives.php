<?php
function restart() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    $ok = $p->rollBack() && $p->beginTransaction();
    return $ok;
}
function ensure() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    $p->inTransaction() || $p->beginTransaction();
}
function conditional() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    if ($p->commit() && $p->beginTransaction()) {
        return 1;
    }
    return 0;
}
function keyed(array $log) {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    $log[$p->rollBack()] = $p->beginTransaction();
    return $log;
}
function deferred() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    $later = function () use ($p) { return $p->beginTransaction(); };
    $p->commit();
    return $later;
}
