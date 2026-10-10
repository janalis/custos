<?php
function good(PDO $pdo) { $s = $pdo->query('SELECT COUNT(*) FROM items'); $count = $s->fetchColumn(); }
