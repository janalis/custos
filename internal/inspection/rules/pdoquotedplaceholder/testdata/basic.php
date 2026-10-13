<?php
function bad(PDO $pdo) { $s = $pdo->prepare("SELECT * FROM items WHERE label = ':label'"); <warning descr="Remove SQL quotes around this parameter marker.">$s->execute(['label' => 'new'])</warning>; }
