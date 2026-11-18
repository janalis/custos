<?php
function good(PDO $pdo) { $s = $pdo->prepare('SELECT :a, :b'); $a = 1; $b = 2; $s->bindParam(':a', $a); $s->bindParam(':b', $b); $s->execute(); }

function unknownStatementEffects(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a, :b');
    $value = 1;
    $s->bindParam(':a', $value);
    inspectStatement($s);
    $value = 2;
    $s->bindParam(':b', $value);
    $s->execute();
}

function unknownArrayCarrierEffects(PDO $pdo) {
    $s = $pdo->prepare('SELECT :a, :b');
    $value = 1;
    $s->bindParam(':a', $value);
    inspectStatement([$s]);
    $value = 2;
    $s->bindParam(':b', $value);
    $s->execute();
}

function conditionalStatementEffects(PDO $pdo, $inspect) {
    $s = $pdo->prepare('SELECT :a, :b');
    $value = 1;
    $s->bindParam(':a', $value);
    if ($inspect) {
        inspectStatement($s);
    }
    $value = 2;
    $s->bindParam(':b', $value);
    $s->execute();
}
