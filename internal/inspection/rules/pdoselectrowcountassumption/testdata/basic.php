<?php
function bad(PDO $pdo) { $s = $pdo->query('SELECT * FROM items'); $count = <warning descr="Use an explicit count query for portable row counts.">$s->rowCount()</warning>; }
