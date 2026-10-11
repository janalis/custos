<?php
$db = new SQLite3(':memory:');
$r = $db->query('SELECT 1 AS item');
$row = $r->fetchArray();
echo json_encode($row);
echo $GLOBALS['row'][0];
