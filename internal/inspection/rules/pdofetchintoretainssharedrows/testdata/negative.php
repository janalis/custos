<?php
$object = new stdClass(); $pdo = new PDO($dsn); $s = $pdo->query('SELECT name FROM people'); $s->setFetchMode(PDO::FETCH_INTO, $object); while ($row = $s->fetch()) { $rows[] = clone $row; }
