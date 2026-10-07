<?php
class Db { public static $stmt; }
function load(\PDO $pdo) {
    Db::$stmt = $pdo->query('SELECT 1');
}
