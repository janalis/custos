<?php
$p=new PDO("mysql:host=localhost");$p->beginTransaction();<warning descr="Keep implicitly committing DDL outside rollback-dependent transactions.">$p->exec("CREATE TABLE t (id INT)")</warning>;$p->rollBack();
