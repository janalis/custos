<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); <error descr="Match SQL placeholders to the effective parameter bindings.">$s->execute(['other' => 2])</error>; }
