<?php
class Gateway { public function query($sql) { return new Rows(); } }
class Rows { public function rowCount() { return 0; } }
class Wrapped extends PDO {
    #[\ReturnTypeWillChange]
    public function query($sql, ...$rest) { return new Rows(); }
}
function userGateway(Gateway $g) { $s = $g->query('SELECT * FROM items'); return $s->rowCount(); }
function overridden(Wrapped $w) { $s = $w->query('SELECT * FROM items'); return $s->rowCount(); }
function affected(PDO $pdo) { $s = $pdo->query('UPDATE items SET seen = 1'); return $s->rowCount(); }
function dynamicSql(PDO $pdo, $sql) { $s = $pdo->query($sql); return $s->rowCount(); }
function dynamicMethod(PDO $pdo, $m) { $s = $pdo->query('SELECT * FROM items'); return $s->$m(); }
function reassigned(PDO $pdo, $other) { $s = $pdo->query('SELECT * FROM items'); $s = $other; return $s->rowCount(); }
