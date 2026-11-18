<?php
class LocalConnection { public function prepare($sql) {} } function benign(LocalConnection $c) { $c->prepare('SELECT :a, ?'); }
function good(PDO $pdo) { $s = $pdo->prepare("SELECT ':ignore', a FROM items WHERE b = :b"); }
