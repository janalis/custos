<?php
$pdo = new PDO($dsn); $rows = $pdo->query("SELECT 'x' AS category, 7 AS value")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo <warning descr="Read the group column from the group key.">$rows['x'][0]['category']</warning>;
