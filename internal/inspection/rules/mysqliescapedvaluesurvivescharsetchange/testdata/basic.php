<?php
$db = new mysqli(); $db->set_charset('utf8mb4'); $v = $db->real_escape_string($input); if ($db->set_charset('gbk')) { <warning descr="Escape values after selecting the connection charset.">$db->query("SELECT '$v'")</warning>; }
