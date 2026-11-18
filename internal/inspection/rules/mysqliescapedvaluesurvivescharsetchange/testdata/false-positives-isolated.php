<?php
$db = new mysqli(); $db->set_charset('gbk'); $v = $db->real_escape_string($input); $db->query("SELECT '$v'");
