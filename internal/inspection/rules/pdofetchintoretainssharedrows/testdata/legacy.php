<?php
// The statement type is proven through PDO::query() on targets below PHP 8.0.
function collect(PDO $pdo)
{
    $rows = array();
    $target = new stdClass();
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        <warning descr="Clone fetched objects before retaining snapshots.">$rows[] = $tag</warning>;
    }
    return $rows;
}

function collectPrepared(PDO $pdo)
{
    $rows = array();
    $target = new stdClass();
    $statement = $pdo->prepare('SELECT label FROM tags');
    $statement->execute();
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        <warning descr="Clone fetched objects before retaining snapshots.">$rows[] = $tag</warning>;
    }
    return $rows;
}

function unknownStatement($statement)
{
    $rows = array();
    $target = new stdClass();
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        $rows[] = $tag;
    }
    return $rows;
}
