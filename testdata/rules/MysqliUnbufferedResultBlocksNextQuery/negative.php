<?php
$db=new mysqli();$r=$db->query("SELECT id FROM users",MYSQLI_STORE_RESULT);$db->query("SELECT 1");$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);$r->free();$db->query("SELECT 1");$db->close();
