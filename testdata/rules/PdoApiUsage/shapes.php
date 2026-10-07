<?php
trait Prepares {
    public function prepare($sql) { return $this; }
    public function execute() {}
}
class Fake { use Prepares; }
function first(\PDO $db, $stmt) {
    $stmt->execute();                       // no previous statement
}
function shapes(\PDO $db, Fake $fake) {
    if ($db) {}
    $a->execute();                          // previous statement is not an expression
    $b = $db->query('SELECT 1');
    $b->execute();                          // not prepare()
    $c = (new ArrayObject())->prepare('x');
    $c->execute();                          // class without prepare()
    $d = $fake->prepare('SELECT 1');
    $d->execute();                          // prepare() declared in a trait
}
