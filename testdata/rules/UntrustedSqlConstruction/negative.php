<?php
function good(PDO $pdo) { $s = $pdo->prepare('SELECT * FROM people WHERE id = :id'); $s->execute(['id' => $_GET['id']]); }
