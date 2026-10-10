<?php
$object = new stdClass(); $pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { <warning descr="Clone fetched objects before retaining snapshots.">$rows[] = $row</warning>; }
