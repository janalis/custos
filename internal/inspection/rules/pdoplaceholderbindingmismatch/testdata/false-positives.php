<?php
class LocalStatement { public function execute($params) {} } function benign(LocalStatement $s) { $s->execute(['wrong' => 1]); }
function viaClosure(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a AS a, :b AS b');
    $bind = function () use ($s) { $s->bindValue(':b', 2); };
    $s->bindValue(':a', 1);
    $bind();
    $s->execute();
}
function viaArrow(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a AS a, :b AS b');
    $s->bindValue(':a', 1);
    array_map(fn ($v) => $s->bindValue(':b', $v), [2]);
    $s->execute();
}
