<?php
function bad(PDO $p){$p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT * FROM t WHERE x = NULL"</warning>);$p->exec(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"UPDATE t SET x=1 WHERE NULL = y"</warning>);}
