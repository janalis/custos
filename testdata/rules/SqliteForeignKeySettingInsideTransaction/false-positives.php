<?php

class AuditDb extends SQLite3
{
    public function finish()
    {
        $this->exec('COMMIT');
    }
}

function closedInTryCatch()
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    try {
        $db->exec('CREATE TABLE t (a)');
        $db->exec('COMMIT');
    } catch (Exception $e) {
        $db->exec('ROLLBACK');
    }
    $db->exec('PRAGMA foreign_keys=ON');
}

function closedInBothBranches(bool $ok)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    if ($ok) {
        $db->exec('COMMIT');
    } else {
        $db->exec('ROLLBACK');
    }
    $db->exec('PRAGMA foreign_keys=OFF');
}

function maybeClosed(bool $ok)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    $db->exec('INSERT INTO t VALUES (1)') or $db->exec('ROLLBACK');
    $db->exec('PRAGMA foreign_keys=ON');
}

function maybeOpened(bool $batch)
{
    $pdo = new PDO('sqlite::memory:');
    if ($batch) {
        $pdo->beginTransaction();
    }
    $pdo->exec('PRAGMA foreign_keys = ON');
}

function closedByQuery()
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    $db->query('COMMIT');
    $db->exec('PRAGMA foreign_keys=ON');

    $db->exec('BEGIN');
    $db->querySingle('END');
    $db->exec('PRAGMA foreign_keys=OFF');

    $pdo = new PDO('sqlite::memory:');
    $pdo->beginTransaction();
    $pdo->query('ROLLBACK');
    $pdo->exec('PRAGMA foreign_keys = ON');
}

function unknownStatements(string $sql, string $method)
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    $db->query($sql);
    $db->exec('PRAGMA foreign_keys=ON');

    $db->exec('BEGIN');
    $db->query('SELECT 1; COMMIT');
    $db->exec('PRAGMA foreign_keys=ON');

    $db->exec('BEGIN');
    $db->$method();
    $db->exec('PRAGMA foreign_keys=ON');
}

function preparedCommit()
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    $commit = $db->prepare('COMMIT');
    $commit->execute();
    $db->exec('PRAGMA foreign_keys=ON');
}

function reconnected()
{
    $db = new SQLite3(':memory:');
    $db->exec('BEGIN');
    $db->close();
    $db->open(':memory:');
    $db->exec('PRAGMA foreign_keys=ON');
}

function userMethod()
{
    $db = new AuditDb(':memory:');
    $db->exec('BEGIN');
    $db->finish();
    $db->exec('PRAGMA foreign_keys=ON');
}

function otherDriver()
{
    $pdo = new PDO('mysql:host=localhost');
    $pdo->beginTransaction();
    $pdo->exec('PRAGMA foreign_keys = ON');
}
