<?php
function f(SQLite3Stmt $s) { <error descr="Start SQLite binding positions at one.">$s->bindValue(0, 'ready')</error>; }
