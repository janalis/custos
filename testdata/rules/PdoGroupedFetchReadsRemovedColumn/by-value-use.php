<?php

function counted_then_read(PDO $pdo) {
    $rows = $pdo->query("SELECT category, value FROM t")->fetchAll(PDO::FETCH_GROUP | PDO::FETCH_ASSOC);
    foreach ($rows as $key => $group) {
        echo $key, count($group);
    }
    echo count($rows);
    return <warning descr="Read the group column from the group key.">$rows['x'][0]['category']</warning>;
}
