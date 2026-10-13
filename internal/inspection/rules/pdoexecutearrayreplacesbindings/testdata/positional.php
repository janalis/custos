<?php
function positional(PDO $pdo) { $s = $pdo->prepare('SELECT ?'); $s->bindValue(1, 7); <warning descr="Omit the empty parameter array to retain existing bindings.">$s->execute([])</warning>; }
function bare(PDO $pdo, $v) { $s = $pdo->prepare('SELECT :id'); $s->bindParam('id', $v); <warning descr="Omit the empty parameter array to retain existing bindings.">$s->execute([])</warning>; }
