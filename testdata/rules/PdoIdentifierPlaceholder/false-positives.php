<?php
class LocalConnection { public function prepare($sql) {} } function benign(LocalConnection $c) { $c->prepare('SELECT * FROM :table'); }
function valueAfterFrom(PDO $pdo) {
    $a = $pdo->prepare('SELECT 1 AS hit WHERE 2 IS DISTINCT FROM :v');
    $b = $pdo->prepare('SELECT 1 AS hit WHERE 2 IS NOT DISTINCT FROM ?');
    $c = $pdo->prepare('SELECT EXTRACT(YEAR FROM :d)');
    $d = $pdo->prepare('SELECT SUBSTRING(name FROM :start FOR :len) FROM people');
    $e = $pdo->prepare("SELECT TRIM(BOTH 'x' FROM :s)");
    $f = $pdo->prepare('SELECT OVERLAY(name PLACING :p FROM ? FOR 2) FROM people');
    $g = $pdo->prepare('SELECT * FROM people WHERE id = ? -- FROM ?');
    return [$a, $b, $c, $d, $e, $f, $g];
}
