<?php
$p=new PDO("sqlite::memory:");$p->beginTransaction();$p->exec("CREATE TABLE t (id INT)");$p->rollBack();$m=new PDO("mysql:host=localhost");$m->beginTransaction();$m->exec("CREATE TEMPORARY TABLE t (id INT)");$m->rollBack();$m->query("SELECT 1");
