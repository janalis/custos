<?php
function good(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { $pdo->rollBack(); return false; } }
