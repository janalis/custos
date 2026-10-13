<?php
$db=new SQLite3(":memory:");$r=$db->query("SELECT 1 AS item");$row=$r->fetchArray(SQLITE3_ASSOC);echo json_encode($row);
