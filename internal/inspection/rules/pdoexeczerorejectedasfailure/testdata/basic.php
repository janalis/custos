<?php
$pdo = new PDO($dsn); if (!<warning descr="Compare database execution failure strictly with false.">$pdo->exec('DELETE FROM jobs WHERE done = 1')</warning>) { throw new RuntimeException(); }
