<?php
function bad(PDO $pdo) { $s = <error descr="Use one SQL placeholder style per statement.">$pdo->prepare('SELECT * FROM items WHERE a = :a AND b = ?')</error>; }
