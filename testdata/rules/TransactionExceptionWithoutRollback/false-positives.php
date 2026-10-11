<?php
function returnsRollback(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return $pdo->rollBack(); } }
function rollsBackInEveryArm(PDO $pdo, $keep) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return $keep ? $pdo->commit() : $pdo->rollBack(); } }
function guardedFinally(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return false; } finally { if ($pdo->inTransaction()) { $pdo->rollBack(); } } }
function plainFinally(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { return false; } finally { $pdo->rollBack(); } }
