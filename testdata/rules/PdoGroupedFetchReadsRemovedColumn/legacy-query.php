<?php

function chained(PDO $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    return <warning descr="Read the group column from the group key.">$rows['x'][0]['category']</warning>;
}

function through_statement(PDO $pdo) {
    $statement = $pdo->query("SELECT category, value FROM t");
    $rows = $statement->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    return <warning descr="Read the group column from the group key.">$rows['x'][0]['category']</warning>;
}

function prepared(PDO $pdo) {
    $statement = $pdo->prepare("SELECT category, value FROM t");
    $statement->execute();
    $rows = $statement->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    return <warning descr="Read the group column from the group key.">$rows['x'][0]['category']</warning>;
}

class LegacyGateway extends PDO
{
    public function query($sql)
    {
        return new LegacyResult($sql);
    }
}

function overridden(LegacyGateway $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    return $rows['x'][0]['category'];
}
