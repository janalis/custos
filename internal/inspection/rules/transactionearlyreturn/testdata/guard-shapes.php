<?php
// The finally block closes the transaction whenever one is still active.
function shortCircuit(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } $pdo->commit(); } finally { $pdo->inTransaction() && $pdo->rollBack(); } }
function shortCircuitOr(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } $pdo->commit(); } finally { !$pdo->inTransaction() or $pdo->rollBack(); } }
function elseArm(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } $pdo->commit(); } finally { if (!$pdo->inTransaction()) { $state = 'closed'; } else { $pdo->rollBack(); } } }
// A cleanup that may be skipped while the transaction is active proves nothing.
function otherOperand(PDO $pdo, bool $skip, bool $undo) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); } finally { echo "cleanup"; $undo && $pdo->inTransaction() && $pdo->rollBack(); } }
function elseifArm(PDO $pdo, bool $skip, bool $undo) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { <warning descr="Close the owned transaction before returning.">return;</warning> } $pdo->commit(); } finally { if ($undo) { $state = 'kept'; } elseif ($pdo->inTransaction()) { $pdo->rollBack(); } } }
