<?php
$db=new SQLite3(':memory:'); $db->createFunction('constant_seven', fn()=>strlen(random_bytes(7)), 0, SQLITE3_DETERMINISTIC); var_dump($db->querySingle('SELECT constant_seven()'));

$db->createFunction('fixed_bounds', fn() => random_int(7, 7), 0, SQLITE3_DETERMINISTIC);
$db->createFunction('fixed_parameter', fn($n) => random_int($n, $n), 1, SQLITE3_DETERMINISTIC);
