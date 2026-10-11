<?php
function bad(PDO $pdo) { <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>; }

function sinks(PDO $pdo) {
    <error descr="Bind untrusted values as SQL parameters.">$pdo->exec('DELETE FROM people WHERE id = ' . $_POST['id'])</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->prepare("SELECT * FROM people WHERE id = {$_GET['id']}")</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query(query: 'SELECT * FROM people WHERE id = ' . $_GET['id'])</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query($_GET)</error>;
}

function throughVariables(PDO $pdo, bool $flag) {
    $where = 'name = "' . $_GET['name'] . '"';
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE ' . $where)</error>;
    $order = 'id';
    if ($flag) {
        $order = $_GET['order'];
    }
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people ORDER BY ' . $order)</error>;
}

function selections(PDO $pdo) {
    $column = match ($_GET['sort']) { 'name' => 'name', default => $_GET['sort'] };
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people ORDER BY ' . $column)</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people ORDER BY ' . (strlen($_GET['sort']) ?: $_GET['fallback']))</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people ORDER BY ' . (isset($_GET['sort']) ? 'id' : $_GET['fallback']))</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people ORDER BY ' . ($_GET['sort'] ?? 'id'))</error>;
}

function lists(PDO $pdo) {
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . implode(',', $_GET['ids']) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . implode($_GET['glue'], ['1', '2']) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . implode(',', ['1', $_GET['id']]) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . implode($_GET['ids']) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . join(',', array_fill(0, 3, $_GET['id'])) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT 1 ' . str_repeat($_GET['tail'], 2))</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT 1 ' . trim($_GET['tail']))</error>;
}

function operators(PDO $pdo) {
    // Two request arrays form a union whose text is still request text.
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id IN (' . implode(',', $_GET['a'] + $_POST['b']) . ')')</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE flags = ' . ($_GET['a'] | $_GET['b']))</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE flags = ' . ~$_GET['a'])</error>;
    <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id = ' . ($_GET['a'] + '1'))</error>;
}

class Repository extends PDO {}

function inherited(Repository $db) { <error descr="Bind untrusted values as SQL parameters.">$db->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>; }

function run(PDO $pdo, $sql) { return $pdo->query($sql); }

function wrapped(PDO $pdo) { <error descr="Bind untrusted values as SQL parameters.">run($pdo, 'SELECT * FROM people WHERE id = ' . $_GET['id'])</error>; }
