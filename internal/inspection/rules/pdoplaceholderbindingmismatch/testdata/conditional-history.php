<?php
function redundantBinding(PDO $pdo, $flag) {
    $s = $pdo->prepare('SELECT :a, :b');
    $s->bindValue(':a', 1);
    if ($flag) { $s->bindValue(':a', 2); }
    <error descr="Match SQL placeholders to the effective parameter bindings.">$s->execute()</error>;
}
function optionalBinding(PDO $pdo, $flag) {
    $s = $pdo->prepare('SELECT :a, :b');
    $s->bindValue(':a', 1);
    if ($flag) { $s->bindValue(':b', 2); }
    $s->execute();
}
