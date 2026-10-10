<?php
$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindValue(':n', $n); $n = 11; <warning descr="Rebind the changed value before execution.">$s->execute()</warning>;
