<?php
$db=new mysqli();$r=$db->query("SELECT id FROM users",MYSQLI_STORE_RESULT);$db->query("SELECT 1");$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);$r->free();$db->query("SELECT 1");$db->close();
function consumedInSameStatement(mysqli $db){$r=$db->query("SELECT id FROM users",MYSQLI_USE_RESULT);if($r->fetch_all()&&($x=$db->query("SELECT 1"))){return $x;}return null;}
