<?php

function insertThenToggle(bool $verbose)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    if ($verbose) {
        $db->exec('INSERT INTO log VALUES (1)');
    }
    $db->query('SELECT 1');
    <warning descr="Change foreign-key enforcement before the transaction.">$db->exec('PRAGMA foreign_keys=OFF')</warning>;
}

function openedByQuery()
{
    $db = new SQLite3(':memory:');
    $db->query('BEGIN IMMEDIATE');
    <warning descr="Change foreign-key enforcement before the transaction.">$db->exec('PRAGMA foreign_keys=ON')</warning>;
}

function reopened(bool $flush)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    if ($flush) {
        $db->exec('COMMIT');
    }
    $db->exec('BEGIN');
    <warning descr="Change foreign-key enforcement before the transaction.">$db->exec('PRAGMA foreign_keys=ON')</warning>;
}

function pdoTransaction()
{
    $pdo = new PDO('sqlite::memory:');
    $pdo->beginTransaction();
    $pdo->prepare('SELECT 1');
    <warning descr="Change foreign-key enforcement before the transaction.">$pdo->exec('PRAGMA foreign_keys = OFF')</warning>;
}

function abortedBranchLeaves(bool $failed)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    if ($failed) {
        $db->exec('ROLLBACK');
        return;
    }
    <warning descr="Change foreign-key enforcement before the transaction.">$db->exec('PRAGMA foreign_keys=ON')</warning>;
}
