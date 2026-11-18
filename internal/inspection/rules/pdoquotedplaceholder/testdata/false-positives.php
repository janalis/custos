<?php
function good(PDO $pdo) { $s = $pdo->prepare('SELECT * FROM items WHERE label = :label'); $s->execute(['label' => 'new']); }
