<?php
function good(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); $s->execute(); }
