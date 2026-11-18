<?php
$pdo = new PDO($dsn); $s = $pdo->prepare('SELECT :n'); $n = 7; $s->bindParam(':n', $n); $n = 11; $s->execute();
