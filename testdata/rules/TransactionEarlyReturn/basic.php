<?php
function bad(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); }
function skippable(PDO $pdo, $saved) { if (!$pdo->beginTransaction()) { return false; } <warning descr="Close the owned transaction before returning.">return $saved && $pdo->commit();</warning> }
function otherGuard(PDO $pdo, bool $skip, bool $undo) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); } finally { if ($undo) { $pdo->rollBack(); } } }
