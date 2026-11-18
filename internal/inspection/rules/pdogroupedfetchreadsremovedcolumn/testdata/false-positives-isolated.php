<?php
$pdo = new PDO($dsn); $rows = $pdo->query("SELECT 'x' AS category, 7 AS value")->fetchAll(PDO::FETCH_GROUP + PDO::FETCH_ASSOC); echo $rows['x'][0]['value'];
