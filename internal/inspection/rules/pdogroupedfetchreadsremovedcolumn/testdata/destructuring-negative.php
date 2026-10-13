<?php
function destructuring_targets($raw, $uid, $dsn, $items) {
$pdo = new PDO($dsn);
$rows = $pdo->query("SELECT 'x' AS category, 7 AS value")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC);
[$rows['x'][0]['category']] = ['x'];
}
