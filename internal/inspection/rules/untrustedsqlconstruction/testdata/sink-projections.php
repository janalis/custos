<?php
function numericRows(PDO $db, $rows) { $db->query('SELECT ' . implode(',', $rows)); }
function numericKeys(PDO $db, $rows) { $db->query('SELECT ' . implode(',', array_keys($rows))); }
function sendNumeric(PDO $db) {
    $rows = array_map('intval', $_GET['rows']);
    numericRows($db, $rows);
    <error descr="Bind untrusted values as SQL parameters.">numericKeys($db, $rows)</error>;
}
final class QueryGateway {
    public function send(PDO $db, $sql) { $db->query($sql); }
    public static function dispatch(PDO $db, $sql) { $db->exec($sql); }
}
class OpenQueryGateway {
    public function send(PDO $db, $sql) { $db->query($sql); }
}
function queries(QueryGateway $exact, OpenQueryGateway $open, PDO $db) {
    <error descr="Bind untrusted values as SQL parameters.">$exact->send($db, $_GET['sql'])</error>;
    <error descr="Bind untrusted values as SQL parameters.">QueryGateway::dispatch(sql: $_GET['sql'], db: $db)</error>;
    $open->send($db, $_GET['sql']);
}
