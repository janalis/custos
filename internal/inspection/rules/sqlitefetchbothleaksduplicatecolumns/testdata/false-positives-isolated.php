<?php
function f(SQLite3Result $r) { $row=$r->fetchArray(SQLITE3_ASSOC); echo json_encode($row); }
