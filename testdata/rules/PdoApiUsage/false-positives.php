<?php
class Other { public function prepare($sql) { return $this; } public function execute() {} }

function work(\PDO $conn, Other $other, $unknown)
{
    $bound = $conn->prepare('SELECT * FROM t WHERE id = ?');
    $bound->execute([7]);

    $empty = $conn->prepare('SELECT 0');
    $empty->execute([]);

    $later = $conn->prepare('SELECT 1');
    log_it('x');
    $later->execute();

    $used = $conn->prepare('SELECT 2');
    if ($used->execute()) {
        return $used;
    }

    $wrapped = ($conn->prepare('SELECT 3'));
    $wrapped->execute();

    $a = $conn->prepare('SELECT 4');
    $b->execute();

    $o = $other->prepare('SELECT 5');
    $o->execute();

    $u = $unknown->prepare('SELECT 6');
    $u->execute();

    $ok = $conn->prepare('SELECT 7');
    $res = $ok->execute();

    $conn->query('SELECT 8');
    return $res;
}
