<?php
class Db { function query($s) {} }
class AuditedPdo extends PDO { function query($s, ...$rest) {} }
function shapes(PDO $p, mysqli $m, Db $d, AuditedPdo $a, PDOStatement $st, $name){
    $p->query("SELECT id FROM t WHERE x IS NOT NULL = 1");
    $p->query("SELECT id FROM t WHERE x IS NULL = y IS NULL");
    $p->query("SELECT id FROM t WHERE id IN (SELECT a = NULL FROM u)");
    $p->query("SELECT (SELECT 1 FROM u WHERE a = 1), x = NULL FROM t");
    $d->query("SELECT id FROM t WHERE x = NULL");
    $a->query("SELECT id FROM t WHERE x = NULL");
    $m->exec("SELECT id FROM t WHERE x = NULL");
    $p->quote("SELECT id FROM t WHERE x = NULL");
    $p->$name("SELECT id FROM t WHERE x = NULL");
    $st->execute(["SELECT id FROM t WHERE x = NULL"]);
}
function either(PDO|mysqli $x){$x->query("SELECT id FROM t WHERE x = NULL");}
