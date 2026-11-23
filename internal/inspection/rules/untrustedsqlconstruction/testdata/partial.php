<?php
function q(PDO $db) {
    // A partial application passes the arguments it binds on every call.
    $rows = <error descr="Bind untrusted values as SQL parameters.">$db->query("SELECT * FROM t WHERE id = " . $_GET['id'], ...)</error>;
    // The query comes with each call.
    $later = $db->query(?);
}
