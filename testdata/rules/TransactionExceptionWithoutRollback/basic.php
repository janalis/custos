<?php
function bad(PDO $pdo) { if (!$pdo->beginTransaction()) { return; } try { throw new RuntimeException(); } catch (Throwable $e) { <warning descr="Roll back the active transaction on the caught failure path.">return false;</warning> } }
