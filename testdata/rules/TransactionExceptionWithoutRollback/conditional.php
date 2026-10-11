<?php
function oneArm(PDO $pdo, $keep) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { <warning descr="Roll back the active transaction on the caught failure path.">return $keep ? $pdo->rollBack() : false;</warning> } }
function otherGuard(PDO $pdo, $debug) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { <warning descr="Roll back the active transaction on the caught failure path.">return false;</warning> } finally { if ($debug) { $pdo->rollBack(); } } }
