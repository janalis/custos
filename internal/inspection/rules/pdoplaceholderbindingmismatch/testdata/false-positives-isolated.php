<?php
function good(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->execute(['id' => 2]); }
