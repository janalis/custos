<?php

function restored_by_reference_loop(PDO $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    foreach ($rows as $key => &$group) {
        foreach ($group as &$row) {
            $row['category'] = $key;
        }
    }
    unset($group, $row);
    return $rows['x'][0]['category'];
}

function restored_by_walk(PDO $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    array_walk($rows, function (&$group, $key) {
        foreach ($group as &$row) {
            $row['category'] = $key;
        }
    });
    return $rows['x'][0]['category'];
}

function restored_through_alias(PDO $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    $first = &$rows['x'];
    $first[0]['category'] = 'x';
    return $rows['x'][0]['category'];
}

function handed_to_collaborator(PDO $pdo, $hydrator) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    $hydrator->restoreKeys($rows);
    return $rows['x'][0]['category'];
}
