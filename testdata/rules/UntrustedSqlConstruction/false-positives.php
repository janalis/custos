<?php
function benign(PDO $pdo) { $id = intval($_GET['id']); $pdo->query('SELECT * FROM items WHERE id = ' . $id); }

// Only placeholders reach the statement text; the values are bound.
function byIds(PDO $pdo) {
    $stmt = $pdo->prepare('SELECT * FROM items WHERE id IN (' . implode(',', array_fill(0, count($_GET['ids']), '?')) . ')');
    $stmt->execute($_GET['ids']);
    $marks = str_repeat('?,', count($_GET['ids']) - 1) . '?';
    $pdo->prepare("SELECT * FROM items WHERE id IN ($marks)");
    $pdo->query('SELECT ' . implode(',', [count($_GET['ids']), '1']));
    $pdo->prepare('SELECT * FROM items WHERE id IN (' . join(', ', ['?', '?']) . ') AND tag = ' . sizeof($_GET['ids']));
}

// Numbers and booleans cannot carry SQL syntax.
function numeric(PDO $pdo) {
    $pdo->query('SELECT * FROM items LIMIT ' . count($_GET['ids']));
    $pdo->query('SELECT * FROM items WHERE len = ' . strlen($_GET['q']));
    $pdo->query('SELECT * FROM items WHERE pos = ' . strpos($_GET['q'], 'x'));
    $pdo->query('SELECT * FROM items LIMIT 10 OFFSET ' . ($_GET['page'] * 10));
    $pdo->query('SELECT * FROM items LIMIT 10 OFFSET ' . ($_GET['page'] - 1));
    $pdo->query('SELECT * FROM items WHERE id = ' . ($_GET['id'] + 0));
    $pdo->query('SELECT * FROM items WHERE id = ' . (1.5 + $_GET['id']));
    $pdo->query('SELECT * FROM items WHERE id = ' . ($_GET['id'] + strlen($_GET['q'])));
    $pdo->query('SELECT * FROM items WHERE id = ' . ($_GET['id'] + (int) $_GET['step']));
    $pdo->query('SELECT * FROM items WHERE id = ' . ($_GET['id'] + ($_GET['step'] % 5)));
    $pdo->query('SELECT * FROM items WHERE id = ' . ($_GET['id'] + (2 + $_GET['step'])));
    $pdo->query('SELECT * FROM items WHERE id = ' . +$_GET['id']);
    $pdo->query('SELECT * FROM items WHERE id = ' . -$_GET['id']);
    $pdo->query('SELECT * FROM items LIMIT ' . (isset($_GET['n']) ? count($_GET['n']) : 10));
    $pdo->query('SELECT * FROM items LIMIT ' . (count($_GET['n']) ?: 10));
    $pdo->query('SELECT * FROM items LIMIT ' . ($_GET['limit'] ?? count($_GET['n'])) * 2);
}

// The subject only chooses between constant keywords.
function listing(PDO $pdo) {
    $dir = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
    return $pdo->query('SELECT * FROM items ORDER BY id ' . $dir);
}

// Not PDO sinks.
class Gateway { function query($sql) {} }
class Recorder extends PDO { public function query($sql, ...$rest) { return false; } }
function others(Gateway $g, Recorder $r, PDO $pdo, $name) {
    $g->query('SELECT ' . $_GET['x']);
    $r->query('SELECT ' . $_GET['x']);
    $pdo->quote($_GET['x']);
    $pdo->$name('SELECT ' . $_GET['x']);
    $pdo->query(...['SELECT ' . $_GET['x']]);
}

function run(PDO $pdo, $sql) { return $pdo->query($sql); }
function relay(PDO $pdo, $sql) { return run($pdo, $sql); }
function shell($command) { system($command); }
function wrapped(PDO $pdo) {
    run($pdo, 'SELECT * FROM items WHERE id = ' . intval($_GET['id']));
    shell('ls');
    strlen($_GET['x']);
}
