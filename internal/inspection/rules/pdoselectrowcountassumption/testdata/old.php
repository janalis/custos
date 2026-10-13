<?php
function queried(PDO $pdo) { $s = $pdo->query('SELECT * FROM items'); return <warning descr="Use an explicit count query for portable row counts.">$s->rowCount()</warning>; }
function chained(PDO $pdo) { return <warning descr="Use an explicit count query for portable row counts.">$pdo->query('SELECT id FROM items')->rowCount()</warning>; }
function prepared(PDO $pdo) { $s = $pdo->prepare('SELECT * FROM items'); $s->execute(); return <warning descr="Use an explicit count query for portable row counts.">$s->rowCount()</warning>; }
function affected(PDO $pdo) { $s = $pdo->query('DELETE FROM items'); return $s->rowCount(); }
function unknownOrigin($s) { return $s->rowCount(); }
function otherMethod(PDO $pdo) { $s = $pdo->query('SELECT * FROM items'); return $s->columnCount(); }
