<?php
function bad(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); }
