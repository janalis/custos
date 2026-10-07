<?php
class Db { public static $stmt; }
function load(\PDO $pdo) {
    Db::$stmt = $pdo->prepare('SELECT 1');
    <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">db::$stmt->execute()</weak_warning>;
}
