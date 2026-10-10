<?php
function good(PDO $pdo, bool $skip) { if (!$pdo->beginTransaction()) { return; } try { if ($skip) { return; } } finally { $pdo->rollBack(); } }
