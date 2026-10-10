<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); <warning descr="Omit the empty parameter array to retain existing bindings.">$s->execute([])</warning>; }
