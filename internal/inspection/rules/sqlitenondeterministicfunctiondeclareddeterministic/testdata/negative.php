<?php
$db = new SQLite3(':memory:'); $db->createFunction('twice', fn($n) => $n * 2, 1, SQLITE3_DETERMINISTIC);
