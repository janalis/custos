<?php
$pdo=new PDO("sqlite::memory:");$pdo->beginTransaction();$pdo->commit();$pdo->beginTransaction();$pdo->rollBack();$pdo->beginTransaction();$pdo->query("SELECT 1");
function direct($p){$p->beginTransaction();}
