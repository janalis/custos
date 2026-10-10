<?php
$db=new SQLite3(":memory:");$r=$db->query("SELECT 1 AS item");$row=<warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray()</warning>;echo json_encode($row);
