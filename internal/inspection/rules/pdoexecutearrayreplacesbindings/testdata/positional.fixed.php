<?php
function positional(PDO $pdo) { $s = $pdo->prepare('SELECT ?'); $s->bindValue(1, 7); $s->execute(); }
function bare(PDO $pdo, $v) { $s = $pdo->prepare('SELECT :id'); $s->bindParam('id', $v); $s->execute(); }
