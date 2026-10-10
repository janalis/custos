<?php
$pdo = new PDO($dsn); if ($pdo->exec('DELETE FROM jobs WHERE done = 1') === false) { throw new RuntimeException(); }
