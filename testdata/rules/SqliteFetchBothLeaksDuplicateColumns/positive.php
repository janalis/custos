<?php
function f(SQLite3Result $r) { $row=<warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray(SQLITE3_BOTH)</warning>; echo json_encode($row); }
