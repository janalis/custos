<?php
// PDO::quote encloses and escapes a value for SQL; plain text does not.
function pdo(PDO $db) { return $db->query("SELECT * FROM t WHERE a = " . $db->quote($_GET["a"])); }
function pdoRaw(PDO $db) { return <error descr="Bind untrusted values as SQL parameters.">$db->query("SELECT * FROM t WHERE a = " . $_GET["a"])</error>; }
