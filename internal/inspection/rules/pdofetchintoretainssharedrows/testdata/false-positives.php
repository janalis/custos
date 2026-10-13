<?php
// A new target is installed before every following fetch.
function freshTarget(PDO $pdo)
{
    $rows = [];
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, new stdClass());
    while ($tag = $statement->fetch()) {
        $rows[] = $tag;
        $statement->setFetchMode(PDO::FETCH_INTO, new stdClass());
    }
    return $rows;
}

// The mode may have been replaced on the way to the loop.
function conditionalMode(PDO $pdo, bool $plain)
{
    $rows = [];
    $target = new stdClass();
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    if ($plain) {
        $statement->setFetchMode(PDO::FETCH_OBJ);
    }
    while ($tag = $statement->fetch()) {
        $rows[] = $tag;
    }
    return $rows;
}

// Only one fetched row is ever retained.
function firstRow(PDO $pdo)
{
    $rows = [];
    $target = new stdClass();
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        $rows[] = $tag;
        break;
    }
    return $rows;
}

function firstMatch(PDO $pdo, string $wanted)
{
    $rows = [];
    $target = new stdClass();
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        if ($tag->label === $wanted) {
            $rows[] = $tag;
            return $rows;
        }
    }
    return $rows;
}

function cloned(PDO $pdo)
{
    $rows = [];
    $target = new stdClass();
    $statement = $pdo->query('SELECT label FROM tags');
    $statement->setFetchMode(PDO::FETCH_INTO, $target);
    while ($tag = $statement->fetch()) {
        $rows[] = clone $tag;
    }
    return $rows;
}
