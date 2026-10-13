<?php
function f(SQLite3Result $r) { if($r->finalize()){ <error descr="Read the SQLite result before finalizing it.">$r->fetchArray()</error>; } }
