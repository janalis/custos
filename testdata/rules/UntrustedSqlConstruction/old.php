<?php
function bad(PDO $pdo) {
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->exec('DELETE FROM people WHERE id = ' . $_GET['id'])</error>;
    $pdo->query('SELECT * FROM people LIMIT ' . count($_GET['ids']));
}

class Repository extends PDO {}
class Recorder extends PDO { public function query($sql) { return false; } }

function inherited(Repository $db, Recorder $r) {
    <error descr="Bind untrusted values as SQL parameters.">$db->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>;
    $r->query('SELECT * FROM people WHERE id = ' . $_GET['id']);
}
