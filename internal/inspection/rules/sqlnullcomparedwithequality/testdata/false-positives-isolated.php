<?php
function valid(PDO $p){$p->query("SELECT '= NULL', `NULL` FROM t WHERE x IS NULL");$p->query("SELECT * FROM t WHERE x >= NULL");$p->query("SELECT * FROM t /* x = NULL */");$p->query($sql);$p->beginTransaction();$p->query("SELECT @NULL = x");}
function unknown($p){$p->query("SELECT x = NULL");}

function assignNull(PDO $p){$p->exec("UPDATE t SET deleted_at=NULL WHERE id=1");$p->exec("INSERT INTO t VALUES (NULL)");}
