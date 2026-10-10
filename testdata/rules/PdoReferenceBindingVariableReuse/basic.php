<?php
function bad(PDO $pdo) { $s = $pdo->prepare('SELECT :a, :b'); $v = 0; $s->bindParam(':a', $v); $v = 1; $s->bindParam(':b', $v); $v = 2; <warning descr="Bind distinct parameters to distinct scalar identities.">$s->execute()</warning>; }
