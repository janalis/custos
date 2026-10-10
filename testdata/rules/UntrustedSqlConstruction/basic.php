<?php
function bad(PDO $pdo) { <error descr="Bind untrusted values as SQL parameters.">$pdo->query('SELECT * FROM people WHERE id = ' . $_GET['id'])</error>; }
