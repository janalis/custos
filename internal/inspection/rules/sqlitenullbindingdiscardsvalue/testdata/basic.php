<?php
function f(SQLite3Stmt $s) { <warning descr="Pass null when using the SQLite null binding type.">$s->bindValue(1, 'ready', SQLITE3_NULL)</warning>; }
