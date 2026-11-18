<?php
function good(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } } finally { $pdo->rollBack(); } }
function commitInReturn(PDO $pdo) { if (!$pdo->beginTransaction()) { return false; } return $pdo->commit(); }
function eitherBranch(PDO $pdo, $ok) { if (!$pdo->beginTransaction()) { return false; } return $ok ? $pdo->commit() : $pdo->rollBack(); }
function guardedCleanup(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } $pdo->commit(); } finally { if ($pdo->inTransaction()) { $pdo->rollBack(); } } }
