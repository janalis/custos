<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); <error descr="Match SQL placeholders to the effective parameter bindings.">$s->execute(['other' => 2])</error>; }
function laterCapture(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a AS a, :b AS b');
    $log = function () use ($pdo) { return $pdo->errorCode(); };
    $s->bindValue(':a', 1);
    <error descr="Match SQL placeholders to the effective parameter bindings.">$s->execute()</error>;
    return fn () => $s->fetch();
}
